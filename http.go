package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var defaultUserAgent = "migoServer"

// Get new MigoHTTPRequest with GET method.
func Get(url string) (mihttp *MigoHTTPRequest) {
	var req http.Request
	req.Method = "GET"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHTTPRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Post new MigoHTTPRequest with Post method.
func Post(url string) (mihttp *MigoHTTPRequest) {
	var req http.Request
	req.Method = "POST"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHTTPRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Put new MigoHTTPRequest with Put method.
func Put(url string) (mihttp *MigoHTTPRequest) {
	var req http.Request
	req.Method = "PUT"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHTTPRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Delete new MigoHTTPRequest with Delete method.
func Delete(url string) (mihttp *MigoHTTPRequest) {
	var req http.Request
	req.Method = "DELETE"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHTTPRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Head new MigoHTTPRequest with Head method.
func Head(url string) (mihttp *MigoHTTPRequest) {
	var req http.Request
	req.Method = "HEAD"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHTTPRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// MigoHTTPRequest migo http request struct.
type MigoHTTPRequest struct {
	url             string
	req             *http.Request
	params          map[string]string
	debug           bool
	connectTimeout  time.Duration
	rwTimeout       time.Duration
	tlsClientConfig *tls.Config
}

// Debug will echo the debug info when isdebug is set true.
func (m *MigoHTTPRequest) Debug(isdebug bool) *MigoHTTPRequest {
	m.debug = isdebug

	return m
}

// SetTimeout set connect timeout and read-write timeout.
func (m *MigoHTTPRequest) SetTimeout(connectTimeout, rwTimeout time.Duration) *MigoHTTPRequest {
	m.connectTimeout = connectTimeout
	m.rwTimeout = connectTimeout

	return m
}

// SetConnectTimeout set connect timeout.
func (m *MigoHTTPRequest) SetConnectTimeout(connectTimeout time.Duration) *MigoHTTPRequest {
	m.connectTimeout = connectTimeout

	return m
}

// SetRWTimeout set read-write timeout.
func (m *MigoHTTPRequest) SetRWTimeout(rwTimeout time.Duration) *MigoHTTPRequest {
	m.rwTimeout = rwTimeout

	return m
}

// SetTLSClientConfig set https tls connection configurations.
func (m *MigoHTTPRequest) SetTLSClientConfig(config *tls.Config) *MigoHTTPRequest {
	m.tlsClientConfig = config

	return m
}

// Header add request header.
func (m *MigoHTTPRequest) Header(key, value string) *MigoHTTPRequest {
	m.req.Header.Add(key, value)

	return m
}

// SetCookie add cookie into the request.
func (m *MigoHTTPRequest) SetCookie(cookie *http.Cookie) *MigoHTTPRequest {
	m.req.Header.Add("cookie", cookie.String())

	return m
}

// Param add http request params, such as ?key1=value1&key2=value2...
func (m *MigoHTTPRequest) Param(key, value string) *MigoHTTPRequest {
	m.params[key] = value

	return m
}

// Body add request raw body.
// support type string and []byte.
func (m *MigoHTTPRequest) Body(data interface{}) *MigoHTTPRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		m.req.Body = ioutil.NopCloser(bf)
		m.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		m.req.Body = ioutil.NopCloser(bf)
		m.req.ContentLength = int64(len(t))
	}

	return m
}

// getResponse is the main http request method.
func (m *MigoHTTPRequest) getResponse() (resp *http.Response, err error) {
	var paramBody string
	if len(m.params) > 0 {
		var buf bytes.Buffer
		for k, v := range m.params {
			buf.WriteString(url.QueryEscape(k))
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
			buf.WriteByte('&')
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	if m.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Index(m.url, "?") != -1 {
			m.url += "&" + paramBody
		} else {
			m.url = m.url + "?" + paramBody
		}
	} else if m.req.Method == "POST" && m.req.Body == nil && len(paramBody) > 0 {
		m.Header("Content-Type", "application/x-www-form-urlencoded")
		m.Body(paramBody)
	}

	u, err := url.Parse(m.url)
	if u.Scheme == "" {
		m.url = "http://" + m.url
		u, err = url.Parse(m.url)
	}
	if err != nil {
		return
	}

	m.req.URL = u
	if m.debug {
		dump, er := httputil.DumpRequest(m.req, true)
		if er != nil {
			err = er
			return
		}
		fmt.Println(string(dump))
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: m.tlsClientConfig,
			Dial:            TimeoutDialer(m.connectTimeout, m.rwTimeout),
		},
	}
	resp, err = client.Do(m.req)
	if err != nil {
		return
	}

	return
}

// String generate data with string format.
func (m *MigoHTTPRequest) String() (data string, err error) {
	dataByte, err := m.Bytes()
	if err != nil {
		return
	}

	return string(dataByte), nil
}

// Bytes generate data with []byte format.
func (m *MigoHTTPRequest) Bytes() (data []byte, err error) {
	resp, err := m.getResponse()
	if err != nil {
		return
	}
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

// ToFile save data into the file.
func (m *MigoHTTPRequest) ToFile(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()

	resp, err := m.getResponse()
	if err != nil {
		return
	}
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return
	}

	return
}

// ToJSON generate data with json format.
func (m *MigoHTTPRequest) ToJSON(v interface{}) (err error) {
	data, err := m.Bytes()
	if err != nil {
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}

	return
}

// ToXML generate data with XML format.
func (m *MigoHTTPRequest) ToXML(v interface{}) (err error) {
	data, err := m.Bytes()
	if err != nil {
		return
	}
	err = xml.Unmarshal(data, v)
	if err != nil {
		return
	}

	return
}

// Response the main http request method for the outside.
func (m *MigoHTTPRequest) Response() (resp *http.Response, err error) {
	return m.getResponse()
}

// TimeoutDialer dailer the timeout time.
func TimeoutDialer(connectTimeout, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (c net.Conn, err error) {
		conn, err := net.DialTimeout(netw, addr, connectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}
