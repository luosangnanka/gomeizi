/*===========================================
*   Copyright (C) 2013 All rights reserved.
*
*   company      : xiaomi
*   author       : zhangye
*   email        : zhangye@xiaomi.com
*   date         : 2013-12-18 15:22:25
*   description  : logging for the crawler
*
=============================================*/
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// 错误写入log
func watchingErrors() {
	now := time.Now().Format("2006-01-02")
	filePath := gBaseLog + now + "_err.log"

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open error log file filed", err.Error())
		os.Exit(1)
	}
	defer func() {
		file.Close()
		gQuitChan <- true
	}()

	logTime := time.Now().Format(gTimeFormat)

	for v := range gErrorChan {
		file.WriteString(logTime + " " + v.Error() + "\n")
	}

	return
}

// 抓取log写入文件
func watchingLog() {
	now := time.Now().Format("2006-01-02")
	filePath := gBaseLog + now + "_meizi.log"

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open meizi log file filed", err.Error())
		os.Exit(1)
	}
	defer func() {
		file.Close()
		gQuitChan <- true
	}()

	logTime := time.Now().Format(gTimeFormat)

	for v := range gDataChan {
		file.WriteString(logTime + " " + v + "\n")
	}

	return
}

// panic重定向到.panic文件
func watchPanic() {
	var panicFile *os.File
	panicFile, err := os.OpenFile(".panic", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panicFile, err = os.OpenFile("/dev/null", os.O_RDWR, 0)
	}
	if err == nil {
		fd := panicFile.Fd()
		syscall.Dup2(int(fd), int(os.Stderr.Fd()))
	}
}
