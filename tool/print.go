package tool

import (
	"log"
)

func ErrorLog(err error, msg string) error {
	if err != nil {
		log.Printf("[x] %s: %s", err, msg)
	}
	return err
}

func ErrorPanic(err error, msg string) {
	if err != nil {
		log.Fatalf("[x] %s: %s", err, msg)
	}
}

func PrettyPrint(msg ...interface{}) {
	log.Println("[*]", msg)
}
