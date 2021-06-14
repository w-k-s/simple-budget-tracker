package util

import "log"

func CheckError(err error, message string) {
	if err != nil {
		log.Fatalf("%s. Reason: '%s'", message, err)
	}
}
