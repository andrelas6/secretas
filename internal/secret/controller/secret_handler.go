package controller

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type SecretHandler struct {
	// when this needs deps, it will come here
}

func (h SecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	secret, err := decodeSecretDTO(r)

	if err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(secret); err != nil {
		log.Printf("encoding secret response: %v", err)
	}
}

func decodeSecretDTO(r *http.Request) (SecretDTO, error) {
	var secretDto SecretDTO
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&secretDto); err != nil {
		return SecretDTO{}, errors.New("invalid json")
	}

	return secretDto, nil
}
