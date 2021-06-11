package main

import (
	"fmt"
)

// default constants
const (
	// show or hide debugs in terminal
	// true:debug | false:deployment
	constDebug = true
)

// Debug based on fmt.Println.
func Debug(s ...interface{}) {
	if constDebug {
		fmt.Println(s...)
	}
}
