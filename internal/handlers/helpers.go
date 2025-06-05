package handlers

import (
    "encoding/json"
    "net/http"
)

// RespondWithError envoie une réponse d'erreur au format JSON
func RespondWithError(w http.ResponseWriter, code int, message string) {
    RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON envoie une réponse au format JSON
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, err := json.Marshal(payload)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Erreur lors de la sérialisation de la réponse"))
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}