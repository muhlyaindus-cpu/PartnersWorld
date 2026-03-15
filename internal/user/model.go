package user

import "time"

// User представляет продавца или партнёра в системе.
type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	Country      string    `json:"country" db:"country"`
	CompanyName  *string   `json:"company_name,omitempty" db:"company_name"`
	Description  *string   `json:"description,omitempty" db:"description"`
	AvatarURL    *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Rating       float64   `json:"rating" db:"rating"`
	Services     *string   `json:"services,omitempty" db:"services"`
	Portfolio    *string   `json:"portfolio,omitempty" db:"portfolio"`
	HourlyRate   *float64  `json:"hourly_rate,omitempty" db:"hourly_rate"`
	ProjectRate  *float64  `json:"project_rate,omitempty" db:"project_rate"`
	ContactInfo  *string   `json:"contact_info,omitempty" db:"contact_info"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest — структура для тела запроса регистрации.
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"` // "exporter" или "partner"
	Country     string `json:"country"`
	CompanyName string `json:"company_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// LoginRequest — структура для тела запроса логина.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateProfileRequest — структура для обновления профиля (PATCH/PUT)
// Все поля опциональны (указатели), чтобы различать неуказанные и пустые значения.
type UpdateProfileRequest struct {
	CompanyName *string  `json:"company_name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Services    *string  `json:"services,omitempty"`
	Portfolio   *string  `json:"portfolio,omitempty"`
	HourlyRate  *float64 `json:"hourly_rate,omitempty"`
	ProjectRate *float64 `json:"project_rate,omitempty"`
	ContactInfo *string  `json:"contact_info,omitempty"`
	// AvatarURL можно будет добавить позже
}
