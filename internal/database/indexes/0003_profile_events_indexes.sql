-- Indexes untuk profil, kategori favorit, dan pencarian event terdekat.
CREATE INDEX IF NOT EXISTS idx_user_profiles_city_id ON user_profiles(city_id);
CREATE INDEX IF NOT EXISTS idx_events_city_id ON events(city_id);
CREATE INDEX IF NOT EXISTS idx_events_province_id ON events(province_id);
CREATE INDEX IF NOT EXISTS idx_user_favorite_categories_category_id ON user_favorite_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_events_status_end_date ON events(status, end_date);
