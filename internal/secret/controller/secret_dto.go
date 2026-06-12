package controller

type SecretDTO struct {
	Name           string   `json:"name"`
	EncryptedValue string   `json:"encrypted_value"`
	IV             string   `json:"iv"`
	Tags           []string `json:"tags"`
	NotesEncrypted string   `json:"notes_encrypted"`
}
