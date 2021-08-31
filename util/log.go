package util

import (
	"fmt"
	"log"
	"os"
	"path"
)

var (
	infoLogger  *log.Logger = nil
	errorLogger *log.Logger = nil
)

var (
	Info  func(v ...interface{})
	Error func(v ...interface{})
)

func init() {
	s, err := os.Stat(GetConfig().LogPath)
	if err != nil {
		if !os.IsNotExist(err) {
			_, _ = fmt.Fprintln(os.Stderr, "Get log path stat failed. Error:", err)
			os.Exit(-1)
		} else {
			if err := os.MkdirAll(GetConfig().LogPath, os.ModePerm); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Create log path failed. Error:", err)
				os.Exit(-1)
			}
		}
	} else {
		if !s.IsDir() {
			if err := os.MkdirAll(GetConfig().LogPath, os.ModePerm); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Create log path failed. Error:", err)
				os.Exit(-1)
			}
		}
	}

	logName := path.Join(GetConfig().LogPath, "reset-authentication.log")
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Open log file failed. Error:", err)
		os.Exit(-1)
	}

	infoLogger = log.New(logFile, "[Info] ", log.Ldate|log.Ltime)
	errorLogger = log.New(logFile, "[Error] ", log.Ldate|log.Ltime)

	Info = infoLogger.Println
	Error = errorLogger.Println
}
