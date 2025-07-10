package models

import (
	"crypto/rand"
	"fmt"
)

// generateUUID génère un UUID simple pour les models
func generateUUID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Définir les bits de version (4 bits pour la version 4)
	bytes[6] = (bytes[6] & 0x0f) | 0x40

	// Définir les bits de variante (2 bits pour la variante RFC 4122)
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:]), nil
}
