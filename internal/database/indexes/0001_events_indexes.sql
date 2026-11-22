-- Indexes untuk performa query events
CREATE INDEX IF NOT EXISTS idx_events_organizer_id ON events(organizer_id);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_city ON events(city);
CREATE INDEX IF NOT EXISTS idx_events_start_date ON events(start_date);
CREATE INDEX IF NOT EXISTS idx_bird_categories_slug ON bird_categories(slug);
CREATE INDEX IF NOT EXISTS idx_sub_events_event_id ON sub_events(event_id);
CREATE INDEX IF NOT EXISTS idx_sub_events_status ON sub_events(status);
CREATE INDEX IF NOT EXISTS idx_event_tickets_event_id ON event_tickets(event_id);
CREATE INDEX IF NOT EXISTS idx_event_tickets_user_id ON event_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_event_tickets_status ON event_tickets(status);
CREATE INDEX IF NOT EXISTS idx_event_tickets_qr_code ON event_tickets(qr_code);
CREATE INDEX IF NOT EXISTS idx_event_organizer_users_organizer_id ON event_organizer_users(organizer_id);
CREATE INDEX IF NOT EXISTS idx_event_organizer_users_user_id ON event_organizer_users(user_id);