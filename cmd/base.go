package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func Die(stopCh chan struct{}) {
	time.Sleep(time.Second * 100)
	fmt.Println("...")
	close(stopCh)
}

func Wait(f func(), stopCh chan struct{}) {
	fmt.Println("waiting...")
	exit := make(chan os.Signal)
	// signal.Notify(exit, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT)
	signal.Notify(exit, os.Kill, os.Interrupt)
	for {
		select {
		case <-exit:
			close(stopCh)
			f()
			return
		}
	}
}
