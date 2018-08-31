package tool

import (
	"fmt"
	"log"
)

//error log
func ErrorLog(err error, msg string) error {
	if err != nil {
		fmt.Printf("[x] %s: %s", err, msg)
	}
	return err
}

//error log
func ErrorLogP(msg string) {
	fmt.Println("[x]", msg)
}

//error log
func ErrorLogf(err error, format string, v ...interface{}) error {
	if err != nil {
		fmt.Printf("[x] %s "+format, err, v)
	}
	return err
}

//fatal log
func ErrorPanic(err error, msg string) {
	if err != nil {
		log.Panicf("[x] %s: %s", err, msg)
	}
}

//pretty log
func PrettyPrint(v ...interface{}) {
	fmt.Printf("[*] %+v", v)
}

func PrettyPrintf(format string, v ...interface{}) {
	fmt.Printf("[*] "+format, v)
}
