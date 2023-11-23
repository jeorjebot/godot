package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/jeorjebot/godot/pkg/task"
)

type ChannelSet struct {
	Enqueue chan<- *task.Task
	Remove  chan<- int
	List    chan map[int]task.Task
	alive   chan bool
}

func fmtTaskList(taskList map[int]task.Task, short bool) (string, string) {
	var s string

	// get the map keys
	keys := make([]int, 0, len(taskList))
	for k := range taskList {
		keys = append(keys, k)
	}

	// sort the keys
	sort.Ints(keys)

	for _, k := range keys {
		if short {
			// short string
			s += fmt.Sprintf("[%d]\t%s", k, taskList[k].ShortString())
		} else {
			// long string
			s += fmt.Sprintf("[%d]\t%s", k, taskList[k].LongString())

		}
	}

	if short {
		headerStr := "[ID]\tCommand\tPath\tStatus\tError"
		return headerStr, s
	}
	headerStr := "[ID]\tCommand\tPath\tQueue\tExec\tExitCode\tStatus\tError\tLogFile"
	return headerStr, s
}

func (cs *ChannelSet) addTask(w http.ResponseWriter, r *http.Request) {

	cs.alive <- true // request received, the server is alive
	var t task.Task
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cs.Enqueue <- &t

	// return a message
	fmt.Fprintf(w, "\n[*] Task added \n")
}

func (cs *ChannelSet) getTask(w http.ResponseWriter, r *http.Request) {
	cs.alive <- true // request received, the server is alive
	var t struct {
		ID int `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cs.List <- nil
	tasks := <-cs.List

	if _, ok := tasks[t.ID]; !ok {
		fmt.Fprintf(w, "\n[*] Task not found \n")
		return
	}

	headerStr := "[ID]\tCommand\tPath\tQueue\tExec\tExitCode\tStatus\tError"
	fmt.Fprintf(w, "\n%s \n\n%s", headerStr, tasks[t.ID].LongString())
}

func (cs *ChannelSet) listTasks(w http.ResponseWriter, r *http.Request) {
	cs.alive <- true // request received, the server is alive
	var t struct {
		Short bool `json:"short"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// request the list of tasks
	cs.List <- nil
	tasks := <-cs.List

	headerStr, tasksStr := fmtTaskList(tasks, t.Short)

	fmt.Fprintf(w, "%s\n%s ", headerStr, tasksStr)
}

func (cs *ChannelSet) listHistory(w http.ResponseWriter, r *http.Request) {
	cs.alive <- true // request received, the server is alive

	// request the list of tasks
	cs.List <- nil
	actualTasks := <-cs.List

	var until int64
	if len(actualTasks) > 0 {
		// get the lesser key of the map
		minKey := 999999
		for k := range actualTasks {
			if k < minKey {
				minKey = k
			}
		}
		until = actualTasks[minKey].EnqueueTime
	} else {
		until = time.Now().UnixMilli()
	}
	tasks, err := task.History(until)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	headerStr, tasksStr := fmtTaskList(tasks, false)
	fmt.Fprintf(w, "\n%s \n\n%s ", headerStr, tasksStr)
}

func (cs *ChannelSet) removeTask(w http.ResponseWriter, r *http.Request) {
	cs.alive <- true // request received, the server is alive
	var t struct {
		ID int `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cs.Remove <- t.ID
	cs.List <- nil
	tasks := <-cs.List

	if _, ok := tasks[t.ID]; !ok {
		fmt.Fprintf(w, "\n[*] Task not found \n")
		return
	} else {
		fmt.Fprintf(w, "\n[*] Task removed \n")
		return
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	// get operation to check if the server is online,
	// return a 200 status
	fmt.Fprintf(w, "\n[*] Pong\n")
}

func clean(w http.ResponseWriter, r *http.Request) {
	// clean the storage
	task.CleanStorage()
	fmt.Fprintf(w, "\n[*] Storage cleaned\n")
}

func NewServer(enqueue chan<- *task.Task, remove chan<- int, list chan map[int]task.Task) chan bool {

	cs := &ChannelSet{
		Enqueue: enqueue,
		Remove:  remove,
		List:    list,
	}

	shutdown := make(chan bool)

	// serve the http sever on a goroutine
	go func() {

		// every time a request is received, send a signal on the alive channel
		// if the server doesn't receive a signal on the alive channel for 10 minutes, it shuts down
		alive := make(chan bool)
		cs.alive = alive
		timeout := time.After(1 * time.Minute)

		// shutdown the server if it doesn't receive a signal on the alive channel for 10 minutes
		go func() {
			for {
				select {
				case <-timeout:
					// check if a task is still running
					cs.List <- nil
					tasks := <-cs.List

					stillRunning := false
					for _, t := range tasks {
						if t.Status == "running" {
							stillRunning = true
						}
					}

					if stillRunning {
						fmt.Println("\n[*] A task is still running: postponing timeout...")
						timeout = time.After(1 * time.Minute)
					} else {
						fmt.Println("\n[*] Server timeout...")
						shutdown <- true
					}

				case <-alive:
					// reset timeout
					timeout = time.After(1 * time.Minute)
				}
			}
		}()

		mux := http.NewServeMux()
		mux.HandleFunc("/add", cs.addTask)
		mux.HandleFunc("/get", cs.getTask)
		mux.HandleFunc("/list", cs.listTasks)
		mux.HandleFunc("/remove", cs.removeTask)
		mux.HandleFunc("/history", cs.listHistory)
		mux.HandleFunc("/clean", clean)
		mux.HandleFunc("/ping", ping)
		http.ListenAndServe(":8080", mux)
	}()

	return shutdown
}
