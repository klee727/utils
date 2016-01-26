package main

import (
	"time"

	"github.com/yangzhao28/utils/timer/intime"
)

func main() {
	it := intime.NewInTime()
	it.LogPrinter()
	go it.RunCmdServer(":11281")

	it.Append("test1", "echo", "hello", 5*time.Second, false, true)
	it.Append("test2", "echo", "no!!", 10*time.Second, false, true)
	it.Ps()

	it.Wait()
}
