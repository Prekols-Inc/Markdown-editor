package main

import "time"

type HealthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type UploadResponse struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type EditResponse struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type DeleteResponse struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type GetAllFilesResponse struct {
	Files []string `json:"files"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageReponse struct {
	Message string `json:"message"`
}
