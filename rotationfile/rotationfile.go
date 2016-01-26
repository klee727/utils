package rotationfile

import (
	// "fmt"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Rotator struct {
	baseFileName     string
	currentFileName  string
	internalFile     *os.File
	rotationByTime   int
	nextRotationTime int64
	fileLock         sync.Mutex
}

func (this *Rotator) GetCurrentFileName() string {
	return this.currentFileName
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

func (this *Rotator) createSymLink(currentName string) error {
	fmt.Println("enter symlink")
	linkName := this.baseFileName
	if info, err := os.Lstat(linkName); !os.IsNotExist(err) {
		if info != nil && (info.Mode()&os.ModeSymlink == os.ModeSymlink) {
			// remove old link file
			os.Remove(linkName)
		} else {
			// link exist but not symlink, use some alter name as linkname
			linkName += ".alt"
			return nil
		}
	}
	if info, err := os.Lstat(currentName); err != nil {
		if !info.Mode().IsRegular() {
			os.Symlink(currentName, linkName)
		}
	}
	return nil
}

func (this *Rotator) switchFile(now time.Time) error {
	if this.rotationByTime != NoRotation {
		logFileName := this.baseFileName
		logFileName += "." + now.Format(GetTimeFormat(this.rotationByTime))
		// fmt.Println("next log-name will be", logFileName, ".")
		switch this.rotationByTime {
		default:
			break
		case MinutelyRotation:
			this.nextRotationTime = now.Add(time.Minute).Add(-time.Duration(now.Second()) * time.Second).Unix()
		case HourlyRotation:
			this.nextRotationTime = now.Add(time.Hour).Add(-time.Duration(now.Second()+now.Minute()*60) * time.Second).Unix()
		case DailyRotation:
			this.nextRotationTime = now.Add(24 * time.Hour).Add(-time.Duration(now.Hour()*3600+now.Minute()*60+now.Second()) * time.Second).Unix()
		}
		// fmt.Println("next rotation time-point will be", this.nextRotationTime, " vs now ", now.Unix(), ".")
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err == nil {
			this.currentFileName = logFileName
			this.internalFile = logFile
			this.createSymLink(logFileName)
			// fmt.Println("file swapped.")
		}
		return err
	}
	return nil
}

func (this *Rotator) Create(name string, rotationByTime int) {
	if name[len(name)-1] == '\\' || name[len(name)-1] == '/' {
		this.baseFileName = name + "default.log"
	} else {
		this.baseFileName = name
	}
	if strings.LastIndexAny(name, "\\/") != -1 {
		os.MkdirAll(name[0:strings.LastIndexAny(name, "\\/")], 0766)
	}
	// fmt.Println("name:", name)
	this.rotationByTime = rotationByTime
	now := time.Now()
	this.switchFile(now)
}

func (this *Rotator) Write(p []byte) (n int, err error) {
	now := time.Now()
	if now.Unix() >= this.nextRotationTime {
		this.fileLock.Lock()
		defer this.fileLock.Unlock()
		if now.Unix() >= this.nextRotationTime {
			if err := this.switchFile(now); err != nil {
				return 0, err
			}
		}
	}
	return this.internalFile.Write(p)
}

func (this *Rotator) Close() {
	this.internalFile.Close()
}
