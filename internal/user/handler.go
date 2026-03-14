package user

import (
	"encoding/json"
	"log"
	"net/http"
	"partnersale/internal/auth"
	"partnersale/internal/httpx"
	"strconv"
	"strings"
)

// RegisterRequest — тело запроса регистрации (уже должно быть в model.go)
// LoginRequest — тело запроса логина (уже должно быть в model.go)
// UpdateProfileRequest — тело запроса обновления профиля (уже должно быть в model.go)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Register обрабатывает POST /api/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Role == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Email, password and role are required")
		return
	}
	if req.Role != "exporter" && req.Role != "partner" {
		httpx.WriteError(w, http.StatusBadRequest, "Role must be 'exporter' or 'partner'")
		return
	}

	existing, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Error checking existing user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if existing != nil {
		httpx.WriteError(w, http.StatusConflict, "Email already registered")
		return
	}

	id, err := h.repo.CreateUser(req.Email, req.Password, req.Role, req.Country, req.CompanyName, req.Description)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "User registered successfully",
	})
}

// Login обрабатывает POST /api/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Email and password required")
		return
	}

	ok, user, err := h.repo.CheckPassword(req.Email, req.Password)
	if err != nil {
		log.Printf("Error checking password: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if !ok || user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":      user.ID,
			"email":   user.Email,
			"role":    user.Role,
			"country": user.Country,
			"company": func() string {
				if user.CompanyName != nil {
					return *user.CompanyName
				}
				return ""
			}(),
		},
	})
}

// GetMyProfile возвращает профиль текущего пользователя.
func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if user == nil {
		httpx.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile обновляет профиль текущего пользователя.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if req.CompanyName != nil {
		updates["company_name"] = *req.CompanyName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Services != nil {
		updates["services"] = *req.Services
	}
	if req.Portfolio != nil {
		updates["portfolio"] = *req.Portfolio
	}
	if req.HourlyRate != nil {
		updates["hourly_rate"] = *req.HourlyRate
	}
	if req.ProjectRate != nil {
		updates["project_rate"] = *req.ProjectRate
	}
	if req.ContactInfo != nil {
		updates["contact_info"] = *req.ContactInfo
	}

	if len(updates) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	if err := h.repo.UpdateUser(claims.UserID, updates); err != nil {
		log.Printf("Error updating user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("Error getting updated user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetPublicProfile возвращает публичный профиль пользователя по ID.
func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем ID из URL (ожидаем /api/user/{id})
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/api/user/")
	if idStr == "" || idStr == path {
		httpx.WriteError(w, http.StatusBadRequest, "User ID required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.repo.GetUserByID(id)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if user == nil {
		httpx.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	// Публичный профиль (все поля, которые могут быть NULL, объявлены как указатели)
	publicProfile := struct {
		ID          int64    `json:"id"`
		Role        string   `json:"role"`
		Country     string   `json:"country"`
		CompanyName *string  `json:"company_name,omitempty"`
		Description *string  `json:"description,omitempty"`
		Services    *string  `json:"services,omitempty"`
		Portfolio   *string  `json:"portfolio,omitempty"`
		HourlyRate  *float64 `json:"hourly_rate,omitempty"`
		ProjectRate *float64 `json:"project_rate,omitempty"`
		ContactInfo *string  `json:"contact_info,omitempty"`
		AvatarURL   *string  `json:"avatar_url,omitempty"`
		Rating      float64  `json:"rating"`
	}{
		ID:          user.ID,
		Role:        user.Role,
		Country:     user.Country,
		CompanyName: user.CompanyName,
		Description: user.Description,
		Services:    user.Services,
		Portfolio:   user.Portfolio,
		HourlyRate:  user.HourlyRate,
		ProjectRate: user.ProjectRate,
		ContactInfo: user.ContactInfo,
		AvatarURL:   user.AvatarURL,
		Rating:      user.Rating,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicProfile)
}

// GetUserCount возвращает общее количество зарегистрированных пользователей.
func (h *Handler) GetUserCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var count int64
	err := h.repo.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Error getting user count: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get user count")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]int64{"count": count})
}
