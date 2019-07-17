package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/maoqide/kubeutil/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := cmd.NewKubeCommand()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
