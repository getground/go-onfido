package onfido_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getground/go-onfido"
	"github.com/stretchr/testify/assert"

	"github.com/gorilla/mux"
)

func TestNewSdkToken_NonOKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\": \"things went bad\"}"))
	}))
	defer srv.Close()

	client := onfido.NewClient("123")
	client.Endpoint = srv.URL

	token, err := client.NewSdkToken(context.Background(), "123")
	if err == nil {
		t.Fatal("expected to see an error")
	}
	if token != nil {
		t.Fatal("token returned")
	}
}

func TestNewSdkToken_ApplicantsRetrieved(t *testing.T) {
	expected := onfido.SdkToken{
		ApplicantID: "klj25h2jk5j4k5jk35",
		Token:       "423423m4n234czxKJKDLF",
	}
	expectedJson, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	m := mux.NewRouter()
	m.HandleFunc("/sdk_token", func(w http.ResponseWriter, r *http.Request) {
		var tk onfido.SdkToken
		json.NewDecoder(r.Body).Decode(&tk)
		assert.Equal(t, expected.ApplicantID, tk.ApplicantID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(expectedJson)
	}).Methods("POST")
	srv := httptest.NewServer(m)
	defer srv.Close()

	client := onfido.NewClient("123")
	client.Endpoint = srv.URL

	token, err := client.NewSdkToken(context.Background(), expected.ApplicantID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected.ApplicantID, token.ApplicantID)
	assert.Equal(t, expected.Token, token.Token)
}
