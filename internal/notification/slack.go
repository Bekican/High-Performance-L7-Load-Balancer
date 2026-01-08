package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackMessage struct {
	Text string `json:"text"`
}

func SendSlackNotification(webhookUrl string, message string) {
	if webhookUrl == "" {
		return
	}

	payload := SlackMessage{Text: message}
	jsonPayload, _ := json.Marshal(payload)

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonPayload))

	if err != nil {
		fmt.Printf("Notification Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Notification Error Status: %d\n", resp.StatusCode)
	}
}
