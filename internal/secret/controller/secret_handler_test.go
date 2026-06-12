package controller_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrelas6/secretas/internal/secret/controller"
)

func TestSecretHandlerSuccess(t *testing.T) {
	req, rec, h := setupTest(`{"name":"aws-key","encrypted_value":"qY3k==","iv":"Hn8s"}`)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Got %d. Wanted %d", rec.Code, http.StatusOK)
	}

	if !strings.Contains(rec.Body.String(), "aws-key") {
		t.Errorf("Got %s. Wanted %s", rec.Body.String(), "aws-key")
	}
}

func TestSecretHandlerErrorsWhenUnknownJsonField(t *testing.T) {
	req, rec, h := setupTest(`{"name":"aws-key","encrypted_value":"qY3k==","iv":"Hn8s", "unknown": "t"}`)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Got %d. Expected %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSecretHandlerErrorsWhenInvalidField(t *testing.T) {
	req, rec, h := setupTest(`{"name": 10,"encrypted_value":"qY3k==","iv":"Hn8s", "unknown": "t"}`)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Got %d. Expected %d", rec.Code, http.StatusBadRequest)
	}
}

func setupTest(body string) (*http.Request, *httptest.ResponseRecorder, controller.SecretHandler) {
	req := httptest.NewRequest("GET", "/secret", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h := controller.SecretHandler{}

	return req, rec, h
}
