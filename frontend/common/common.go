// Package common contains all the default values needed to run the code files.
// It also simplifies some of the codes for re-use.
package common

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
)

// default constants
const (
	// show or hide debugs in terminal
	// true:debug | false:deployment
	ConstDebug = true

	// whether to save server logs to file
	// opposite of ConstDebug
	ConstDebugSaveLogs = !ConstDebug

	// length of fixed size
	ConstMaxLengthName = 50
)

// Debug based on fmt.Println.
func Debug(s ...interface{}) {
	if ConstDebug {
		fmt.Println(s...)
	}
}

// Debugf based on fmt.Printf.
func Debugf(format string, s ...interface{}) {
	if ConstDebug {
		fmt.Printf(format+"\n", s...)
	}
}

// ToInt convert string to int based on strconv.Atoi.
func ToInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	return n, err
}

// ToStr convert int to string based on strconv.Itoa.
func ToStr(n int) string {
	return strconv.Itoa(n)
}

// GenerateUUID generates random string of length 36.
func GenerateUUID() string {
	id := uuid.NewV4()
	return id.String()
}

// GetEnv gets value by key from .env file
func GetEnv(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//Debug(".env file loaded")
	return os.Getenv(key)
}
