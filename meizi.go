/*===========================================
*   Copyright (C) 2013 All rights reserved.
*
*   company      : xiaomi
*   author       : zhangye
*   email        : zhangye@xiaomi.com
*   date         : 2013-12-18 15:22:25
*   description  : main for the crawler
*
=============================================*/
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	gBaseLog    = "./log/"
	gTimeFormat = "2006-01-02 15:04:05"
	gErrorChan  = make(chan error, 128)
	gDataChan   = make(chan string, 128)
	gQuitChan   = make(chan bool, 10)
	gNumCpu     = runtime.NumCPU()
)

var (
	gBaseUrl    = "http://jandan.net"
	gPartUrl    = "ooxx"
	gStartPage  = flag.Int("s", 0, "start of the page to fetch, the default is 0")
	gEndPage    = flag.Int("e", 1, "end of the page to fetch, the defult is 1")
	gSavePath   = flag.String("f", "", "the path to save your images")
	gCpuProfile = flag.String("cpu", "", "write cpu profile to file")
)

var (
	gRegN    = regexp.MustCompile(`\s`)
	gRegList = regexp.MustCompile(`<ol class="commentlist".*</ol>`)
	gRegImg  = regexp.MustCompile(`<p><img src="(.+?)"`)
	gSave    = regexp.MustCompile(`.+?/`)
)

func main() {
	runtime.GOMAXPROCS(gNumCpu)
	go watchingErrors()
	go watchingLog()
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
	if *gCpuProfile != "" {
		f, err := os.Create(*gCpuProfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	watchPanic()

	beforeFetchTime := time.Now()
	getLotImg()
	afterFetchTime := time.Now()
	diff := afterFetchTime.Sub(beforeFetchTime)

	close(gErrorChan)
	<-gQuitChan

	log := fmt.Sprintf("抓取%d页用时%s", *gEndPage-*gStartPage, diff)
	fmt.Println(log)
	gDataChan <- log

	close(gDataChan)
	<-gQuitChan

	return
}
