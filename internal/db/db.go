package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // вместо "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		log.Fatal("Не удалось подключиться к БД:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Не удалось выполнить ping БД:", err)
	}
	log.Println("База данных подключена успешно")
	return db
}

func RunMigrations(db *sql.DB) {
	// Миграция 1: создаём таблицу users
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        role TEXT NOT NULL CHECK(role IN ('exporter', 'partner')),
        country TEXT,
        company_name TEXT,
        description TEXT,
        avatar_url TEXT,
        rating REAL DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := db.Exec(createUsersTable); err != nil {
		log.Fatal("Не удалось создать таблицу users:", err)
	}

	// Миграция 2: создаём таблицу partnership_requests
	createRequestsTable := `
    CREATE TABLE IF NOT EXISTS partnership_requests (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        title TEXT NOT NULL,
        description TEXT NOT NULL,
        country TEXT NOT NULL,
        category TEXT,
        budget REAL,
        status TEXT NOT NULL DEFAULT 'open',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );`
	if _, err := db.Exec(createRequestsTable); err != nil {
		log.Fatal("Не удалось создать таблицу partnership_requests:", err)
	}

	// Миграция 3: добавляем поле type
	addTypeColumn := `ALTER TABLE partnership_requests ADD COLUMN type TEXT NOT NULL DEFAULT 'need';`
	if _, err := db.Exec(addTypeColumn); err != nil {
		// Игнорируем ошибку, если колонка уже существует
		log.Printf("Note: adding type column (might already exist): %v", err)
	}
	// Миграция 4: добавляем поля профиля по одному
	alterQueries := []string{
		"ALTER TABLE users ADD COLUMN services TEXT;",
		"ALTER TABLE users ADD COLUMN portfolio TEXT;",
		"ALTER TABLE users ADD COLUMN hourly_rate REAL;",
		"ALTER TABLE users ADD COLUMN project_rate REAL;",
		"ALTER TABLE users ADD COLUMN contact_info TEXT;",
	}
	for _, query := range alterQueries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Note: adding profile field (might already exist): %v", err)
		}
	}

	// Миграция 5: таблица откликов на запросы (responses)
	createResponsesTable := `
	CREATE TABLE IF NOT EXISTS partnership_responses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_id INTEGER NOT NULL,
		responder_user_id INTEGER NOT NULL,
		message TEXT NOT NULL,
		terms TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (request_id) REFERENCES partnership_requests(id) ON DELETE CASCADE,
		FOREIGN KEY (responder_user_id) REFERENCES users(id) ON DELETE CASCADE,
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
