package utils

import (
	"context"
	"errors"
)

// GetUserIDFromContext récupère l'ID utilisateur depuis le contexte
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

// GetUserRoleFromContext récupère le rôle utilisateur depuis le contexte
func GetUserRoleFromContext(ctx context.Context) (string, error) {
	userRole, ok := ctx.Value("user_role").(string)
	if !ok || userRole == "" {
		return "", errors.New("user role not found in context")
	}
	return userRole, nil
}

// GetConversationIDFromContext récupère l'ID de conversation depuis le contexte
func GetConversationIDFromContext(ctx context.Context) (string, error) {
	conversationID, ok := ctx.Value("conversation_id").(string)
	if !ok || conversationID == "" {
		return "", errors.New("conversation ID not found in context")
	}
	return conversationID, nil
}
