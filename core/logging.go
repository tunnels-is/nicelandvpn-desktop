package core

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// We only want to see wails.io logs while in development mode
func (l *LoggerInterface) Print(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func (l *LoggerInterface) Trace(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func (l *LoggerInterface) Debug(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func (l *LoggerInterface) Info(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func (l *LoggerInterface) Warning(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func (l *LoggerInterface) Error(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func (l *LoggerInterface) Fatal(message string) {
	if !PRODUCTION {
		log.Println(message)
	}
}

func InitLogfile() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r, string(debug.Stack()))
		}
	}()

	GLOBAL_STATE.LogPath = GLOBAL_STATE.BasePath
	GLOBAL_STATE.LogFileName = GLOBAL_STATE.LogPath + time.Now().Format("2006-01-02-15-04-05") + ".log"

	var err error
	LogFile, err = os.Create(GLOBAL_STATE.LogFileName)
	if err != nil {
		CreateErrorLog("", "Unable to create log file: ", err)
		return
	}

	err = os.Chmod(GLOBAL_STATE.LogFileName, 0o777)
	if err != nil {
		CreateErrorLog("", "Unable to change ownership of log file: ", err)
		return
	}

	CreateLog("", "New log file created: ", LogFile.Name())
	GLOBAL_STATE.LogFileInitialized = true
}

func GET_FUNC(skip int) string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(skip, pc)
	f := runtime.FuncForPC(pc[0])
	name := f.Name()
	sn := strings.Split(name, ".")
	if sn[len(sn)-1] == "func1" {
		return sn[len(sn)-2]
	}
	return sn[len(sn)-1]
}

func CreateErrorLog(Type string, Line ...interface{}) {
	if !C.DebugLogging {
		return
	}

	if Type == "" {
		Type = "general"
	}
	select {
	case LogQueue <- LogItem{
		Type: Type,
		Line: time.Now().Format("01-02 15:04:05") + " || " + GET_FUNC(3) + " || " + "ERROR || " + fmt.Sprint(Line...),
	}:
	default:
		ErrorLog(false, "COULD NOT PLACE LOG IN THE LOG QUEUE")
	}
}

func CreateLog(Type string, Line ...interface{}) {
	if !C.DebugLogging {
		return
	}

	if Type == "" {
		Type = "general"
	}
	select {
	case LogQueue <- LogItem{
		Type: Type,
		Line: time.Now().Format("01-02 15:04:05") + " || " + GET_FUNC(3) + " || " + fmt.Sprint(Line...),
	}:
	default:
		ErrorLog(false, "COULD NOT PLACE LOG IN THE LOG QUEUE")
	}
}

func StartLogQueueProcessor(MONITOR chan int) {
	defer func() {
		MONITOR <- 1
	}()
	defer RecoverAndLogToFile()

	var assigned bool = false
	toFile := false
	CreateLog("general", "Logging module started")

	for {
		logItem := <-LogQueue
		if C.DebugLogging {
			fmt.Println(logItem.Type, " || ", logItem.Line)
		}

		if strings.Contains(logItem.Line, "FILE") {
			toFile = true
		} else {
			toFile = false
		}

		for i := range L.LOGS {
			if L.LOGS[i] == "" {
				L.LOGS[i] = logItem.Line
				assigned = true
				break
			}
		}

		if !assigned {
			// If the general log slice is full
			// we truncate and start from index 0
			assigned = false
			for i := range L.LOGS {
				L.LOGS[i] = ""
			}
			L.LOGS[0] = logItem.Line
		}

		if LogFile != nil {
			if toFile {
				_, err := LogFile.WriteString(logItem.Line + "\n")
				if err != nil {
					ErrorLog(err)
				}
			}
		} else {
			ErrorLog("Log file not initialized")
		}

	}
}

func ErrorLog(err interface{}, msgs ...interface{}) {
	if C.DebugLogging && !PRODUCTION {
		log.Println(TAG_ERROR+" ||", fmt.Sprint(msgs...), " >> system error:", err)
	}
}

func Log(lines ...interface{}) {
	if C.DebugLogging && !PRODUCTION {
		log.Println(TAG_GENERAL+" ||", fmt.Sprint(lines...))
	}
}
