package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jaykay/vereinstool/backend/db/generated"
)

type DecisionsHandler struct {
	queries *generated.Queries
}

func NewDecisions(queries *generated.Queries) *DecisionsHandler {
	return &DecisionsHandler{queries: queries}
}

func (h *DecisionsHandler) List(w http.ResponseWriter, r *http.Request) {
	meetingIDStr := r.URL.Query().Get("meeting_id")

	var decisions []generated.Decision
	var err error

	if meetingIDStr != "" {
		var meetingID int64
		if _, e := parseIntParam(meetingIDStr); e != nil {
			jsonError(w, "Ungültige meeting_id", http.StatusBadRequest)
			return
		}
		meetingID, _ = parseIntParam(meetingIDStr)
		decisions, err = h.queries.ListDecisionsByMeeting(r.Context(), meetingID)
	} else {
		decisions, err = h.queries.ListDecisions(r.Context())
	}
	if err != nil {
		jsonError(w, "Fehler beim Laden der Beschlüsse", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, len(decisions))
	for i, d := range decisions {
		result[i] = decisionResponse(&d)
	}
	jsonOK(w, result)
}

func (h *DecisionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TopicID      int64  `json:"topic_id"`
		MeetingID    int64  `json:"meeting_id"`
		Text         string `json:"text"`
		VotesYes     *int64 `json:"votes_yes"`
		VotesNo      *int64 `json:"votes_no"`
		VotesAbstain *int64 `json:"votes_abstain"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Text == "" || req.MeetingID == 0 || req.TopicID == 0 {
		jsonError(w, "Text, Sitzung und Thema erforderlich", http.StatusBadRequest)
		return
	}

	user := UserFromContext(r.Context())
	decision, err := h.queries.CreateDecision(r.Context(), generated.CreateDecisionParams{
		TopicID:      req.TopicID,
		MeetingID:    req.MeetingID,
		Text:         req.Text,
		VotesYes:     nullInt64(req.VotesYes),
		VotesNo:      nullInt64(req.VotesNo),
		VotesAbstain: nullInt64(req.VotesAbstain),
		RecordedBy:   user.ID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Erstellen", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, decisionResponse(&decision))
}

func (h *DecisionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	existing, err := h.queries.GetDecision(r.Context(), id)
	if err != nil {
		jsonError(w, "Beschluss nicht gefunden", http.StatusNotFound)
		return
	}

	var req struct {
		Text         *string `json:"text"`
		VotesYes     *int64  `json:"votes_yes"`
		VotesNo      *int64  `json:"votes_no"`
		VotesAbstain *int64  `json:"votes_abstain"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	text := existing.Text
	votesYes := existing.VotesYes
	votesNo := existing.VotesNo
	votesAbstain := existing.VotesAbstain

	if req.Text != nil {
		text = *req.Text
	}
	if req.VotesYes != nil {
		votesYes = sql.NullInt64{Int64: *req.VotesYes, Valid: true}
	}
	if req.VotesNo != nil {
		votesNo = sql.NullInt64{Int64: *req.VotesNo, Valid: true}
	}
	if req.VotesAbstain != nil {
		votesAbstain = sql.NullInt64{Int64: *req.VotesAbstain, Valid: true}
	}

	err = h.queries.UpdateDecision(r.Context(), generated.UpdateDecisionParams{
		Text:         text,
		VotesYes:     votesYes,
		VotesNo:      votesNo,
		VotesAbstain: votesAbstain,
		ID:           id,
	})
	if err != nil {
		jsonError(w, "Fehler beim Aktualisieren", http.StatusInternalServerError)
		return
	}

	updated, _ := h.queries.GetDecision(r.Context(), id)
	jsonOK(w, decisionResponse(&updated))
}

func decisionResponse(d *generated.Decision) map[string]any {
	resp := map[string]any{
		"id":         d.ID,
		"topic_id":   d.TopicID,
		"meeting_id": d.MeetingID,
		"text":       d.Text,
		"recorded_by": d.RecordedBy,
		"created_at": d.CreatedAt.Format(time.RFC3339),
	}
	if d.VotesYes.Valid {
		resp["votes_yes"] = d.VotesYes.Int64
	}
	if d.VotesNo.Valid {
		resp["votes_no"] = d.VotesNo.Int64
	}
	if d.VotesAbstain.Valid {
		resp["votes_abstain"] = d.VotesAbstain.Int64
	}
	return resp
}

func nullInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func parseIntParam(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscan(s, &v)
	return v, err
}
