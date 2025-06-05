// filepath: /home/kaowarstail/Documents/PEC2/project/Only-Flick-Go/cmd/api/routes.go
package main

import (
	"github.com/Kaowarstail/Only-Flick-Go/internal/routes"
	"github.com/gorilla/mux"
)

// registerRoutes délègue l'enregistrement des routes au package routes
func registerRoutes(router *mux.Router) {
	// Délégation à notre package routes
	routes.RegisterRoutes(router)
}
