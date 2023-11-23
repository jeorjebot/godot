package main

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jeorjebot/godot/pkg/cli"
	"github.com/jeorjebot/godot/pkg/client"
	"github.com/jeorjebot/godot/pkg/server"
	"github.com/jeorjebot/godot/pkg/task"
)

var CLI struct {
	Add     cli.AddCmd     `cmd:"" help:"Add a new task."`
	List    cli.ListCmd    `cmd:"" help:"List the tasks of this session."`
	History cli.HistoryCmd `cmd:"" help:"List the tasks of past sessions."`
	Task    cli.TaskCmd    `cmd:"" help:"Task commands."`
	Clean   cli.Clean      `cmd:"" help:"Clean past session files."`
	Version cli.Version    `cmd:"" help:"Print version information and quit."`
}

func main() {

	// check if the storage is present, if not, create it
	task.UseStorage()

	// when the program starts, it checks if the server it's online
	// if it's not online, it starts the server
	// otherwise, it starts the client

	err := client.Ping()
	if err != nil {

		// Server
		fmt.Println("Starting server...")

		enqueue, remove, list := task.TaskManager()
		shutdown := server.NewServer(enqueue, remove, list)

		// CLI
		ctx := kong.Parse(&CLI)
		err = ctx.Run()
		ctx.FatalIfErrorf(err)

		// remain online: print * every 3 seconds
		for {
			fmt.Print("*")
			time.Sleep(3 * time.Second)

			// check if the server is in shutdown mode
			select {
			case <-shutdown:
				fmt.Println("Server is shutting down...")
				time.Sleep(3 * time.Second)
				return
			default:
				continue
			}
		}
	}

	// Client

	// CLI
	ctx := kong.Parse(&CLI)
	err = ctx.Run()
	ctx.FatalIfErrorf(err)

}
