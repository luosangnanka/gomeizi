/*===========================================
*   Copyright (C) 2013 All rights reserved.
*
*   company      : xiaomi
*   author       : zhangye
*   email        : zhangye@xiaomi.com
*   date         : 2013-12-18 15:22:25
*   description  : utils for the crawler
*
=============================================*/
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// 通过循环获得所有图片的url
func getLotImg() {
	for i := *gStartPage; i < *gEndPage; i++ {
		go fetchInfo(i)
	}

	for i := *gStartPage; i < *gEndPage; i++ {
		<-gQuitChan
	}

	return
}

func fetchInfo(pageNum int) {
	url := fmt.Sprintf("%s/%s/page-%d", gBaseUrl, gPartUrl, pageNum)
	downloadInfo, err := http.Get(url)
	if err != nil {
		gErrorChan <- err
		return
	}

	body, err := ioutil.ReadAll(downloadInfo.Body)
	if err != nil {
		gErrorChan <- err
		return
	}

	defer func() {
		downloadInfo.Body.Close()
		gQuitChan <- true
	}()

	bodyStr := gRegN.ReplaceAllString(string(body), " ")

	//查找匹配 <ol class="commentlist"><"/ol">
	liCommentM := gRegList.Find([]byte(bodyStr))

	//查找匹配 <p><img src="xxx" /></p>
	imgUrls := gRegImg.FindAllSubmatch(liCommentM, -1)

	for _, v := range imgUrls {
		go saveImg(string(v[1]))
	}

	return
}

// 保存图片
func saveImg(url string) {
	fmt.Println(url)
	gDataChan <- url

	imageName := gSave.ReplaceAllString(url, "")
	path := *gSavePath + imageName

	err := url2FIle(url, path)
	if err != nil {
		gErrorChan <- err
		return
	}

	return
}

// 将图片url转换成对应路径的文件
func url2FIle(url, path string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	f, err := os.Create(path)
	if err != nil {
		return
	}

	defer func() {
		resp.Body.Close()
		f.Close()
	}()

	f.Write(body)

	return
}
