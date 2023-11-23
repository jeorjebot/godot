package task

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Task struct {
	Command     string `json:"command"`
	Path        string `json:"path"`
	EnqueueTime int64  `json:"enqueue_time"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	ExitCode    string `json:"exit_code"`
	Status      string `json:"status"`
	Error       string `json:"error"`
	LogFile     string `json:"log_file"` // the path of the file containing the log of the task
}

func (t *Task) Run() {

	logFile, logFileName := NewLogFile() // save output on a file

	// set the start time, end time, exit code, and status of the task
	t.StartTime = time.Now().UnixMilli()
	t.EndTime = 0
	t.Status = "running"
	t.LogFile = logFileName
	t.Error = "-"

	splittedCommand := strings.Split(t.Command, " ")

	// execute the command of the task
	cmd := exec.Command(splittedCommand[0], splittedCommand[1:]...)
	cmd.Dir = t.Path
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err := cmd.Run()

	// set the end time, exit code, and status of the task
	t.EndTime = time.Now().UnixMilli()
	if err != nil {
		t.Status = "failed"
		t.Error = err.Error()
		t.ExitCode = "1" // FIXME get the exit code from the error
	} else {
		t.ExitCode = "0"
		t.Status = "success"
	}

	// log the task on the godot_history file
	LogTask(*t)

}

func (t Task) HistoryString() string {
	return fmt.Sprintf("%s\t%s\t%s\n", t.Command, t.Path, t.LogFile)
}

func (t Task) ShortString() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\n", t.Command, t.Path, t.Status, t.LogFile)
}

func (t Task) LongString() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", t.Command, t.Path, FmtTime(t.EnqueueTime, t.StartTime), FmtTime(t.StartTime, t.EndTime), t.ExitCode, t.Status, t.Error, t.LogFile)
}

func FmtTime(start, end int64) string {
	if start == 0 || end == 0 {
		return "-"
	}
	totalTime := (end - start) / 1000
	hours := totalTime / 3600
	minutes := (totalTime % 3600) / 60
	seconds := totalTime % 60
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

func CopyTaskList(taskList map[int]*Task) map[int]Task {
	list := make(map[int]Task)
	for k, v := range taskList {
		list[k] = *v
	}
	return list
}

// TaskManager holds the list of tasks, and manage the execution in order of the tasks.
// It receive the incoming tasks from the Listener, and dispatch them in order to the Worker.
// The sync with the Worker is managed by sharing the ownership of the Task back and forth.
func TaskManager() (chan<- *Task, chan<- int, chan map[int]Task) {

	taskList := make(map[int]*Task)

	// channels for communicating with the listener
	enqueue := make(chan *Task, 100)
	remove := make(chan int)
	list := make(chan map[int]Task)

	// channel for communicating with the worker
	exec := make(chan *Task)
	done := make(chan *Task)
	taskOffset := MaxTaskIDInHistory() + 1
	taskID := 0 + taskOffset

	// start the workers
	go Worker(exec, done)

	go func() {
		for {
			select {

			case t := <-enqueue:
				// add enqueue time
				t.EnqueueTime = time.Now().UnixMilli()

				// if there are no task in execution, enqueue one
				if taskID == len(taskList)+taskOffset {
					exec <- t
				}
				taskList[len(taskList)+taskOffset] = t

			case id := <-remove:
				// if the taskID exists and is not running, mark it as removed
				if _, ok := taskList[id]; ok && taskList[id].Status != "running" {
					taskList[id].Status = "removed"
				}

			case <-list:
				// copy the taskList and send it to the listener
				list <- CopyTaskList(taskList)

			case <-done:
				taskID++
				// while taskID < len(taskList) there are tasks to execute
				for taskID < len(taskList)+taskOffset {
					if taskList[taskID].Status != "removed" {
						exec <- taskList[taskID]
						break
					}
					taskID++
				}
			}
		}
	}()

	return enqueue, remove, list
}

func Worker(exec <-chan *Task, done chan<- *Task) {
	for {
		task := <-exec
		task.Run()
		done <- task
	}
}
