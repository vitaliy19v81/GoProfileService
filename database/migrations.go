package database

import (
	"database/sql"
	"fmt"
)

// RunMigrations выполняет все миграции.
func RunMigrations(db *sql.DB) error {
	if err := CreateProfilesTable(db); err != nil {
		return err
	}

	// Если нужно, можно добавить вызовы других миграций здесь.
	return nil
}

// CreateProfilesTable Создание таблицы для профилей пользователей
func CreateProfilesTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS profiles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL, -- Связь с таблицей пользователей из сервиса регистрации
		first_name VARCHAR(255) NOT NULL, -- Имя пользователя
		last_name VARCHAR(255) NOT NULL, -- Фамилия
		middle_name VARCHAR(255), -- Отчество (опционально)
		phone VARCHAR(20), -- Номер телефона
		address TEXT, -- Адрес
		birthday DATE, -- Дата рождения
		additional_data JSONB, -- Дополнительные данные
		created_at TIMESTAMPTZ DEFAULT now(), -- Дата создания
		updated_at TIMESTAMPTZ DEFAULT now(), -- Дата последнего обновления
		CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
	CREATE INDEX IF NOT EXISTS idx_profiles_phone ON profiles(phone);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create profiles table: %w", err)
	}
	return nil
}

// CreateTestProfile Пример создания тестового профиля
func CreateTestProfile(db *sql.DB, userID, fullName, avatarURL string) error {
	query := `
	INSERT INTO profiles (user_id, full_name, avatar_url)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id) DO NOTHING;`
	_, err := db.Exec(query, userID, fullName, avatarURL)
	if err != nil {
		return fmt.Errorf("failed to create test profile: %w", err)
	}
	return nil
}
