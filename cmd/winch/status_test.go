package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/healthz" {
			_, _ = w.Write([]byte(`{"service":"winch-sandbox","status":"ok"}`))
		} else {
			_, _ = w.Write([]byte(`{"profile":"container-standard","unenforcedControls":["network-egress"]}`))
		}
	}))
	defer server.Close()
	var out bytes.Buffer
	if err := runStatus([]string{"--url", server.URL}, &out); err != nil {
		t.Fatal(err)
	}
	want := "service: winch-sandbox (ok)\nprofile: container-standard\nunenforced control: network-egress\n"
	if out.String() != want {
		t.Fatalf("got %q", out.String())
	}
}
func TestStatusRejectsInvalidPosture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			_, _ = w.Write([]byte(`{"service":"winch-sandbox","status":"ok"}`))
		} else {
			_, _ = w.Write([]byte(`{"profile":"container-standard","unenforcedControls":[]}`))
		}
	}))
	defer server.Close()
	var out bytes.Buffer
	if err := runStatus([]string{"--url", server.URL}, &out); err == nil {
		t.Fatal("accepted empty controls")
	}
	if out.Len() != 0 {
		t.Fatalf("printed success output: %q", out.String())
	}
}
