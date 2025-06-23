package middleware

import (
	"net/http"
)

// CORS configure les en-têtes CORS pour permettre les requêtes cross-origin
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtenir l'origine de la requête
		origin := r.Header.Get("Origin")
		
		// Liste des origines autorisées (en développement, on accepte localhost sur tous les ports)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:49219", // Port de votre app Flutter Web
			"http://127.0.0.1:3000",
			"http://127.0.0.1:49219",
		}
		
		// Vérifier si l'origine est autorisée ou si c'est en développement
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}
		
		// En développement, accepter tous les localhost
		if !originAllowed && origin != "" {
			if len(origin) > 16 && origin[:16] == "http://localhost" {
				originAllowed = true
			}
			if len(origin) > 14 && origin[:14] == "http://127.0.0" {
				originAllowed = true
			}
		}
		
		// Configuration des en-têtes CORS
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:49219")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Bearer, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight pour 24h
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		// Gérer les requêtes preflight OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Appeler le gestionnaire suivant
		next.ServeHTTP(w, r)
	})
}
