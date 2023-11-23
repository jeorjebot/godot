package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
)

func PrintOutput(body string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, body)
}

func AddTaskRequest(endpoint, command, path string) error {
	payload := []byte(fmt.Sprintf(`{"command":"%s", "path":"%s"}`, command, path))
	body, err := SendRequest(endpoint, payload)
	if err != nil {
		log.Fatalf("Error adding task: %v", err)
	}
	PrintOutput(body)
	return err
}

func ListRequest(endpoint string, short bool) error {
	payload := []byte(fmt.Sprintf(`{"short":%t}`, short))
	body, err := SendRequest(endpoint, payload)
	if err != nil {
		return err
	}
	PrintOutput(body)
	return err
}

func HistoryRequest(endpoint string) error {
	payload := []byte(`{}`)
	Body, err := SendRequest(endpoint, payload)
	if err != nil {
		return err
	}
	PrintOutput(Body)
	return err
}

func RemoveTaskRequest(endpoint string, id int) error {
	payload := []byte(fmt.Sprintf(`{"id":%d}`, id))
	body, err := SendRequest(endpoint, payload)
	if err != nil {
		return err
	}
	fmt.Println(body)
	return err
}

func GetTaskRequest(endpoint string, id int) error {
	payload := []byte(fmt.Sprintf(`{"id":%d}`, id))
	body, err := SendRequest(endpoint, payload)
	if err != nil {
		return err
	}
	PrintOutput(body)
	return err
}

func Ping() error {
	endpoint := "ping"
	payload := []byte(`{}`)
	_, err := SendRequest(endpoint, payload)
	return err
}

func Clean() error {
	endpoint := "clean"
	payload := []byte(`{}`)
	_, err := SendRequest(endpoint, payload)
	return err
}

func SendRequest(endpoint string, payload []byte) (string, error) {
	req, err := http.NewRequest("POST", "http://localhost:8080/"+endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// create a http client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
