package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jeorjebot/godot/pkg/client"
)

type Version struct{}

func (v *Version) Run() error {
	fmt.Println("Version 0.2.0")
	return nil
}

type Clean struct{}

func (c *Clean) Run() error {
	err := client.Clean()
	if err != nil {
		log.Fatalf("Error cleaning tasks: %v", err)
	}
	return nil
}

type AddCmd struct {
	Command []string `arg:"" required:"" name:"command" help:"Command to add for further excution."`
}

func (a *AddCmd) Run() error {

	// path, err := client.ExecPath()
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting the current path: %v", err)
	}
	command := strings.Join(a.Command, " ")
	err = client.AddTaskRequest("add", command, path)
	if err != nil {
		log.Fatalf("Error adding task: %v", err)
	}
	return nil
}

type ListCmd struct {
	Short bool `help:"Display only command, path, status and log file."`
}

func (l *ListCmd) Run() error {
	err := client.ListRequest("list", l.Short)
	if err != nil {
		log.Fatalf("Error listing tasks: %v", err)
	}
	return nil
}

type RmCmd struct {
	ID int `arg:"" required:"" name:"taskID" help:"Remove the task with that ID, if exists and it's not in execution."`
}

func (r *RmCmd) Run() error {
	err := client.RemoveTaskRequest("remove", r.ID)
	if err != nil {
		log.Fatalf("Error removing task: %v", err)
	}
	return nil
}

type GetCmd struct {
	ID int `arg:"" required:"" name:"taskID" help:"Display the task with the specified ID, if exists."`
}

func (l *GetCmd) Run() error {
	err := client.GetTaskRequest("get", l.ID)
	if err != nil {
		log.Fatalf("Error getting task: %v", err)
	}
	return nil
}

type HistoryCmd struct{}

func (h *HistoryCmd) Run() error {
	err := client.HistoryRequest("history")
	if err != nil {
		log.Fatalf("Error getting history: %v", err)
	}
	return nil
}

type TaskCmd struct {
	Add AddCmd `cmd:"" help:"Add a new task."`
	Rm  RmCmd  `cmd:"" help:"Remove a specific task."`
	Get GetCmd `cmd:"" help:"Retrieve a specific task with all details."`
}
