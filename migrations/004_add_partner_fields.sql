-- Добавляем поля для профиля партнёра (и для всех пользователей)
ALTER TABLE users ADD COLUMN services TEXT;
ALTER TABLE users ADD COLUMN portfolio TEXT;
ALTER TABLE users ADD COLUMN hourly_rate REAL;
ALTER TABLE users ADD COLUMN project_rate REAL;
ALTER TABLE users ADD COLUMN contact_info TEXT;