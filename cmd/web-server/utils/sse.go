package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sipeed/picoclaw/cmd/web-server/models"
)

// SendSSEThought 发送SSE思考过程
func SendSSEThought(w http.ResponseWriter, flusher http.Flusher, thought models.Thought) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":    "thought",
		"thought": thought,
	})

	sseData := string(data)
	log.Printf("Sending SSE thought: %s", sseData)
	fmt.Fprintf(w, "data: %s\n\n", sseData)
	flusher.Flush()
}

// SendSSEComplete 发送SSE完成消息
func SendSSEComplete(w http.ResponseWriter, flusher http.Flusher, message, model, errorMsg, conversationID string, thoughts []models.Thought) {
	response := map[string]interface{}{
		"type":      "complete",
		"message":   message,
		"model":     model,
		"timestamp": time.Now(),
	}

	if errorMsg != "" {
		response["error"] = errorMsg
	}

	data, _ := json.Marshal(response)
	sseData := string(data)
	log.Printf("Sending SSE complete: %s", sseData)
	fmt.Fprintf(w, "data: %s\n\n", sseData)
	flusher.Flush()
}
