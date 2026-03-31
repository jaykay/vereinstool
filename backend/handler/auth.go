package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/jaykay/vereinstool/backend/config"
	"github.com/jaykay/vereinstool/backend/service"
)

type AuthHandler struct {
	auth   *service.Auth
	mailer *service.Mailer
	cfg    *config.Config
}

func NewAuth(auth *service.Auth, mailer *service.Mailer, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, mailer: mailer, cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonError(w, "E-Mail und Passwort erforderlich", http.StatusBadRequest)
		return
	}

	sessionID, user, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserInactive) {
			jsonError(w, "E-Mail oder Passwort falsch", http.StatusUnauthorized)
			return
		}
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(30 * 24 * time.Hour / time.Second),
	})

	jsonOK(w, map[string]any{
		"user": userResponse(user.ID, user.Email, user.Name, user.Role),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		_ = h.auth.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	jsonOK(w, map[string]string{"status": "ok"})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := readJSON(r, &req); err != nil || req.Email == "" {
		jsonError(w, "E-Mail erforderlich", http.StatusBadRequest)
		return
	}

	token, err := h.auth.CreatePasswordResetToken(r.Context(), req.Email)
	if err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	// Send email if token was created (user exists)
	if token != "" {
		_ = h.mailer.SendPasswordReset(req.Email, token)
	}

	// Always return success to prevent email enumeration
	jsonOK(w, map[string]string{"status": "ok"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.Password == "" {
		jsonError(w, "Token und Passwort erforderlich", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		jsonError(w, "Passwort muss mindestens 8 Zeichen lang sein", http.StatusBadRequest)
		return
	}

	err := h.auth.ResetPassword(r.Context(), req.Token, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidResetToken) {
			jsonError(w, "Ungültiger oder abgelaufener Reset-Link", http.StatusBadRequest)
			return
		}
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		jsonError(w, "Nicht angemeldet", http.StatusUnauthorized)
		return
	}
	jsonOK(w, map[string]any{
		"user": userResponse(user.ID, user.Email, user.Name, user.Role),
	})
}

func userResponse(id int64, email, name, role string) map[string]any {
	return map[string]any{
		"id":    id,
		"email": email,
		"name":  name,
		"role":  role,
	}
}
