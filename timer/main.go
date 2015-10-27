package main

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/op/go-logging"
	"github.com/satori/go.uuid"
	"github.com/yangzhao28/rotatelogger"
)

type Timer struct {
	timer    *time.Timer
	duration time.Duration
	quit     chan bool
	action   func()
}

func NewTimer(duration time.Duration) *Timer {
	return &Timer{
		timer:    time.NewTimer(duration),
		duration: duration,
		quit:     make(chan bool),
	}
}

type TimerType int

const (
	Relative TimerType = iota
	Absolute
)

const (
	MaxJobLogQueueSize = 128
	MaxJobQueueSize    = 128
)

type Job struct {
	id       string
	cmd      string
	args     string
	interval time.Duration
	repeat   bool
	debug    bool
	begin    time.Time

	// Events
	OnOneTurnDone chan *Job
	OnAllDone     chan *Job

	logNotifier  chan string
	ctrlNotifier chan int
}

func NewJob(cmd, args string, interval time.Duration, repeat bool, logNotifier chan string, debug bool) *Job {
	return &Job{
		id:       uuid.NewV4().String(),
		cmd:      cmd,
		args:     args,
		interval: interval,
		begin:    time.Now(),
		repeat:   repeat,
		debug:    debug,

		OnOneTurnDone: nil,
		OnAllDone:     nil,

		logNotifier:  logNotifier,
		ctrlNotifier: make(chan int),
	}
}

type Event int

const (
	AllDone Event = iota
	OneTurnDone
)

func (j *Job) Log(format string, args ...interface{}) {
	if len(j.logNotifier) < MaxJobLogQueueSize {
		prefix := fmt.Sprintf("<%v> job(%v): %v %v --> ", time.Now(), j.id, j.cmd, j.args)
		if len(args) == 0 {
			j.logNotifier <- fmt.Sprintf(prefix + format)
		} else {
			j.logNotifier <- fmt.Sprintf(prefix+format, args)
		}
	}
}

func (j *Job) Launch() {
	timer := time.NewTimer(j.interval)
	go func() {
		for {
			select {
			case <-timer.C:
				var begin int64
				if j.debug {
					j.Log("job triggered")
					begin = time.Now().Unix()
				}
				cmd := exec.Command(j.cmd, j.args)
				_, err := cmd.Output()
				if err != nil {
					j.Log(err.Error())
				}
				if j.debug {
					j.Log("job done, time elapsed %v", time.Now().Unix()-begin)
				}
				if j.OnOneTurnDone != nil {
					j.OnOneTurnDone <- j
				}
			case <-j.ctrlNotifier:
				break
			}
			if j.repeat {
				timer.Reset(j.interval)
			} else {
				break
			}
		}
		j.Log("done")
		if j.OnAllDone != nil {
			j.OnAllDone <- j
		}
	}()
}

func (j *Job) Connect(event Event, notifier chan *Job) {
	switch event {
	case AllDone:
		j.OnAllDone = notifier
	case OneTurnDone:
		j.OnOneTurnDone = notifier
	default:
		break
	}
}

type InTime struct {
	timerCollection map[string]*Job

	doneNotifier chan *Job
	waitGroup    sync.WaitGroup
	logger       *logging.Logger

	logClear    chan bool
	logNotifier chan string
}

func NewInTime() *InTime {
	return &InTime{
		timerCollection: make(map[string]*Job),
		logger:          rotatelogger.NewLogger("InTime", "", "DEBUG"),
		logNotifier:     make(chan string, MaxJobLogQueueSize*10),
		doneNotifier:    make(chan *Job, MaxJobQueueSize),
		logClear:        make(chan bool),
	}
}

func (it *InTime) Append(name string, cmd, args string, interval time.Duration, repeat bool, debug bool) {
	job := NewJob(cmd, args, interval, repeat, it.logNotifier, debug)
	it.timerCollection[name] = job
	job.Connect(AllDone, it.doneNotifier)
	job.Launch()
}

func (it *InTime) Wait() {
	done := 0
	for job := range it.doneNotifier {
		it.logger.Debug("job: %v quit", job.id)
		done += 1
		if done == len(it.timerCollection) {
			it.logClear <- true
			break
		}
	}
	it.waitGroup.Wait()
	it.logger.Notice("No more to do, bye")
}

func (it *InTime) LogPrinter() {
	it.waitGroup.Add(1)
	go func() {
		quit := false
		for {
			select {
			case record := <-it.logNotifier:
				it.logger.Notice(record)
			case <-it.logClear:
				quit = true
			}
			if quit && len(it.logNotifier) == 0 {
				it.waitGroup.Done()
				break
			}
		}
	}()
}

func main() {
	it := NewInTime()
	it.LogPrinter()

	it.Append("test", "echo", "hello", 5*time.Second, false, true)
	it.Wait()
}
