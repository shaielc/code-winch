package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

type health struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}
type posture struct {
	Profile            string   `json:"profile"`
	UnenforcedControls []string `json:"unenforcedControls"`
}

func runStatus(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	baseURL := flags.String("url", "http://127.0.0.1:8080", "sandbox URL")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return errors.New("invalid status arguments")
	}
	var h health
	if err := getJSON(strings.TrimRight(*baseURL, "/")+"/healthz", &h); err != nil || h.Service != "winch-sandbox" || h.Status != "ok" {
		return errors.New("sandbox is unhealthy")
	}
	var p posture
	if err := getJSON(strings.TrimRight(*baseURL, "/")+"/api/session", &p); err != nil || p.Profile != "container-standard" || len(p.UnenforcedControls) == 0 {
		return errors.New("sandbox posture is invalid")
	}
	for _, control := range p.UnenforcedControls {
		if control == "" {
			return errors.New("sandbox posture is invalid")
		}
	}
	if _, err := fmt.Fprintf(output, "service: %s (%s)\nprofile: %s\n", h.Service, h.Status, p.Profile); err != nil {
		return err
	}
	for _, control := range p.UnenforcedControls {
		if _, err := fmt.Fprintf(output, "unenforced control: %s\n", control); err != nil {
			return err
		}
	}
	return nil
}

func getJSON(url string, target any) error {
	response, err := newClient().Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errors.New("unexpected response")
	}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("malformed response")
	}
	return nil
}
