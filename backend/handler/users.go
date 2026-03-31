package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/jaykay/vereinstool/backend/config"
	"github.com/jaykay/vereinstool/backend/db/generated"
	"github.com/jaykay/vereinstool/backend/service"
)

type UsersHandler struct {
	queries *generated.Queries
	auth    *service.Auth
	mailer  *service.Mailer
	cfg     *config.Config
}

func NewUsers(queries *generated.Queries, auth *service.Auth, mailer *service.Mailer, cfg *config.Config) *UsersHandler {
	return &UsersHandler{queries: queries, auth: auth, mailer: mailer, cfg: cfg}
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.queries.ListUsers(r.Context())
	if err != nil {
		jsonError(w, "Fehler beim Laden der Benutzer", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, len(users))
	for i, u := range users {
		result[i] = map[string]any{
			"id":        u.ID,
			"email":     u.Email,
			"name":      u.Name,
			"role":      u.Role,
			"is_active": u.IsActive == 1,
		}
	}
	jsonOK(w, result)
}

func (h *UsersHandler) Invite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" {
		jsonError(w, "E-Mail und Name erforderlich", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" {
		jsonError(w, "Ungültige Rolle", http.StatusBadRequest)
		return
	}

	// Generate temporary password
	tempPwBytes := make([]byte, 12)
	if _, err := rand.Read(tempPwBytes); err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}
	tempPassword := hex.EncodeToString(tempPwBytes)

	hash, err := h.auth.HashPassword(tempPassword)
	if err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	user, err := h.queries.CreateUser(r.Context(), generated.CreateUserParams{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hash,
		Role:         req.Role,
	})
	if err != nil {
		jsonError(w, "Benutzer konnte nicht erstellt werden (E-Mail bereits vergeben?)", http.StatusConflict)
		return
	}

	// Send invitation email (non-blocking: don't fail the request if email fails)
	go h.mailer.SendInvitation(req.Email, req.Name, tempPassword)

	jsonOK(w, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "Ungültige Benutzer-ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     *string `json:"name"`
		Role     *string `json:"role"`
		IsActive *bool   `json:"is_active"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	// Load current user
	existing, err := h.queries.GetUserByID(r.Context(), id)
	if err != nil {
		jsonError(w, "Benutzer nicht gefunden", http.StatusNotFound)
		return
	}

	name := existing.Name
	role := existing.Role
	isActive := existing.IsActive

	if req.Name != nil {
		name = *req.Name
	}
	if req.Role != nil {
		if *req.Role != "admin" && *req.Role != "member" {
			jsonError(w, "Ungültige Rolle", http.StatusBadRequest)
			return
		}
		role = *req.Role
	}
	if req.IsActive != nil {
		if *req.IsActive {
			isActive = 1
		} else {
			isActive = 0
		}
	}

	err = h.queries.UpdateUser(r.Context(), generated.UpdateUserParams{
		Name:     name,
		Role:     role,
		IsActive: isActive,
		ID:       id,
	})
	if err != nil {
		jsonError(w, "Fehler beim Aktualisieren", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]any{
		"id":        id,
		"email":     existing.Email,
		"name":      name,
		"role":      role,
		"is_active": isActive == 1,
	})
}
