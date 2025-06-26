package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Kaowarstail/Only-Flick-Go/config"
)

// CORS configure les en-têtes CORS pour permettre les requêtes cross-origin
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Charger la configuration
		cfg := config.Get()
		
		// Obtenir l'origine de la requête
		origin := r.Header.Get("Origin")
		
		// Vérifier si l'origine est autorisée
		originAllowed := false
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}
		
		// En développement, accepter tous les localhost si pas d'origine spécifique configurée
		if !originAllowed && origin != "" {
			if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0") {
				originAllowed = true
			}
		}
		
		// Configuration des en-têtes CORS
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(cfg.CORS.AllowedOrigins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", cfg.CORS.AllowedOrigins[0])
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Bearer, X-Requested-With")
		
		// Configurer les credentials selon la configuration
		if cfg.CORS.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		// Configurer le cache preflight selon la configuration
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.CORS.MaxAge))
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
