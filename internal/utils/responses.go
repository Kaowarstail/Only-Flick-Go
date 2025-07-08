package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ErrorResponse représente une réponse d'erreur standardisée
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse représente une réponse de succès standardisée
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SendError envoie une réponse d'erreur JSON standardisée
func SendError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error: message,
		Code:  statusCode,
	}

	if err != nil {
		errorResponse.Message = err.Error()
		log.Printf("Error [%d]: %s - %v", statusCode, message, err)
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// SendJSONResponse envoie une réponse JSON standardisée
func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// SendSuccess envoie une réponse de succès JSON standardisée
func SendSuccess(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// ValidateStruct valide une structure en utilisant des tags de validation basiques
func ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Si c'est un pointeur, obtenir l'élément sous-jacent
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", v.Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("validate")

		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			switch {
			case rule == "required":
				if err := validateRequired(field, fieldType.Name); err != nil {
					return err
				}
			case strings.HasPrefix(rule, "min="):
				minStr := strings.TrimPrefix(rule, "min=")
				min, err := strconv.ParseFloat(minStr, 64)
				if err != nil {
					return fmt.Errorf("invalid min value for field %s: %s", fieldType.Name, minStr)
				}
				if err := validateMin(field, fieldType.Name, min); err != nil {
					return err
				}
			case strings.HasPrefix(rule, "max="):
				maxStr := strings.TrimPrefix(rule, "max=")
				max, err := strconv.ParseFloat(maxStr, 64)
				if err != nil {
					return fmt.Errorf("invalid max value for field %s: %s", fieldType.Name, maxStr)
				}
				if err := validateMax(field, fieldType.Name, max); err != nil {
					return err
				}
			case rule == "email":
				if err := validateEmailField(field, fieldType.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateRequired vérifie qu'un champ n'est pas vide
func validateRequired(field reflect.Value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		if field.String() == "" {
			return fmt.Errorf("field %s is required", fieldName)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Pour les entiers, on considère que 0 n'est pas valide pour un champ requis
		if field.Int() == 0 {
			return fmt.Errorf("field %s is required", fieldName)
		}
	case reflect.Float32, reflect.Float64:
		// Pour les floats, on considère que 0.0 n'est pas valide pour un champ requis
		if field.Float() == 0.0 {
			return fmt.Errorf("field %s is required", fieldName)
		}
	case reflect.Bool:
		// Pour les booléens, on accepte false comme valeur valide
		// car c'est souvent intentionnel
	case reflect.Slice, reflect.Array:
		if field.Len() == 0 {
			return fmt.Errorf("field %s is required", fieldName)
		}
	default:
		if field.IsZero() {
			return fmt.Errorf("field %s is required", fieldName)
		}
	}
	return nil
}

// validateMin vérifie qu'une valeur numérique est supérieure ou égale au minimum
func validateMin(field reflect.Value, fieldName string, min float64) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) < min {
			return fmt.Errorf("field %s must be at least %g", fieldName, min)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() < min {
			return fmt.Errorf("field %s must be at least %g", fieldName, min)
		}
	case reflect.String:
		if float64(len(field.String())) < min {
			return fmt.Errorf("field %s must be at least %g characters long", fieldName, min)
		}
	}
	return nil
}

// validateMax vérifie qu'une valeur numérique est inférieure ou égale au maximum
func validateMax(field reflect.Value, fieldName string, max float64) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(field.Int()) > max {
			return fmt.Errorf("field %s must be at most %g", fieldName, max)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() > max {
			return fmt.Errorf("field %s must be at most %g", fieldName, max)
		}
	case reflect.String:
		if float64(len(field.String())) > max {
			return fmt.Errorf("field %s must be at most %g characters long", fieldName, max)
		}
	}
	return nil
}

// validateEmailField vérifie qu'un champ string contient un email valide
func validateEmailField(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("field %s must be a string for email validation", fieldName)
	}

	email := field.String()
	if email != "" && !ValidateEmail(email) {
		return fmt.Errorf("field %s must be a valid email address", fieldName)
	}

	return nil
}

// GenerateTimestamp génère un timestamp Unix
func GenerateTimestamp() int64 {
	return time.Now().Unix()
}
