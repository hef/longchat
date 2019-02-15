package cmd

import (
	"context"
	"github.com/chzyer/readline"
	"github.com/hef/longchat/p2p"
	"github.com/spf13/cobra"
	"io"
	"log"
)

var readlineCmd = &cobra.Command{
	Use:   "readline",
	Short: "Readline Interface",
	Run:   readlineCommand,
}

func init() {
	rootCmd.AddCommand(readlineCmd)
}

func readlineCommand(cmd *cobra.Command, args []string) {

     rl, err := readline.NewEx(&readline.Config{
     	Prompt: "\033[31m»\033[0m ",
     	InterruptPrompt: "^C",
     	EOFPrompt: "exit",
     	UniqueEditLine: true,
	 })

	s, err := p2p.NewServices(rl)
	ctx, cancel := context.WithCancel(context.Background())
	s.Init(ctx)
	go s.Run(ctx)
	if err != nil {
		panic(err)
	}

     if err != nil {
     	panic(err)
	 }
     defer rl.Close()
     for {
     	line, err := rl.Readline()
     	if err == readline.ErrInterrupt {
     		if len(line) == 0 {
     			log.Print("interupt")
     			cancel()
     			break
			} else {
				continue
			}
		} else if err == io.EOF {
			log.Printf("eof")
			break
		}
		 switch {
		 default:
		 	s.Say(line)
		 }
	 }
}