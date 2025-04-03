package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"os"
	"github.com/joho/godotenv"
)

type WebhookRequest struct {
	Plan string `json:"plan"`
}

var (
	sounds = map[string]string{
		"upsell": "upsell.wav",
	}
	authToken = "your-secret-token" // Change this to your desired token
)

func PlaySound(filePath string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" { // macOS
		cmd = exec.Command("afplay", filePath)
	} else { // Linux (Raspberry Pi)
		cmd = exec.Command("aplay", filePath)
	}
	err := cmd.Start()
	if err != nil {
		log.Println("Ошибка воспроизведения:", err)
		return
	}

	log.Println("🔊 Звук запущен в фоновом режиме:", filePath)
} 

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Check for token in Authorization header
	token := r.Header.Get("Authorization")
	if token != "Bearer "+authToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WebhookRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	soundFile, exists := sounds[req.Plan]
	if !exists {
		http.Error(w, "Invalid plan", http.StatusBadRequest)
		return
	}

	soundPath, _ := filepath.Abs(filepath.Join("sounds", soundFile))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sound is playing in the background"))

	go PlaySound(soundPath)
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// Get token from environment variable
	if envToken := os.Getenv("WEBHOOK_TOKEN"); envToken != "" {
		authToken = envToken
	} else {
		log.Fatal("WEBHOOK_TOKEN environment variable is required")
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	port = ":" + port

	http.HandleFunc("/webhook", webhookHandler)

	fmt.Println("🔊 Server is running on port: ", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
