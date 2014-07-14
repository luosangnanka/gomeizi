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

// Get 方法返回实例化 GET 请求的对象
func Get(url string) (mihttp *MigoHttpRequest) {
	var req http.Request
	req.Method = "GET"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHttpRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Post 方法返回实例化 POST 请求的对象
func Post(url string) (mihttp *MigoHttpRequest) {
	var req http.Request
	req.Method = "POST"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHttpRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Put 方法返回实例化 PUT 请求的对象
func Put(url string) (mihttp *MigoHttpRequest) {
	var req http.Request
	req.Method = "PUT"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHttpRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Delete 方法返回实例化 DELETE 请求的对象
func Delete(url string) (mihttp *MigoHttpRequest) {
	var req http.Request
	req.Method = "DELETE"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHttpRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

// Head 方法返回实例化 HEAD 请求的对象
func Head(url string) (mihttp *MigoHttpRequest) {
	var req http.Request
	req.Method = "HEAD"
	req.Header = http.Header{}
	req.Header.Set("User-Agent", defaultUserAgent)

	return &MigoHttpRequest{url, &req, map[string]string{}, false, 60 * time.Second, 60 * time.Second, nil}
}

type MigoHttpRequest struct {
	url             string
	req             *http.Request
	params          map[string]string
	debug           bool
	connectTimeout  time.Duration
	rwTimeout       time.Duration
	tlsClientConfig *tls.Config
}

// Debug 设置在执行请求时是否开启 debug 模式.
func (m *MigoHttpRequest) Debug(isdebug bool) *MigoHttpRequest {
	m.debug = isdebug

	return m
}

// SetTimeout 设置 connect timeout 和 read-write timeout.
func (m *MigoHttpRequest) SetTimeout(connectTimeout, rwTimeout time.Duration) *MigoHttpRequest {
	m.connectTimeout = connectTimeout
	m.rwTimeout = connectTimeout

	return m
}

// SetConnectTimeout 独立设置 connect timeout.
func (m *MigoHttpRequest) SetConnectTimeout(connectTimeout time.Duration) *MigoHttpRequest {
	m.connectTimeout = connectTimeout

	return m
}

// SetRWTimeout 独立设置 read-write timeout.
func (m *MigoHttpRequest) SetRWTimeout(rwTimeout time.Duration) *MigoHttpRequest {
	m.rwTimeout = rwTimeout

	return m
}

// SetTLSClientConfig 设置 https 访问时的 tls connection configurations.
func (m *MigoHttpRequest) SetTLSClientConfig(config *tls.Config) *MigoHttpRequest {
	m.tlsClientConfig = config

	return m
}

// Header 添加请求 header.
func (m *MigoHttpRequest) Header(key, value string) *MigoHttpRequest {
	m.req.Header.Add(key, value)

	return m
}

// SetCookie 添加 cookie 到请求中.
func (m *MigoHttpRequest) SetCookie(cookie *http.Cookie) *MigoHttpRequest {
	m.req.Header.Add("cookie", cookie.String())

	return m
}

// Param 添加请求的参数, 如 ?key1=value1&key2=value2...
func (m *MigoHttpRequest) Param(key, value string) *MigoHttpRequest {
	m.params[key] = value

	return m
}

// Body 添加 request raw body.
// 支持的类型为 string 和 []byte 两种.
func (m *MigoHttpRequest) Body(data interface{}) *MigoHttpRequest {
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

// 请求主体.
func (m *MigoHttpRequest) getResponse() (resp *http.Response, err error) {
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

// 请求成 string
func (m *MigoHttpRequest) String() (data string, err error) {
	dataByte, err := m.Bytes()
	if err != nil {
		return
	}

	return string(dataByte), nil
}

// 请求成 body
func (m *MigoHttpRequest) Bytes() (data []byte, err error) {
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

// 将请求的 data 转存到文件中.
func (m *MigoHttpRequest) ToFile(filename string) (err error) {
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

// 将请求的数据转成 json 格式
func (m *MigoHttpRequest) ToJson(v interface{}) (err error) {
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

// 将请求的数据转换成 XML 格式.
func (m *MigoHttpRequest) ToXML(v interface{}) (err error) {
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

// 执行请求主体
func (m *MigoHttpRequest) Response() (resp *http.Response, err error) {
	return m.getResponse()
}

// TimeoutDialer 处理响应超时时间.
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
