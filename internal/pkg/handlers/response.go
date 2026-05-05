package handlers

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, err string) {
	JSON(w, status, Response{
		Success: false,
		Error:   err,
	})
}

func ValidationError(w http.ResponseWriter, field, message string) {
	JSON(w, http.StatusBadRequest, Response{
		Success: false,
		Error:   "validation error",
		Data: map[string]string{
			"field":   field,
			"message": message,
		},
	})
}
