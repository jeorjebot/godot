package task

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

func UseStorage() (string, error) {
	// if it is the first time the program is run, it creates a .godot folder in the home directory
	// if it is not the first time the program is run, it checks if the .godot folder is present
	// if it is not present, it creates it
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// create the .godot folder inside the home directory
	godotDir := homeDir + "/.godot"
	if _, err := os.Stat(godotDir); os.IsNotExist(err) {
		err := os.Mkdir(godotDir, 0755)
		if err != nil {
			return "", err
		}
	}

	// create the godot_history.log file inside the .godot folder
	// if the file is already present, it does nothing
	_, err = os.OpenFile(godotDir+"/godot_history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	return godotDir, nil

}

func NewLogFile() (*os.File, string) {
	// create a file of the form godot-<timestamp>.log
	// the file is created in the .godot folder

	godotDir, err := UseStorage()
	if err != nil {
		panic(err)
	}
	filename := godotDir + "/" + fmt.Sprintf("%d", time.Now().UnixMilli()) + ".log"
	log, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return log, filename
}

func CleanStorage() {
	// remove all the files in the .godot folder
	godotDir, err := UseStorage()
	if err != nil {
		panic(err)
	}
	// get all the files in the .godot folder
	files, err := os.ReadDir(godotDir)
	if err != nil {
		panic(err)
	}
	// remove all the files in the .godot folder
	for _, file := range files {
		err := os.Remove(godotDir + "/" + file.Name())
		if err != nil {
			panic(err)
		}
	}
	// recreate the godot_history.log file
	_, err = os.OpenFile(godotDir+"/godot_history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

func LogTask(t Task) {
	// log the task object as json
	jsonData, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	// get the godot_history.log file
	godotDir, err := UseStorage()
	if err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(godotDir+"/godot_history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	// write the json data to the file
	fmt.Fprintf(logFile, "%s\n", jsonData)
}

// define a function that read then godot_history.log file
// decode the task objects from the json
// return the task objects that have a timestamp lower than the timestamp in input
func History(timestamp int64) (map[int]Task, error) {
	// get the godot_history.log file
	godotDir, err := UseStorage()
	if err != nil {
		return map[int]Task{}, err
	}
	logFile, err := os.OpenFile(godotDir+"/godot_history.log", os.O_RDONLY, 0644)
	if err != nil {
		return map[int]Task{}, err
	}
	// read the file line by line
	// decode the json into a task object
	// if the timestamp is lower than the timestamp of the task, add the task object to the tasks map
	// if reached the end of the file, return the tasks map
	tasks := map[int]Task{}
	var i int
	for {
		var s string
		for {
			b := make([]byte, 1)
			_, err := logFile.Read(b)
			if err != nil {
				if err == io.EOF {
					return tasks, nil

				}
				return map[int]Task{}, err
			}
			if b[0] == '\n' {
				break
			}

			s += string(b)
		}
		var t Task
		err := json.Unmarshal([]byte(s), &t)
		if err != nil {
			return map[int]Task{}, err
		}
		if t.EndTime < timestamp {
			tasks[i] = t
			i++
		}
		if t.EndTime == 0 {
			break
		}
	}
	return tasks, nil

}

func MaxTaskIDInHistory() int {
	tasks, _ := History(time.Now().UnixMilli())
	max := -1
	for k := range tasks {
		if k > max {
			max = k
		}
	}
	return max
}
