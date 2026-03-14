package user

import (
	"database/sql"
	"fmt"
	"log"
	"strings" // добавить, если нет

	"golang.org/x/crypto/bcrypt"
)

// Repository предоставляет методы для работы с пользователями в БД.
type Repository struct {
	db *sql.DB
}

// NewRepository создаёт новый репозиторий.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser создаёт нового пользователя в базе.

// CheckPassword проверяет пароль для указанного email
func (r *Repository) CheckPassword(email, password string) (bool, *User, error) {
	user, err := r.GetUserByEmail(email)
	if err != nil {
		return false, nil, err
	}
	if user == nil {
		return false, nil, nil // пользователь не найден
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, user, nil // пароль не совпадает
	}
	return true, user, nil
}

// Возвращает ID созданного пользователя или ошибку.
func (r *Repository) CreateUser(email, password, role, country, companyName, description string) (int64, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	// Вставка в БД
	result, err := r.db.Exec(`
        INSERT INTO users (email, password_hash, role, country, company_name, description)
        VALUES (?, ?, ?, ?, ?, ?)`,
		email, string(hashedPassword), role, country, companyName, description,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.Printf("User created with ID: %d", id)
	return id, nil

}

// GetUserByEmail возвращает пользователя по email (понадобится позже для логина).
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var u User
	err := r.db.QueryRow(`
        SELECT 
            id, email, password_hash, role, country, company_name, description, 
            avatar_url, rating, services, portfolio, hourly_rate, project_rate, contact_info,
            created_at, updated_at
        FROM users WHERE email = ?`, email).
		Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Country, &u.CompanyName, &u.Description,
			&u.AvatarURL, &u.Rating, &u.Services, &u.Portfolio, &u.HourlyRate, &u.ProjectRate, &u.ContactInfo,
			&u.CreatedAt, &u.UpdatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &u, nil
}

func (r *Repository) GetUserByID(id int64) (*User, error) {
	var u User
	err := r.db.QueryRow(`
        SELECT 
            id, email, password_hash, role, country, company_name, description, 
            avatar_url, rating, services, portfolio, hourly_rate, project_rate, contact_info,
            created_at, updated_at
        FROM users WHERE id = ?`, id).
		Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Country, &u.CompanyName, &u.Description,
			&u.AvatarURL, &u.Rating, &u.Services, &u.Portfolio, &u.HourlyRate, &u.ProjectRate, &u.ContactInfo,
			&u.CreatedAt, &u.UpdatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &u, nil

}

// UpdateUser обновляет поля пользователя (частичное обновление).
// Принимает ID пользователя и map с именами полей и новыми значениями.
func (r *Repository) UpdateUser(userID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Строим SET часть запроса
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s, updated_at = datetime('now') WHERE id = ?", strings.Join(setClauses, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
