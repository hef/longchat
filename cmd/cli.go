package cmd

import (
	"bufio"
	"context"
	"github.com/spf13/cobra"
	"github.com/hef/longchat/p2p"
	"log"
	"os"
	gologging "github.com/whyrusleeping/go-logging"
	golog "github.com/ipfs/go-log"


)

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Command Line Interface",
	Run:   cliCommand,
}

func init() {
	rootCmd.AddCommand(cliCmd)
}

func cliCommand(cmd *cobra.Command, args []string) {
	if verbose {
		golog.SetAllLoggers(gologging.INFO)
		golog.SetLogLevel("swarm2", "WARNING")
	}

	ctx := context.Background()
	s, err := p2p.NewServices()
	if err != nil {
		log.Printf("error creating services: %s", err)
		return
	}
	err = s.Init(ctx)
	if err != nil {
		log.Printf("error initialziing services: %v", err)
		return
	}

	go s.Run(ctx)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		text = text[:len(text)-1]
		if err != nil {
			log.Printf("error reading string: %v", err)
		}
		switch text {
		case "/self":
			s.ShowSelf()
		case "/peers":
			s.ShowPeers()
		case "/topics":
			s.ShowTopics()
		case "/foo":
			s.ShowFoo()
		default:
			s.Say(text)
		}
	}

}
