package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type WebhookRequest struct {
	Plan string `json:"plan"`
}

var (
	sounds = map[string]string{
		"airwallex": "airwallex.mp3",
		"checkout":  "checkout.mp3",
		"solidgate": "solidgate.mp3",
		"stripe":    "stripe.mp3",
	}
	authToken  = "your-secret-token"
	currentCmd *exec.Cmd
	cmdMutex   sync.Mutex
)

func PlaySound(filePath string) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if currentCmd != nil && currentCmd.Process != nil {
		err := currentCmd.Process.Kill()
		if err != nil {
			log.Println("Error stopping previous sound:", err)
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("afplay", filePath)
	} else {
		cmd = exec.Command("mpg123", filePath)
	}

	err := cmd.Start()
	if err != nil {
		log.Println("Error playing sound:", err)
		return
	}

	currentCmd = cmd

	log.Println("🔊 Sound started in background:", filePath)

	go func() {
		cmd.Wait()
		cmdMutex.Lock()
		if currentCmd == cmd {
			currentCmd = nil
		}
		cmdMutex.Unlock()
	}()
}

func isWorkingHours() bool {
	loc, err := time.LoadLocation("Asia/Almaty") // GMT+5
	if err != nil {
		log.Println("Error loading timezone:", err)
		return false
	}

	now := time.Now().In(loc)
	hour := now.Hour()

	return hour >= 9 && hour < 20
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
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

	if !isWorkingHours() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request received but ignored - outside working hours"))
		return
	}

	soundPath, _ := filepath.Abs(filepath.Join("/home/yerkinmm/Desktop/raspy3-notificator/sounds", soundFile))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sound is playing in the background"))

	go PlaySound(soundPath)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	if envToken := os.Getenv("WEBHOOK_TOKEN"); envToken != "" {
		authToken = envToken
	} else {
		log.Fatal("WEBHOOK_TOKEN environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	port = ":" + port

	http.HandleFunc("/webhook", webhookHandler)

	fmt.Println("🔊 Server is running on port: ", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
