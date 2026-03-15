package partnership

import "time"

// RequestStatus описывает статус запроса
type RequestStatus string

const (
	StatusOpen   RequestStatus = "open"
	StatusClosed RequestStatus = "closed"
)

// PartnershipRequest представляет запрос экспортёра на поиск партнёра
type PartnershipRequest struct {
	ID          int64         `json:"id" db:"id"`
	UserID      int64         `json:"user_id" db:"user_id"`
	Type        string        `json:"type" db:"type"` // "need" или "offer"
	Title       string        `json:"title" db:"title"`
	Description string        `json:"description" db:"description"`
	Country     string        `json:"country" db:"country"`
	Category    string        `json:"category" db:"category"`
	Budget      *float64      `json:"budget,omitempty" db:"budget"`
	Status      RequestStatus `json:"status" db:"status"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// CreateRequestRequest — тело запроса на создание
type CreateRequestRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Country     string   `json:"country"`
	Category    string   `json:"category,omitempty"`
	Budget      *float64 `json:"budget,omitempty"`
}

// PartnershipResponse представляет отклик партнёра на запрос.
type PartnershipResponse struct {
	ID              int64     `json:"id" db:"id"`
	RequestID       int64     `json:"request_id" db:"request_id"`
	ResponderUserID int64     `json:"responder_user_id" db:"responder_user_id"`
	Message         string    `json:"message" db:"message"`
	Terms           *string   `json:"terms,omitempty" db:"terms"` // произвольные условия (текст/JSON)
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// RespondToRequestRequest — тело запроса отклика.
type RespondToRequestRequest struct {
	Message string  `json:"message"`
	Terms   *string `json:"terms,omitempty"`
}
