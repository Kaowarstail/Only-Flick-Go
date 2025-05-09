package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger est un middleware qui enregistre les informations sur chaque requête
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Enregistrement de la requête entrante
		log.Printf(
			"[%s] %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
		)

		// Appel du gestionnaire suivant
		next.ServeHTTP(w, r)

		// Enregistrement de la durée de traitement
		log.Printf(
			"[%s] %s %s - Completed in %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}
