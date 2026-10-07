package main

import (
	"net/http"
	"time"
)

func newClient() *http.Client { return &http.Client{Timeout: 3 * time.Second} }
