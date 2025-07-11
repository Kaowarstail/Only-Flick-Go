-- Migration pour le système de messagerie OnlyFlick
-- Date: 2024-01-15
-- Auteur: System

-- ========== Table conversations_classic ==========
CREATE TABLE IF NOT EXISTS conversations_classic (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_name VARCHAR(255),
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index sur les conversations
CREATE INDEX IF NOT EXISTS idx_conversations_classic_created_at ON conversations_classic(created_at);
CREATE INDEX IF NOT EXISTS idx_conversations_classic_updated_at ON conversations_classic(updated_at);
CREATE INDEX IF NOT EXISTS idx_conversations_classic_is_group ON conversations_classic(is_group);

-- ========== Table conversation_classic_participants ==========
CREATE TABLE IF NOT EXISTS conversation_classic_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Contraintes de clés étrangères
    CONSTRAINT fk_conversation_classic_participants_conversation 
        FOREIGN KEY (conversation_id) REFERENCES conversations_classic(id) ON DELETE CASCADE,
    CONSTRAINT fk_conversation_classic_participants_user 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Contrainte d'unicité pour éviter les doublons
    CONSTRAINT unique_conversation_user UNIQUE (conversation_id, user_id)
);

-- Index sur les participants
CREATE INDEX IF NOT EXISTS idx_conversation_classic_participants_conversation_id ON conversation_classic_participants(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_classic_participants_user_id ON conversation_classic_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_classic_participants_joined_at ON conversation_classic_participants(joined_at);

-- ========== Table messages_classic ==========
CREATE TABLE IF NOT EXISTS messages_classic (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT,
    message_type VARCHAR(50) NOT NULL DEFAULT 'text',
    media_url VARCHAR(500),
    media_type VARCHAR(100),
    media_size BIGINT,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Contraintes de clés étrangères
    CONSTRAINT fk_messages_classic_conversation 
        FOREIGN KEY (conversation_id) REFERENCES conversations_classic(id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_classic_sender 
        FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Contrainte de validation du type de message
    CONSTRAINT check_message_type 
        CHECK (message_type IN ('text', 'image', 'video', 'audio', 'file')),
    
    -- Contrainte pour s'assurer qu'il y a du contenu ou un média
    CONSTRAINT check_message_content 
        CHECK (
            (content IS NOT NULL AND TRIM(content) != '') OR 
            (media_url IS NOT NULL AND TRIM(media_url) != '')
        )
);

-- Index sur les messages
CREATE INDEX IF NOT EXISTS idx_messages_classic_conversation_id ON messages_classic(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_classic_sender_id ON messages_classic(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_classic_created_at ON messages_classic(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_classic_message_type ON messages_classic(message_type);
CREATE INDEX IF NOT EXISTS idx_messages_classic_is_deleted ON messages_classic(is_deleted);

-- Index composite pour optimiser les requêtes de récupération de messages
CREATE INDEX IF NOT EXISTS idx_messages_classic_conversation_created 
    ON messages_classic(conversation_id, created_at DESC);

-- Index pour la recherche de contenu
CREATE INDEX IF NOT EXISTS idx_messages_classic_content_search 
    ON messages_classic USING GIN (to_tsvector('french', content)) 
    WHERE content IS NOT NULL AND NOT is_deleted;

-- ========== Table conversation_classic_read_status ==========
CREATE TABLE IF NOT EXISTS conversation_classic_read_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,
    last_read_message_id UUID,
    last_read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Contraintes de clés étrangères
    CONSTRAINT fk_conversation_classic_read_status_conversation 
        FOREIGN KEY (conversation_id) REFERENCES conversations_classic(id) ON DELETE CASCADE,
    CONSTRAINT fk_conversation_classic_read_status_user 
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_conversation_classic_read_status_message 
        FOREIGN KEY (last_read_message_id) REFERENCES messages_classic(id) ON DELETE SET NULL,
    
    -- Contrainte d'unicité
    CONSTRAINT unique_conversation_user_read_status UNIQUE (conversation_id, user_id)
);

-- Index sur les statuts de lecture
CREATE INDEX IF NOT EXISTS idx_conversation_classic_read_status_conversation_id ON conversation_classic_read_status(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_classic_read_status_user_id ON conversation_classic_read_status(user_id);
CREATE INDEX IF NOT EXISTS idx_conversation_classic_read_status_last_read_at ON conversation_classic_read_status(last_read_at);

-- ========== Triggers pour mise à jour automatique ==========

-- Trigger pour mettre à jour updated_at dans conversations_classic
CREATE OR REPLACE FUNCTION update_conversation_classic_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_conversation_classic_updated_at
    BEFORE UPDATE ON conversations_classic
    FOR EACH ROW
    EXECUTE FUNCTION update_conversation_classic_updated_at();

-- Trigger pour mettre à jour la conversation quand un nouveau message arrive
CREATE OR REPLACE FUNCTION update_conversation_on_new_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations_classic 
    SET updated_at = NEW.created_at 
    WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_conversation_on_new_message
    AFTER INSERT ON messages_classic
    FOR EACH ROW
    EXECUTE FUNCTION update_conversation_on_new_message();

-- ========== Fonctions utilitaires ==========

-- Fonction pour obtenir le nombre de messages non lus
CREATE OR REPLACE FUNCTION get_unread_messages_count(
    p_conversation_id UUID, 
    p_user_id UUID
) RETURNS INTEGER AS $$
DECLARE
    last_read_at TIMESTAMP WITH TIME ZONE;
    unread_count INTEGER;
BEGIN
    -- Récupérer la date de dernière lecture
    SELECT crs.last_read_at INTO last_read_at
    FROM conversation_classic_read_status crs
    WHERE crs.conversation_id = p_conversation_id 
    AND crs.user_id = p_user_id;
    
    -- Si pas de statut de lecture, compter tous les messages
    IF last_read_at IS NULL THEN
        SELECT COUNT(*)::INTEGER INTO unread_count
        FROM messages_classic m
        WHERE m.conversation_id = p_conversation_id
        AND m.sender_id != p_user_id
        AND NOT m.is_deleted;
    ELSE
        -- Compter les messages après la dernière lecture
        SELECT COUNT(*)::INTEGER INTO unread_count
        FROM messages_classic m
        WHERE m.conversation_id = p_conversation_id
        AND m.sender_id != p_user_id
        AND m.created_at > last_read_at
        AND NOT m.is_deleted;
    END IF;
    
    RETURN COALESCE(unread_count, 0);
END;
$$ LANGUAGE plpgsql;

-- Fonction pour nettoyer les anciennes données (optionnel)
CREATE OR REPLACE FUNCTION cleanup_old_messaging_data(days_to_keep INTEGER DEFAULT 365)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Supprimer les messages supprimés plus vieux que X jours
    WITH deleted_messages AS (
        DELETE FROM messages_classic 
        WHERE is_deleted = TRUE 
        AND updated_at < NOW() - INTERVAL '1 day' * days_to_keep
        RETURNING id
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted_messages;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ========== Commentaires de documentation ==========
COMMENT ON TABLE conversations_classic IS 'Table principale pour les conversations de messagerie classique';
COMMENT ON TABLE conversation_classic_participants IS 'Table de liaison pour les participants aux conversations';
COMMENT ON TABLE messages_classic IS 'Table des messages dans les conversations';
COMMENT ON TABLE conversation_classic_read_status IS 'Table de suivi des statuts de lecture par utilisateur';

COMMENT ON COLUMN conversations_classic.conversation_name IS 'Nom de la conversation (optionnel, utilisé pour les groupes)';
COMMENT ON COLUMN conversations_classic.is_group IS 'Indique si la conversation est un groupe ou une conversation directe';

COMMENT ON COLUMN messages_classic.content IS 'Contenu textuel du message';
COMMENT ON COLUMN messages_classic.message_type IS 'Type de message: text, image, video, audio, file';
COMMENT ON COLUMN messages_classic.media_url IS 'URL du fichier média (si applicable)';
COMMENT ON COLUMN messages_classic.media_type IS 'Type MIME du fichier média';
COMMENT ON COLUMN messages_classic.media_size IS 'Taille du fichier média en bytes';
COMMENT ON COLUMN messages_classic.is_deleted IS 'Indique si le message a été supprimé (soft delete)';

-- ========== Données d'exemple (optionnel pour les tests) ==========
/*
-- Exemple d'insertion de données de test
INSERT INTO conversations_classic (id, conversation_name, is_group) VALUES
('123e4567-e89b-12d3-a456-426614174000', NULL, FALSE),
('123e4567-e89b-12d3-a456-426614174001', 'Groupe Test', TRUE);

-- Continuer avec les participants et messages...
*/
