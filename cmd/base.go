package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Die(stopCh chan struct{}) {
	time.Sleep(time.Second * 100)
	fmt.Println("...")
	close(stopCh)
}

func Wait(f func(), stopCh chan struct{}) {
	fmt.Println("waiting...")
	exit := make(chan os.Signal, 1) // Buffer size set to 1 to ensure signal is caught
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case <-exit:
			close(stopCh)
			f() // Call the cleanup function
			return
		}
	}
}
