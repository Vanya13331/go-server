package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestHappyPath(t *testing.T) {
	srv := httptest.NewServer(handlers())
	defer srv.Close()

	url := fmt.Sprintf("%s/api/slow", srv.URL)
	requestBody, err := json.Marshal(map[string]uint64{
		"timeout": 4000,
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("http status is not 200")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(body), `{"status": "ok"}`, "wrong body")
}

func TestSadPath(t *testing.T) {
	srv := httptest.NewServer(handlers())
	defer srv.Close()

	url := fmt.Sprintf("%s/api/slow", srv.URL)
	requestBody, err := json.Marshal(map[string]uint64{
		"timeout": 10000,
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("http status is not 400")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(body), `{"error": "timeout too long"}`, "wrong body")
}
