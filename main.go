package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Configuration
const apiToken = "your-secret-api-token"

type PDU struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	RawPDU    string    `json:"raw_pdu" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Database instance
var db *gorm.DB

func InitDatabase() {
	var err error
	db, err = gorm.Open(sqlite.Open("pdus.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&PDU{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}

type PDURequest struct {
	RawPDU string `json:"raw_pdu"`
}

// HandleSubmitPDU handles POST requests to submit PDUs
func HandleSubmitPDU(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != apiToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var pduReq PDURequest
	if err := json.NewDecoder(r.Body).Decode(&pduReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	pdu := PDU{RawPDU: pduReq.RawPDU}
	if err := db.Create(&pdu).Error; err != nil {
		http.Error(w, "Failed to save PDU", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "PDU saved successfully")
}

func main() {
	InitDatabase()

	http.HandleFunc("/submit-pdu", HandleSubmitPDU)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
