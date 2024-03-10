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

func main() {
	scripts := flag.String("scripts-dir", "", "path to load .expr scripts from")
	flag.Parse()

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
				_, err := expr.Run(prog, env)
				if err != nil {
					fmt.Println("error: ", err)
				}
			}
		}
	}
}
