package main

import "time"

type HealthResponce struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type UploadResponce struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type EditResponce struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type DeleteResponce struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type GetAllFilesResponce struct {
	Files []string `json:"files"`
}

type ErrorResponce struct {
	Error string `json:"error"`
}
