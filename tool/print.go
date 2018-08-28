package tool

import "log"

func ErrorLog(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", err, msg)
	}
}

func ErrorPanic(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", err, msg)
	}
}

func PrettyPrint(msg string, data interface{}) {
	log.Printf("[*] %s: %v", msg, data)
}
