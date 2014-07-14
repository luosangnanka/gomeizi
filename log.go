package main

import (
	"fmt"
	"os"
	"time"
)

// watchError watching all error in the log.
func watchError() {
	now := time.Now().Format("2006-01-02")
	filepath := gBaseLog + now + "_err.log"

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open error log file filed", err.Error())
		os.Exit(1)
	}
	defer func() {
		f.Close()
		gQuitChan <- true
	}()

	for v := range gErrorChan {
		errTime := time.Now().Format(gTimeFormat)
		f.WriteString(fmt.Sprintf("%s %s\n", errTime, v.Error()))
	}

	return
}
