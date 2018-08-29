package tool

import (
	"log"
)

//error log
func ErrorLog(err error, msg string) error {
	if err != nil {
		log.Printf("[x] %s: %s", err, msg)
	}
	return err
}

//fatal log
func ErrorPanic(err error, msg string) {
	if err != nil {
		log.Fatalf("[x] %s: %s", err, msg)
	}
}

//pretty log
func PrettyPrint(msg ...interface{}) {
	log.Println("[*]", msg)
}
