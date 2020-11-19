package app

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cmdutil "github.com/maoqide/kubeutil/cmd"
	"github.com/maoqide/kubeutil/controllers/demo"
	"github.com/maoqide/kubeutil/initialize"
	"github.com/maoqide/kubeutil/options"
	"github.com/maoqide/kubeutil/pkg/client"
	"github.com/maoqide/kubeutil/utils"
)

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
			cmdutil.Wait(func() { fmt.Println("exiting.") }, stopCh)
		},
	}
	flags = cmd.Flags()
	flags.BoolVarP(&opt.Version, "version", "v", false, "Print version information and quit")
	// flags.BoolVar(&opt.Version, "version", false, "Print version information and quit")

	return cmd
}

func printVersion() {
	fmt.Printf("kubeutil version: %s\n", initialize.Version)
	os.Exit(0)
}

func printHelp() {
	fmt.Printf("kubeutil help \n")
	os.Exit(0)
}

func run(stopCh <-chan struct{}) {
	kubeConfig, _ := utils.ReadFile("./config")
	kubeC, _ := client.NewKubeOutClusterClient(kubeConfig)
	sharedInformerFactory, _ := client.NewSharedInformerFactory(kubeC)
	demoController := demo.NewDemoController(
		sharedInformerFactory.Core().V1().Pods(),
		sharedInformerFactory.Apps().V1().Deployments(),
		sharedInformerFactory.Apps().V1().StatefulSets(),
	)
	go sharedInformerFactory.Start(stopCh)
	demoController.Run(5, stopCh)
	fmt.Println("exit")
}
