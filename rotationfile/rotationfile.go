package rotationfile

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Rotator struct {
	baseFileName     string
	currentFileName  string
	internalFile     *os.File
	rotationByTime   int
	nextRotationTime int64
}

func (rotator *Rotator) GetCurrentFileName() string {
	return rotator.currentFileName
}

const (
	NoRotation       = iota
	MinutelyRotation /* just for test */
	HourlyRotation
	DailyRotation
)

func GetTimeFormat(rotationByTime int) string {
	switch rotationByTime {
	default:
		fallthrough
	case NoRotation:
		return ""
	case MinutelyRotation:
		return "20060102-1504"
	case HourlyRotation:
		return "20060102-15"
	case DailyRotation:
		return "20060102"
	}
}

func (rotator *Rotator) Create(name string, rotationByTime int) {
	if name[len(name)-1] == '\\' || name[len(name)-1] == '/' {
		rotator.baseFileName = name + "default.log"
	} else {
		rotator.baseFileName = name
	}
	if strings.LastIndexAny(name, "\\/") != -1 {
		os.MkdirAll(name[0:strings.LastIndexAny(name, "\\/")], 0766)
	}
	fmt.Println("name:", name)
	rotator.rotationByTime = rotationByTime
	now := time.Now()
	rotator.switchFile(now)
}

func (rotator *Rotator) switchFile(now time.Time) error {
	if rotator.rotationByTime != NoRotation {
		logFileName := rotator.baseFileName
		logFileName += "." + now.Format(GetTimeFormat(rotator.rotationByTime))
		fmt.Println("next log-name will be", logFileName, ".")
		switch rotator.rotationByTime {
		default:
			break
		case MinutelyRotation:
			rotator.nextRotationTime = now.Add(time.Minute).Add(-time.Duration(now.Second()) * time.Second).Unix()
		case HourlyRotation:
			rotator.nextRotationTime = now.Add(time.Hour).Add(-time.Duration(now.Second()+now.Minute()*60) * time.Second).Unix()
		case DailyRotation:
			rotator.nextRotationTime = now.Add(24 * time.Hour).Add(-time.Duration(now.Hour()*3600+now.Minute()*60+now.Second()) * time.Second).Unix()
		}
		fmt.Println("next rotation time-point will be", rotator.nextRotationTime, " vs now ", now.Unix(), ".")
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err == nil {
			rotator.currentFileName = logFileName
			rotator.internalFile = logFile
			fmt.Println("file swapped.")
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	return nil
}

func (rotator *Rotator) Write(p []byte) (n int, err error) {
	now := time.Now()
	if now.Unix() >= rotator.nextRotationTime {
		if err := rotator.switchFile(now); err != nil {
			return 0, err
		}
	}
	return rotator.internalFile.Write(p)
}

func (rotator *Rotator) Close() {
	rotator.internalFile.Close()
}
