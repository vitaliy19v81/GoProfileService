package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"profile_service/config"
)

// InitDB Подключение к базе данных
func InitDB() (*sql.DB, error) {
	//dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	//	os.Getenv("DB_HOST"),
	//	os.Getenv("DB_PORT"),
	//	os.Getenv("DB_USER"),
	//	os.Getenv("DB_PASSWORD"),
	//	os.Getenv("DB_NAME"),
	//)
	dsn := config.DsnPostgres // os.Getenv("DSN_POSTGRES")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}
