-- Add your db schema creation SQL for reactions here

-- Messages
CREATE TABLE IF NOT EXISTS messages (
  id uuid DEFAULT gen_random_uuid(),
  message_text TEXT NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  list_of_reactions JSONB DEFAULT '[]',
  reaction_score INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Reactions (TODO: FK ref)
CREATE TABLE IF NOT EXISTS reactions (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY, 
  message_id uuid NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  reaction_type VARCHAR(255) NOT NULL,
  reaction_score INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_reactions_message_id ON reactions (message_id);
