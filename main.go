package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/cache"

	"github.com/maoqide/kubeutil/kube"
	"github.com/maoqide/kubeutil/options"
	"github.com/maoqide/kubeutil/utils"
)

var (
	// GitCommit git commit id
	GitCommit = "Unknown"
	// BuildTime build time
	BuildTime = "Unknown"
	// Version v1.0
	Version = "v1.0"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := NewKubeCommand()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// NewKubeCommand creates a *cobra.Command object with default parameters
func NewKubeCommand() *cobra.Command {
	opt, err := options.NewkubeOptions()
	if err != nil {
		log.Fatalf("unable to initialize command options: %v", err)
	}
	var flags *pflag.FlagSet

	cmd := &cobra.Command{
		Use:  "kubeutil",
		Long: `kube-util is utils for kubernetes.`,
		Run: func(cmd *cobra.Command, args []string) {
			if opt.Version {
				printVersion()
			}
			var stopCh = make(chan struct{})
			go run(stopCh)
			wait(func() { fmt.Println("exiting.") }, stopCh)
		},
	}
	flags = cmd.Flags()
	flags.BoolVarP(&opt.Version, "version", "v", false, "Print version information and quit")
	// flags.BoolVar(&opt.Version, "version", false, "Print version information and quit")

	return cmd
}

func printVersion() {
	fmt.Printf("kubeutil version: %s\n", Version)
	os.Exit(0)
}

func printHelp() {
	fmt.Printf("kubeutil help \n")
	os.Exit(0)
}

func run(stopCh <-chan struct{}) {
	kubeConfig, _ := utils.ReadFile("./config")
	kubeC, _ := kube.NewKubeOutClusterClient(kubeConfig)
	sharedInformerFactory, _ := kube.NewSharedInformerFactory(kubeC)
	podInformer := sharedInformerFactory.Core().V1().Pods().Informer()
	podHandler := kube.NewPodHandler()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			podHandler.HandleAdd(kube.EVENTADD, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			podHandler.HandleUpdate(kube.EVENTUPDATE, oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			podHandler.HandleDel(kube.EVENTDELETE, obj)
		},
	})
	podInformer.Run(stopCh)
	fmt.Println("exit")
}

func die(stopCh chan struct{}) {
	time.Sleep(time.Second * 100)
	fmt.Println("...")
	close(stopCh)
}

func wait(f func(), stopCh chan struct{}) {
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
