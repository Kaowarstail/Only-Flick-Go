-- Migration 002: Ajout des tables pour la messagerie et l'édition de profil
-- Date: 2025-01-08
-- Description: Création des tables pour conversations, messages, liens sociaux, statistiques utilisateur et transactions

-- Table des conversations
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id UUID NOT NULL,
    user2_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_message_at TIMESTAMP WITH TIME ZONE,
    last_message_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Contraintes
    CONSTRAINT unique_conversation UNIQUE(user1_id, user2_id),
    CONSTRAINT different_users CHECK (user1_id != user2_id),
    
    -- Clés étrangères (à adapter selon votre schema existant)
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index pour les conversations
CREATE INDEX IF NOT EXISTS idx_conversations_user1 ON conversations(user1_id);
CREATE INDEX IF NOT EXISTS idx_conversations_user2 ON conversations(user2_id);
CREATE INDEX IF NOT EXISTS idx_conversations_last_message ON conversations(last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversations_active ON conversations(is_active) WHERE is_active = TRUE;

-- Table des messages
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT,
    media_url TEXT,
    media_type VARCHAR(20), -- 'image', 'video', 'audio', 'document'
    thumbnail_url TEXT,
    
    -- Messages payants
    is_paid BOOLEAN DEFAULT FALSE,
    price DECIMAL(10,2),
    is_unlocked BOOLEAN DEFAULT FALSE,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    unlocked_by UUID,
    preview_text TEXT,
    
    -- Métadonnées
    message_type VARCHAR(30) DEFAULT 'text', -- 'text', 'image', 'video', 'audio', 'document', 'paid_text', 'paid_media'
    status VARCHAR(20) DEFAULT 'sent', -- 'sending', 'sent', 'delivered', 'read', 'failed'
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,
    
    -- Contraintes
    CONSTRAINT valid_price CHECK (price IS NULL OR price >= 0.99),
    CONSTRAINT valid_paid_message CHECK (
        (is_paid = FALSE) OR 
        (is_paid = TRUE AND price IS NOT NULL AND price >= 0.99)
    ),
    CONSTRAINT valid_content CHECK (
        content IS NOT NULL OR media_url IS NOT NULL
    ),
    
    -- Clés étrangères
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (unlocked_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Index pour les messages
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(conversation_id, read_at) WHERE read_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_messages_paid ON messages(is_paid, is_unlocked) WHERE is_paid = TRUE;
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);

-- Table des liens sociaux
CREATE TABLE IF NOT EXISTS social_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    links JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Contraintes
    UNIQUE(user_id),
    
    -- Clés étrangères
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Table des statistiques utilisateur
CREATE TABLE IF NOT EXISTS user_stats (
    user_id UUID PRIMARY KEY,
    followers_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    posts_count INTEGER DEFAULT 0,
    total_messages_sent INTEGER DEFAULT 0,
    total_messages_received INTEGER DEFAULT 0,
    total_conversations INTEGER DEFAULT 0,
    total_media_uploaded INTEGER DEFAULT 0,
    total_media_size BIGINT DEFAULT 0,
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Contraintes
    CONSTRAINT positive_counts CHECK (
        followers_count >= 0 AND 
        following_count >= 0 AND 
        posts_count >= 0 AND
        total_messages_sent >= 0 AND
        total_messages_received >= 0 AND
        total_conversations >= 0 AND
        total_media_uploaded >= 0 AND
        total_media_size >= 0
    ),
    
    -- Clés étrangères
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index pour les statistiques utilisateur
CREATE INDEX IF NOT EXISTS idx_user_stats_last_active ON user_stats(last_active_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_stats_followers ON user_stats(followers_count DESC);

-- Table des transactions de messages payants
CREATE TABLE IF NOT EXISTS paid_message_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL,
    buyer_id UUID NOT NULL,
    creator_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    platform_fee DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    creator_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Statut et métadonnées
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'completed', 'failed', 'refunded'
    payment_method VARCHAR(50),
    payment_provider VARCHAR(50),
    transaction_id VARCHAR(255),
    provider_transaction_id VARCHAR(255),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    
    -- Contraintes
    CONSTRAINT positive_amounts CHECK (
        amount > 0 AND 
        platform_fee >= 0 AND 
        creator_amount >= 0 AND
        amount = platform_fee + creator_amount
    ),
    CONSTRAINT valid_currency CHECK (currency IN ('USD', 'EUR', 'GBP', 'CAD')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    
    -- Clés étrangères
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (buyer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index pour les transactions
CREATE INDEX IF NOT EXISTS idx_transactions_message ON paid_message_transactions(message_id);
CREATE INDEX IF NOT EXISTS idx_transactions_buyer ON paid_message_transactions(buyer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_creator ON paid_message_transactions(creator_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON paid_message_transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created ON paid_message_transactions(created_at DESC);

-- Table des gains des créateurs
CREATE TABLE IF NOT EXISTS creator_earnings (
    creator_id UUID PRIMARY KEY,
    total_earnings DECIMAL(12,2) DEFAULT 0.00,
    total_messages_earnings DECIMAL(12,2) DEFAULT 0.00,
    total_subscriptions_earnings DECIMAL(12,2) DEFAULT 0.00,
    total_tips_earnings DECIMAL(12,2) DEFAULT 0.00,
    current_month_earnings DECIMAL(12,2) DEFAULT 0.00,
    last_month_earnings DECIMAL(12,2) DEFAULT 0.00,
    pending_earnings DECIMAL(12,2) DEFAULT 0.00,
    withdrawn_earnings DECIMAL(12,2) DEFAULT 0.00,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_payout_at TIMESTAMP WITH TIME ZONE,
    
    -- Contraintes
    CONSTRAINT positive_earnings CHECK (
        total_earnings >= 0 AND
        total_messages_earnings >= 0 AND
        total_subscriptions_earnings >= 0 AND
        total_tips_earnings >= 0 AND
        current_month_earnings >= 0 AND
        last_month_earnings >= 0 AND
        pending_earnings >= 0 AND
        withdrawn_earnings >= 0
    ),
    
    -- Clés étrangères
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Table des gains mensuels
CREATE TABLE IF NOT EXISTS monthly_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    total_earnings DECIMAL(12,2) DEFAULT 0.00,
    messages_earnings DECIMAL(12,2) DEFAULT 0.00,
    subscriptions_earnings DECIMAL(12,2) DEFAULT 0.00,
    tips_earnings DECIMAL(12,2) DEFAULT 0.00,
    transactions_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Contraintes
    CONSTRAINT valid_month CHECK (month >= 1 AND month <= 12),
    CONSTRAINT valid_year CHECK (year >= 2020),
    CONSTRAINT positive_monthly_earnings CHECK (
        total_earnings >= 0 AND
        messages_earnings >= 0 AND
        subscriptions_earnings >= 0 AND
        tips_earnings >= 0 AND
        transactions_count >= 0
    ),
    UNIQUE(creator_id, year, month),
    
    -- Clés étrangères
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index pour les gains mensuels
CREATE INDEX IF NOT EXISTS idx_monthly_earnings_creator ON monthly_earnings(creator_id, year DESC, month DESC);
CREATE INDEX IF NOT EXISTS idx_monthly_earnings_period ON monthly_earnings(year DESC, month DESC);

-- Table des fichiers média
CREATE TABLE IF NOT EXISTS media_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    media_type VARCHAR(20) NOT NULL, -- 'image', 'video', 'audio', 'document'
    mime_type VARCHAR(100) NOT NULL,
    thumbnail_url TEXT,
    
    -- Métadonnées
    metadata JSONB,
    is_processed BOOLEAN DEFAULT FALSE,
    processing_status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    
    -- Timestamps
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Contraintes
    CONSTRAINT positive_file_size CHECK (file_size > 0),
    CONSTRAINT valid_media_type CHECK (media_type IN ('image', 'video', 'audio', 'document')),
    CONSTRAINT valid_processing_status CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed')),
    
    -- Clés étrangères
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index pour les fichiers média
CREATE INDEX IF NOT EXISTS idx_media_files_user ON media_files(user_id);
CREATE INDEX IF NOT EXISTS idx_media_files_type ON media_files(media_type);
CREATE INDEX IF NOT EXISTS idx_media_files_uploaded ON media_files(uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_files_processing ON media_files(processing_status) WHERE processing_status != 'completed';

-- Trigger pour mettre à jour automatiquement les timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Appliquer le trigger aux tables appropriées
CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_social_links_updated_at BEFORE UPDATE ON social_links
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_stats_updated_at BEFORE UPDATE ON user_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON paid_message_transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_creator_earnings_updated_at BEFORE UPDATE ON creator_earnings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_monthly_earnings_updated_at BEFORE UPDATE ON monthly_earnings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_media_files_updated_at BEFORE UPDATE ON media_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Fonction pour calculer automatiquement les frais de plateforme (20%)
CREATE OR REPLACE FUNCTION calculate_platform_fee()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculer les frais de plateforme (20%)
    NEW.platform_fee = NEW.amount * 0.20;
    NEW.creator_amount = NEW.amount - NEW.platform_fee;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Appliquer le trigger aux transactions
CREATE TRIGGER calculate_transaction_fees BEFORE INSERT ON paid_message_transactions
    FOR EACH ROW EXECUTE FUNCTION calculate_platform_fee();

-- Créer des vues pour faciliter les requêtes

-- Vue pour les conversations avec le dernier message
CREATE OR REPLACE VIEW conversation_details AS
SELECT 
    c.*,
    m.content as last_message_content,
    m.sender_id as last_message_sender_id,
    m.created_at as last_message_created_at,
    u1.username as user1_username,
    u1.profile_picture as user1_profile_picture,
    u2.username as user2_username,
    u2.profile_picture as user2_profile_picture
FROM conversations c
LEFT JOIN messages m ON c.last_message_id = m.id
LEFT JOIN users u1 ON c.user1_id = u1.id
LEFT JOIN users u2 ON c.user2_id = u2.id;

-- Vue pour les messages avec les informations de l'expéditeur
CREATE OR REPLACE VIEW message_details AS
SELECT 
    m.*,
    u.username as sender_username,
    u.profile_picture as sender_profile_picture,
    t.status as transaction_status,
    t.completed_at as transaction_completed_at
FROM messages m
LEFT JOIN users u ON m.sender_id = u.id
LEFT JOIN paid_message_transactions t ON m.id = t.message_id;

-- Commentaire de fin de migration
COMMENT ON SCHEMA public IS 'OnlyFlick Database - Migration 002 completed successfully';
