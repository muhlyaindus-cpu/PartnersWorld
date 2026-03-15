package user

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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
func (r *Repository) CreateUser(email, password, role, country, companyName, description string) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	var id int64
	err = r.db.QueryRow(`
		INSERT INTO users (email, password_hash, role, country, company_name, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		email, string(hashedPassword), role, country, companyName, description,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	log.Printf("User created with ID: %d", id)
	return id, nil
}

// CheckPassword проверяет пароль для указанного email.
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

// GetUserByEmail возвращает пользователя по email.
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var u User
	err := r.db.QueryRow(`
		SELECT
			id, email, password_hash, role, country, company_name, description,
			avatar_url, rating, services, portfolio, hourly_rate, project_rate, contact_info,
			created_at, updated_at
		FROM users
		WHERE email = $1`, email).
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

// GetUserByID возвращает пользователя по ID.
func (r *Repository) GetUserByID(id int64) (*User, error) {
	var u User
	err := r.db.QueryRow(`
		SELECT
			id, email, password_hash, role, country, company_name, description,
			avatar_url, rating, services, portfolio, hourly_rate, project_rate, contact_info,
			created_at, updated_at
		FROM users
		WHERE id = $1`, id).
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
func (r *Repository) UpdateUser(userID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	idx := 1
	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, idx))
		args = append(args, value)
		idx++
	}
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(setClauses, ", "), idx)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// GetUserCount возвращает общее количество пользователей в системе.
func (r *Repository) GetUserCount() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}
	return count, nil
}
