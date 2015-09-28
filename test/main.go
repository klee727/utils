package main

import (
	"fmt"
	"os"
	"time"

	"github.com/op/go-logging"
	"github.com/yangzhao28/rotationfile"
)

var log = logging.MustGetLogger("example")

var format = logging.MustStringFormatter(
	// "[%{color}%{time:15:04:05.000} %{shortfunc} %{level:.4s} %{id:03x}%{color:reset}] %{message}",
	"[%{color}%{level:s}%{color:reset}][%{time:2006-01-02 15:04:05.000} %{shortfile}: %{longfunc}][%{id}] %{message}",
)

func main() {
	fmt.Println("service, online.")
	file := &rotationfile.Rotator{}
	file.Create("log/baselog.log", rotationfile.MinutelyRotation)
	fmt.Println("file, created:", file.GetCurrentFileName(), ".")

	fileBackend := logging.NewLogBackend(file, "", 0)
	consoleBackend := logging.NewLogBackend(os.Stderr, "", 0)

	fileBackendFormatter := logging.NewBackendFormatter(fileBackend, format)
	consoleBackendFormatter := logging.NewBackendFormatter(consoleBackend, format)

	fileBackendLeveled := logging.AddModuleLevel(fileBackendFormatter)
	fileBackendLeveled.SetLevel(logging.ERROR, "")

	consoleBackendLeveled := logging.AddModuleLevel(consoleBackendFormatter)
	consoleBackendLeveled.SetLevel(logging.ERROR, "")

	logging.SetBackend(fileBackendFormatter, consoleBackendFormatter)

	for true {
		log.Debug("debug")
		log.Info("info")
		log.Notice("notice")
		log.Warning("warning")
		log.Error("err")
		log.Critical("crit")

		time.Sleep(time.Second)
	}
}
