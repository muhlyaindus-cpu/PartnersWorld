package partnership

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create добавляет новый запрос в БД
func (r *Repository) Create(req *PartnershipRequest) (int64, error) {
	query := `
        INSERT INTO partnership_requests (user_id, type, title, description, country, category, budget, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id
    `
	var id int64
	err := r.db.QueryRow(query,
		req.UserID, req.Type, req.Title, req.Description, req.Country,
		req.Category, req.Budget, req.Status).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert request: %w", err)
	}
	return id, nil
}

// List возвращает все открытые запросы с возможностью фильтрации по стране
func (r *Repository) List(countryFilter string) ([]PartnershipRequest, error) {
	var rows *sql.Rows
	var err error

	baseQuery := `
        SELECT id, user_id, type, title, description, country, category, budget, status, created_at, updated_at
        FROM partnership_requests
        WHERE status = 'open'
    `

	if countryFilter != "" {
		rows, err = r.db.Query(baseQuery+` AND country = $1 ORDER BY created_at DESC`, countryFilter)
	} else {
		rows, err = r.db.Query(baseQuery + ` ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	var requests []PartnershipRequest
	for rows.Next() {
		var r PartnershipRequest
		err := rows.Scan(
			&r.ID, &r.UserID, &r.Type, &r.Title, &r.Description, &r.Country,
			&r.Category, &r.Budget, &r.Status, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		requests = append(requests, r)
	}
	return requests, nil
}

// GetByID возвращает запрос по ID (может понадобиться для отклика)
func (r *Repository) GetByID(id int64) (*PartnershipRequest, error) {
	var req PartnershipRequest
	err := r.db.QueryRow(`
        SELECT id, user_id, type, title, description, country, category, budget, status, created_at, updated_at
        FROM partnership_requests WHERE id = $1
    `, id).Scan(
		&req.ID, &req.UserID, &req.Type, &req.Title, &req.Description, &req.Country,
		&req.Category, &req.Budget, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get request by id: %w", err)
	}
	return &req, nil
}

// CloseRequest закрывает запрос (меняет статус на closed)
func (r *Repository) CloseRequest(id int64) error {
	_, err := r.db.Exec(`
		UPDATE partnership_requests SET status = 'closed', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("close request: %w", err)
	}
	return nil
}

// CreateOrUpdateResponse создаёт отклик партнёра на запрос или обновляет его (если уже есть).
// Возвращает id отклика и флаг created (true если создан новый).
func (r *Repository) CreateOrUpdateResponse(requestID, responderUserID int64, message string, terms *string) (id int64, created bool, err error) {
	var existingID int64
	err = r.db.QueryRow(
		`SELECT id FROM partnership_responses WHERE request_id = $1 AND responder_user_id = $2`,
		requestID,
		responderUserID,
	).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, false, fmt.Errorf("get existing response id: %w", err)
	}

	if err == sql.ErrNoRows {
		var newID int64
		scanErr := r.db.QueryRow(
			`INSERT INTO partnership_responses (request_id, responder_user_id, message, terms, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 RETURNING id`,
			requestID,
			responderUserID,
			message,
			terms,
		).Scan(&newID)
		if scanErr != nil {
			return 0, false, fmt.Errorf("insert response: %w", scanErr)
		}
		return newID, true, nil
	}

	_, execErr := r.db.Exec(
		`UPDATE partnership_responses
		 SET message = $1, terms = $2, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $3`,
		message,
		terms,
		existingID,
	)
	if execErr != nil {
		return 0, false, fmt.Errorf("update response: %w", execErr)
	}
	return existingID, false, nil
}

// ListResponsesByRequest возвращает отклики по заявке.
// Если responderUserID != nil — возвращает только отклики конкретного партнёра.
func (r *Repository) ListResponsesByRequest(requestID int64, responderUserID *int64) ([]PartnershipResponse, error) {
	base := `
		SELECT id, request_id, responder_user_id, message, terms, created_at, updated_at
		FROM partnership_responses
		WHERE request_id = $1
	`
	var (
		rows *sql.Rows
		err  error
	)
	if responderUserID != nil {
		rows, err = r.db.Query(base+` AND responder_user_id = $2 ORDER BY created_at DESC`, requestID, *responderUserID)
	} else {
		rows, err = r.db.Query(base+` ORDER BY created_at DESC`, requestID)
	}
	if err != nil {
		return nil, fmt.Errorf("list responses: %w", err)
	}
	defer rows.Close()

	var out []PartnershipResponse
	for rows.Next() {
		var rr PartnershipResponse
		if scanErr := rows.Scan(
			&rr.ID,
			&rr.RequestID,
			&rr.ResponderUserID,
			&rr.Message,
			&rr.Terms,
			&rr.CreatedAt,
			&rr.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("scan response: %w", scanErr)
		}
		out = append(out, rr)
	}
	return out, nil
}
