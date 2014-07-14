package main

import (
	"flag"
	"regexp"
)

// some global configs.
var (
	gBaseLog    = "./log/"
	gTimeFormat = "2006-01-02 15:04:05"
	gErrorChan  = make(chan error, 128)
	gUrlChan    = make(chan string, 128)
	gQuitChan   = make(chan bool, 10)
)

var (
	gBaseUrl   = "http://jandan.net"
	gPartUrl   = "ooxx"
	gStartPage = flag.Int("s", 0, "start of the page to fetch, default 0")
	gEndPage   = flag.Int("e", 1, "end of the page to fetch, defult 1")
	gSavePath  = flag.String("f", "", "the path to save your images")
)

var (
	gRegN    = regexp.MustCompile(`\s`)
	gRegList = regexp.MustCompile(`<ol class="commentlist".*</ol>`)
	gRegImg  = regexp.MustCompile(`<p><img src="(.+?)"`)
	gSave    = regexp.MustCompile(`.+?/`)
)
