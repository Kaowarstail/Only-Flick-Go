package utils

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// ValidateEmail valide le format d'un email
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePassword valide la force d'un mot de passe
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Le mot de passe doit contenir au moins 8 caractères"
	}

	if len(password) > 100 {
		return false, "Le mot de passe ne peut pas dépasser 100 caractères"
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Le mot de passe doit contenir au moins une majuscule"
	}
	if !hasLower {
		return false, "Le mot de passe doit contenir au moins une minuscule"
	}
	if !hasNumber {
		return false, "Le mot de passe doit contenir au moins un chiffre"
	}
	if !hasSpecial {
		return false, "Le mot de passe doit contenir au moins un caractère spécial"
	}

	return true, ""
}

// ValidateUsername valide le format d'un nom d'utilisateur
func ValidateUsername(username string) (bool, string) {
	if len(username) < 3 {
		return false, "Le nom d'utilisateur doit contenir au moins 3 caractères"
	}

	if len(username) > 30 {
		return false, "Le nom d'utilisateur ne peut pas dépasser 30 caractères"
	}

	// Seuls les lettres, chiffres, underscores et tirets sont autorisés
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return false, "Le nom d'utilisateur ne peut contenir que des lettres, chiffres, underscores et tirets"
	}

	// Ne peut pas commencer ou finir par un underscore ou un tiret
	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, "-") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, "-") {
		return false, "Le nom d'utilisateur ne peut pas commencer ou finir par un underscore ou un tiret"
	}

	return true, ""
}

// ValidateName valide le format d'un prénom ou nom
func ValidateName(name string) (bool, string) {
	if len(name) == 0 {
		return true, "" // Les noms sont optionnels
	}

	if len(name) > 50 {
		return false, "Le nom ne peut pas dépasser 50 caractères"
	}

	// Seules les lettres, espaces, apostrophes et tirets sont autorisés
	nameRegex := regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s'-]+$`)
	if !nameRegex.MatchString(name) {
		return false, "Le nom ne peut contenir que des lettres, espaces, apostrophes et tirets"
	}

	return true, ""
}

// SanitizeInput nettoie les entrées utilisateur
func SanitizeInput(input string) string {
	// Supprimer les espaces en début et fin
	input = strings.TrimSpace(input)

	// Remplacer les caractères de contrôle
	input = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(input, "")

	return input
}

// ValidateUUID vérifie si une string est un UUID valide
func ValidateUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// GenerateUUID génère un UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// TimeNow retourne le temps actuel UTC
func TimeNow() time.Time {
	return time.Now().UTC()
}

// ContainsIgnoreCase vérifie si une string contient une autre (insensible à la casse)
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// IsValidURL vérifie si une string est une URL valide
func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	// Simple validation pour URL
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}

// SanitizeMessageContent nettoie le contenu d'un message
func SanitizeMessageContent(content string) string {
	// Supprimer les espaces en début et fin
	content = strings.TrimSpace(content)

	// Supprimer les caractères de contrôle mais garder les retours à la ligne
	content = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(content, "")

	return content
}
