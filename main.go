package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"code.dny.dev/gopherxlr/ipc"
	"code.dny.dev/gopherxlr/websocket"
	"github.com/expr-lang/expr"
)

var (
	version = "unknown"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	scripts := flag.String("scripts-dir", "", "path to load .expr scripts from")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Parameters:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s, Commit: %s, Date: %s\n", version, commit, date)
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
	}
	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stdout, "Version: %s, Commit: %s, Date: %s\n", version, commit, date)
		os.Exit(0)
	}

	if *scripts == "" {
		fmt.Println("please provide a directory to load scripts from")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	programs, err := LoadPrograms(*scripts)
	if err != nil {
		fmt.Println("failed to load programs: ", err)
		os.Exit(1)
	}

	addr := ipc.MustGetAddress(ctx)
	changes := make(chan websocket.StatusChange, 10)
	go func() {
		websocket.Listen(ctx, addr, changes)
	}()

Main:
	for {
		select {
		case <-ctx.Done():
			fmt.Print("\nreceived Ctrl-c, terminating...\n")
			break Main
		case res, ok := <-changes:
			if !ok {
				continue
			}
			env := Env{
				StatusChange: res,
				Context:      ctx,
			}
			for _, prog := range programs {
				output, err := expr.Run(prog.prog, env)
				if err != nil {
					fmt.Printf("error executing %s: %s\n", prog.file, err)
				}
				if output != nil {
					fmt.Printf("error from %s: %s\n", prog.file, output)
				}
			}
		}
	}
}
