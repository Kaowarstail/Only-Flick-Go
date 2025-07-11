package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Kaowarstail/Only-Flick-Go/internal/services"
	"github.com/gorilla/mux"
)

// MetricsMiddleware middleware pour collecter les métriques HTTP
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrapper pour capturer le code de statut
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Exécuter la requête
		next.ServeHTTP(recorder, r)

		// Calculer la durée
		duration := time.Since(start).Seconds()

		// Obtenir la route et la méthode
		route := getRoute(r)
		method := r.Method
		status := strconv.Itoa(recorder.statusCode)

		// Enregistrer les métriques
		services.RecordHTTPMetrics(method, route, status, duration)
	})
}

// responseRecorder pour capturer le code de statut
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// getRoute extrait la route à partir de la requête
func getRoute(r *http.Request) string {
	// Essayer d'obtenir la route depuis Gorilla Mux
	if route := mux.CurrentRoute(r); route != nil {
		if pathTemplate, err := route.GetPathTemplate(); err == nil {
			return pathTemplate
		}
	}
	
	// Fallback vers le chemin de la requête
	return r.URL.Path
}
