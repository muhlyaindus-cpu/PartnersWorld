package partnership

import (
	"encoding/json"
	"log"
	"net/http"
	"partnersale/internal/auth"
	"partnersale/internal/httpx"
	"strconv"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest обрабатывает POST /api/requests
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем пользователя из контекста (установлен middleware)
	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Определяем тип запроса на основе роли
	var reqType string
	switch claims.Role {
	case "exporter":
		reqType = "need"
	case "partner":
		reqType = "offer"
	default:
		httpx.WriteError(w, http.StatusForbidden, "Invalid user role")
		return
	}

	// Декодируем тело запроса
	var req CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if req.Title == "" || req.Description == "" || req.Country == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Title, description and country are required")
		return
	}

	// Создаём объект запроса
	partReq := &PartnershipRequest{
		UserID:      claims.UserID,
		Type:        reqType,
		Title:       req.Title,
		Description: req.Description,
		Country:     req.Country,
		Category:    req.Category,
		Budget:      req.Budget,
		Status:      StatusOpen,
	}

	id, err := h.repo.Create(partReq)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Request created successfully",
	})
}

// ListRequests обрабатывает GET /api/requests
func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Фильтр по стране (опционально)
	country := r.URL.Query().Get("country")

	requests, err := h.repo.List(country)
	if err != nil {
		log.Printf("Error listing requests: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to list requests")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// RespondToRequest обрабатывает POST /api/requests/{id}/respond
func (h *Handler) RespondToRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Только партнёр может откликаться
	if claims.Role != "partner" {
		httpx.WriteError(w, http.StatusForbidden, "Only partners can respond to requests")
		return
	}

	// Извлекаем ID запроса из URL
	idStr := r.PathValue("id") // если используешь chi или gorilla/mux, может быть по-другому
	if idStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Request ID required")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	// Проверяем, что запрос существует и открыт
	req, err := h.repo.GetByID(id)
	if err != nil {
		log.Printf("Error getting request: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if req == nil {
		httpx.WriteError(w, http.StatusNotFound, "Request not found")
		return
	}
	if req.Status != StatusOpen {
		httpx.WriteError(w, http.StatusBadRequest, "Request is already closed")
		return
	}

	var body RespondToRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if body.Message == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Message is required")
		return
	}

	responseID, created, err := h.repo.CreateOrUpdateResponse(id, claims.UserID, body.Message, body.Terms)
	if err != nil {
		log.Printf("Error creating response: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to record response")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	httpx.WriteJSON(w, status, map[string]any{
		"id":      responseID,
		"created": created,
		"message": "Response recorded",
	})
}

// ListResponses обрабатывает GET /api/requests/{id}/responses
// Правила:
// - владелец заявки видит все отклики
// - партнёр видит только свой отклик на эту заявку
func (h *Handler) ListResponses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Request ID required")
		return
	}
	requestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	req, err := h.repo.GetByID(requestID)
	if err != nil {
		log.Printf("Error getting request: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if req == nil {
		httpx.WriteError(w, http.StatusNotFound, "Request not found")
		return
	}

	// Владелец заявки (обычно exporter) — видит все отклики.
	if claims.UserID == req.UserID {
		responses, listErr := h.repo.ListResponsesByRequest(requestID, nil)
		if listErr != nil {
			log.Printf("Error listing responses: %v", listErr)
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to list responses")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, responses)
		return
	}

	// Партнёр — видит только свой отклик.
	if claims.Role == "partner" {
		responses, listErr := h.repo.ListResponsesByRequest(requestID, &claims.UserID)
		if listErr != nil {
			log.Printf("Error listing responses: %v", listErr)
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to list responses")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, responses)
		return
	}

	httpx.WriteError(w, http.StatusForbidden, "Forbidden")
}
