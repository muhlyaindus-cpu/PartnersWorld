package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDB подключается к PostgreSQL, используя строку из переменной окружения DATABASE_URL
// или значение по умолчанию для локального Docker.
func InitDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://partners_user:mysecretpassword@localhost:5432/partners_db?sslmode=disable"
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Не удалось выполнить ping БД:", err)
	}
	log.Println("База данных PostgreSQL подключена успешно")
	return db
}

// RunMigrations проверяет наличие таблиц и, если их нет, создаёт их.
// Поскольку ты уже создал таблицы вручную, эта функция может просто удостовериться, что таблицы существуют.
// Если таблицы отсутствуют, будут выполнены миграции (на случай чистой установки).
func RunMigrations(db *sql.DB) {
	// Проверяем, есть ли уже таблица users
	var tableCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public'`).Scan(&tableCount)
	if err != nil {
		log.Fatal("Ошибка проверки таблиц:", err)
	}
	if tableCount > 0 {
		log.Println("Таблицы уже существуют, миграции не требуются")
		return
	}

	log.Println("Таблицы не найдены, выполняем миграции...")

	// Миграция 1: создаём таблицу users
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('exporter', 'partner')),
		country TEXT,
		company_name TEXT,
		description TEXT,
		avatar_url TEXT,
		rating REAL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createUsersTable); err != nil {
		log.Fatal("Не удалось создать таблицу users:", err)
	}

	// Миграция 2: создаём таблицу partnership_requests
	createRequestsTable := `
	CREATE TABLE IF NOT EXISTS partnership_requests (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		country TEXT NOT NULL,
		category TEXT,
		budget REAL,
		status TEXT NOT NULL DEFAULT 'open',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createRequestsTable); err != nil {
		log.Fatal("Не удалось создать таблицу partnership_requests:", err)
	}

	// Миграция 3: добавляем поле type (если не добавлено)
	addTypeColumn := `ALTER TABLE partnership_requests ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'need';`
	if _, err := db.Exec(addTypeColumn); err != nil {
		log.Printf("Note: adding type column (might already exist): %v", err)
	}

	// Миграция 4: добавляем поля профиля в users (по одному)
	alterQueries := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS services TEXT;",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS portfolio TEXT;",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS hourly_rate REAL;",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS project_rate REAL;",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS contact_info TEXT;",
	}
	for _, query := range alterQueries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Note: adding profile field (might already exist): %v", err)
		}
	}

	// Миграция 5: таблица откликов на запросы
	createResponsesTable := `
	CREATE TABLE IF NOT EXISTS partnership_responses (
		id SERIAL PRIMARY KEY,
		request_id INTEGER NOT NULL REFERENCES partnership_requests(id) ON DELETE CASCADE,
		responder_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		message TEXT NOT NULL,
		terms TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (request_id, responder_user_id)
	);`
	if _, err := db.Exec(createResponsesTable); err != nil {
		log.Fatal("Не удалось создать таблицу partnership_responses:", err)
	}
	// Индексы для быстрых выборок
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_partnership_responses_request_id ON partnership_responses(request_id);`); err != nil {
		log.Printf("Note: creating responses index (request_id): %v", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_partnership_responses_responder_user_id ON partnership_responses(responder_user_id);`); err != nil {
		log.Printf("Note: creating responses index (responder_user_id): %v", err)
	}

	log.Println("Миграции применены успешно")
}
