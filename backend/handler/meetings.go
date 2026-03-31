package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/jaykay/vereinstool/backend/db/generated"
)

type MeetingsHandler struct {
	queries *generated.Queries
}

func NewMeetings(queries *generated.Queries) *MeetingsHandler {
	return &MeetingsHandler{queries: queries}
}

func (h *MeetingsHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var meetings []generated.Meeting
	var err error
	if status != "" {
		meetings, err = h.queries.ListMeetingsByStatus(r.Context(), status)
	} else {
		meetings, err = h.queries.ListMeetings(r.Context())
	}
	if err != nil {
		jsonError(w, "Fehler beim Laden der Sitzungen", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, len(meetings))
	for i, m := range meetings {
		result[i] = meetingResponse(&m)
	}
	jsonOK(w, result)
}

func (h *MeetingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	meeting, err := h.queries.GetMeeting(r.Context(), id)
	if err != nil {
		jsonError(w, "Sitzung nicht gefunden", http.StatusNotFound)
		return
	}

	attendees, _ := h.queries.ListAttendees(r.Context(), id)

	jsonOK(w, map[string]any{
		"meeting":   meetingResponse(&meeting),
		"attendees": attendeeList(attendees),
	})
}

func (h *MeetingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title        string `json:"title"`
		ScheduledAt  string `json:"scheduled_at"`
		DurationMins int64  `json:"duration_mins"`
		Location     string `json:"location"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.ScheduledAt == "" {
		jsonError(w, "Titel und Datum erforderlich", http.StatusBadRequest)
		return
	}

	if req.DurationMins <= 0 {
		req.DurationMins = 90
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		jsonError(w, "Ungültiges Datumsformat (RFC3339 erwartet)", http.StatusBadRequest)
		return
	}

	user := UserFromContext(r.Context())
	meeting, err := h.queries.CreateMeeting(r.Context(), generated.CreateMeetingParams{
		Title:        req.Title,
		ScheduledAt:  scheduledAt,
		DurationMins: req.DurationMins,
		Location:     sql.NullString{String: req.Location, Valid: req.Location != ""},
		CreatedBy:    user.ID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Erstellen", http.StatusInternalServerError)
		return
	}

	// Creator is automatically an attendee
	_ = h.queries.AddAttendee(r.Context(), generated.AddAttendeeParams{
		MeetingID: meeting.ID,
		UserID:    user.ID,
	})

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, meetingResponse(&meeting))
}

func (h *MeetingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	existing, err := h.queries.GetMeeting(r.Context(), id)
	if err != nil {
		jsonError(w, "Sitzung nicht gefunden", http.StatusNotFound)
		return
	}

	if existing.Status != "open" {
		jsonError(w, "Nur offene Sitzungen können bearbeitet werden", http.StatusConflict)
		return
	}

	var req struct {
		Title        *string `json:"title"`
		ScheduledAt  *string `json:"scheduled_at"`
		DurationMins *int64  `json:"duration_mins"`
		Location     *string `json:"location"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	title := existing.Title
	scheduledAt := existing.ScheduledAt
	durationMins := existing.DurationMins
	location := existing.Location

	if req.Title != nil {
		title = *req.Title
	}
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			jsonError(w, "Ungültiges Datumsformat", http.StatusBadRequest)
			return
		}
		scheduledAt = t
	}
	if req.DurationMins != nil {
		durationMins = *req.DurationMins
	}
	if req.Location != nil {
		location = sql.NullString{String: *req.Location, Valid: *req.Location != ""}
	}

	err = h.queries.UpdateMeeting(r.Context(), generated.UpdateMeetingParams{
		Title:        title,
		ScheduledAt:  scheduledAt,
		DurationMins: durationMins,
		Location:     location,
		ID:           id,
	})
	if err != nil {
		jsonError(w, "Fehler beim Aktualisieren", http.StatusInternalServerError)
		return
	}

	updated, _ := h.queries.GetMeeting(r.Context(), id)
	jsonOK(w, meetingResponse(&updated))
}

func (h *MeetingsHandler) Start(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, "open", "active")
}

func (h *MeetingsHandler) Close(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, "active", "closed")
}

func (h *MeetingsHandler) changeStatus(w http.ResponseWriter, r *http.Request, requiredStatus, newStatus string) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	meeting, err := h.queries.GetMeeting(r.Context(), id)
	if err != nil {
		jsonError(w, "Sitzung nicht gefunden", http.StatusNotFound)
		return
	}

	if meeting.Status != requiredStatus {
		jsonError(w, "Sitzung hat nicht den erwarteten Status", http.StatusConflict)
		return
	}

	err = h.queries.UpdateMeetingStatus(r.Context(), generated.UpdateMeetingStatusParams{
		Status: newStatus,
		ID:     id,
	})
	if err != nil {
		jsonError(w, "Fehler beim Statuswechsel", http.StatusInternalServerError)
		return
	}

	meeting.Status = newStatus
	jsonOK(w, meetingResponse(&meeting))
}

// Attendee management

func (h *MeetingsHandler) AddAttendee(w http.ResponseWriter, r *http.Request) {
	meetingID, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := readJSON(r, &req); err != nil || req.UserID == 0 {
		jsonError(w, "user_id erforderlich", http.StatusBadRequest)
		return
	}

	err = h.queries.AddAttendee(r.Context(), generated.AddAttendeeParams{
		MeetingID: meetingID,
		UserID:    req.UserID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Hinzufügen", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

func (h *MeetingsHandler) RemoveAttendee(w http.ResponseWriter, r *http.Request) {
	meetingID, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	userID, err := paramInt64(r, "userId")
	if err != nil {
		jsonError(w, "Ungültige User-ID", http.StatusBadRequest)
		return
	}

	err = h.queries.RemoveAttendee(r.Context(), generated.RemoveAttendeeParams{
		MeetingID: meetingID,
		UserID:    userID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Entfernen", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

// Helpers

func paramInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func meetingResponse(m *generated.Meeting) map[string]any {
	resp := map[string]any{
		"id":            m.ID,
		"title":         m.Title,
		"scheduled_at":  m.ScheduledAt.Format(time.RFC3339),
		"duration_mins": m.DurationMins,
		"status":        m.Status,
		"created_by":    m.CreatedBy,
		"created_at":    m.CreatedAt.Format(time.RFC3339),
	}
	if m.Location.Valid {
		resp["location"] = m.Location.String
	} else {
		resp["location"] = nil
	}
	return resp
}

func attendeeList(attendees []generated.ListAttendeesRow) []map[string]any {
	result := make([]map[string]any, len(attendees))
	for i, a := range attendees {
		result[i] = map[string]any{
			"id":      a.ID,
			"email":   a.Email,
			"name":    a.Name,
			"role":    a.Role,
			"present": a.Present == 1,
		}
	}
	return result
}
