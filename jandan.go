package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	go watchError()
	flag.Parse()
	if *gEndPage == 0 || *gSavePath == "" {
		fmt.Printf("Usage %s\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	path := *gSavePath
	if path[len(path)-2:] != "/" {
		*gSavePath += "/"
	}

	before := time.Now()

	for i := 0; i < 10; i++ {
		go savePic()
	}

	for i := *gStartPage; i < *gEndPage; i++ {
		go getUrls(i)
	}

	for i := *gStartPage; i < *gEndPage; i++ {
		<-gQuitChan
	}

	close(gURLChan)
	for i := 0; i < 10; i++ {
		<-gQuitChan
	}

	after := time.Now()
	diff := after.Sub(before)

	close(gErrorChan)
	<-gQuitChan

	fmt.Printf("fetching %d pages finished, use time: %s\n", *gEndPage-*gStartPage, diff)

	return
}

// getUrls get all images url in the web page.
func getUrls(pageNum int) {

	defer func() {
		gQuitChan <- true
	}()

	url := fmt.Sprintf("%s/%s/page-%d", gBaseURL, gPartURL, pageNum)
	downloadInfo, err := Get(url).Response()
	if err != nil {
		gErrorChan <- err
		return
	}
	defer downloadInfo.Body.Close()

	body, err := ioutil.ReadAll(downloadInfo.Body)
	if err != nil {
		gErrorChan <- err
		return
	}

	bodyStr := gRegN.ReplaceAllString(string(body), " ")

	// match <ol class="commentlist"><"/ol">
	liCommentM := gRegList.Find([]byte(bodyStr))

	// match <p><img src="xxx" /></p>
	imgUrls := gRegImg.FindAllSubmatch(liCommentM, -1)

	for _, v := range imgUrls {
		gURLChan <- string(v[1])
	}

	return
}

// savePic saving pictures in the file.
func savePic() {
	defer func() {
		gQuitChan <- true
	}()

	for v := range gURLChan {
		fmt.Println(v)

		imageName := gSave.ReplaceAllString(v, "")
		if len(imageName) > 20 {
			imageName = RandomString(5) + imageName[15:]
		}
		path := *gSavePath + imageName

		err := url2File(v, path)
		if err != nil {
			gErrorChan <- err
			continue
		}

	}
}

func url2File(url, path string) (err error) {
	resp, err := Get(url).Response()
	if err != nil {
		return
	}
	if resp.Body == nil {
		return fmt.Errorf("body is nil")
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	f, err := os.Create(path)
	if err != nil {
		return
	}

	defer f.Close()

	f.Write(body)

	return
}

// RandomString generate a random string.
func RandomString(l int) string {
	var result bytes.Buffer
	var temp string
	for i := 0; i < l; {
		if string(RandInt(65, 90)) != temp {
			temp = string(RandInt(65, 90))
			result.WriteString(temp)
			i++
		}
	}
	return result.String()
}

// RandInt generate a random int.
func RandInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}
