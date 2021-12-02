package model

import (
	"log"
	"os"
)

func init() {
	logPath := "../log/logfile.log"
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("file open error : %v", err)
	}
	log.SetOutput(f)
	log.Println("---------- Service started ---------")
}
