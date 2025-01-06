package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	port     = "8080"
	apiToken = "your-secret-api-token"
)

type PDU struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	RawPDU    string    `json:"raw_pdu" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

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

func jsonResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func HandleSubmitPDU(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != apiToken {
		jsonResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	var pduReq PDURequest
	if err := json.NewDecoder(r.Body).Decode(&pduReq); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request payload",
		})
		return
	}

	pdu := PDU{RawPDU: pduReq.RawPDU}
	if err := db.Create(&pdu).Error; err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to save PDU",
		})
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"id":     pdu.ID,
	})
}

func main() {
	InitDatabase()

	http.HandleFunc("/submit-pdu", HandleSubmitPDU)

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
