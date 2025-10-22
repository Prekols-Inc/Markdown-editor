package main

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HealthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type CheckAuthResponse struct {
	Authenticated bool `json:"authenticated"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type RefreshResponce struct {
	Message string `json:"message"`
}
