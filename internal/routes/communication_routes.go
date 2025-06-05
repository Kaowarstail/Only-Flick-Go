package routes

import (
	"net/http"

	"github.com/Kaowarstail/Only-Flick-Go/internal/middleware"
	"github.com/gorilla/mux"
)

// RegisterCommunicationRoutes enregistre toutes les routes pour la communication (messages et notifications)
func RegisterCommunicationRoutes(router *mux.Router) {
	// Routes pour la messagerie (toutes protégées)
	router.Handle("/messages", middleware.JWTAuth(handleGetConversations)).Methods("GET")
	router.Handle("/messages", middleware.JWTAuth(handleSendMessage)).Methods("POST")
	router.Handle("/messages/{userId}", middleware.JWTAuth(handleGetUserMessages)).Methods("GET")
	router.Handle("/messages/{id}/read", middleware.JWTAuth(handleMarkMessageRead)).Methods("PUT")
	router.Handle("/messages/{id}", middleware.JWTAuth(handleDeleteMessage)).Methods("DELETE")

	// Routes pour les notifications (toutes protégées)
	router.Handle("/notifications", middleware.JWTAuth(handleGetNotifications)).Methods("GET")
	router.Handle("/notifications/{id}/read", middleware.JWTAuth(handleMarkNotificationRead)).Methods("PUT")
	router.Handle("/notifications/read-all", middleware.JWTAuth(handleMarkAllNotificationsRead)).Methods("PUT")
	router.Handle("/notifications/unread-count", middleware.JWTAuth(handleGetUnreadNotificationsCount)).Methods("GET")
}

// Définitions temporaires des gestionnaires
var (
	// Message handlers
	handleGetConversations = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleSendMessage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetUserMessages = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleMarkMessageRead = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleDeleteMessage = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})

	// Notification handlers
	handleGetNotifications = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleMarkNotificationRead = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleMarkAllNotificationsRead = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
	handleGetUnreadNotificationsCount = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Endpoint non implémenté"))
	})
)
