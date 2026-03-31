package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/jaykay/vereinstool/backend/db/generated"
)

type TopicsHandler struct {
	queries *generated.Queries
	db      *sql.DB
}

func NewTopics(queries *generated.Queries, db *sql.DB) *TopicsHandler {
	return &TopicsHandler{queries: queries, db: db}
}

func (h *TopicsHandler) ListByMeeting(w http.ResponseWriter, r *http.Request) {
	meetingID, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	topics, err := h.queries.ListTopicsByMeeting(r.Context(), sql.NullInt64{Int64: meetingID, Valid: true})
	if err != nil {
		jsonError(w, "Fehler beim Laden der Themen", http.StatusInternalServerError)
		return
	}

	// Get current user's votes for this meeting
	user := UserFromContext(r.Context())
	votedTopicIDs, _ := h.queries.ListUserVotesForMeeting(r.Context(), generated.ListUserVotesForMeetingParams{
		MeetingID: sql.NullInt64{Int64: meetingID, Valid: true},
		UserID:    user.ID,
	})
	votedSet := make(map[int64]bool, len(votedTopicIDs))
	for _, id := range votedTopicIDs {
		votedSet[id] = true
	}

	result := make([]map[string]any, len(topics))
	for i, t := range topics {
		result[i] = topicResponse(&t, votedSet[t.ID])
	}
	jsonOK(w, result)
}

func (h *TopicsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MeetingID     int64  `json:"meeting_id"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		Category      string `json:"category"`
		EstimatedMins int64  `json:"estimated_mins"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.MeetingID == 0 {
		jsonError(w, "Titel und Sitzung erforderlich", http.StatusBadRequest)
		return
	}

	if req.EstimatedMins <= 0 {
		req.EstimatedMins = 10
	}

	user := UserFromContext(r.Context())
	topic, err := h.queries.CreateTopic(r.Context(), generated.CreateTopicParams{
		MeetingID:     sql.NullInt64{Int64: req.MeetingID, Valid: true},
		Title:         req.Title,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		Category:      sql.NullString{String: req.Category, Valid: req.Category != ""},
		SubmittedBy:   user.ID,
		EstimatedMins: req.EstimatedMins,
	})
	if err != nil {
		jsonError(w, "Fehler beim Erstellen", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, topicResponse(&topic, false))
}

func (h *TopicsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	existing, err := h.queries.GetTopic(r.Context(), id)
	if err != nil {
		jsonError(w, "Thema nicht gefunden", http.StatusNotFound)
		return
	}

	var req struct {
		Title         *string `json:"title"`
		Description   *string `json:"description"`
		Category      *string `json:"category"`
		EstimatedMins *int64  `json:"estimated_mins"`
		Status        *string `json:"status"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	title := existing.Title
	description := existing.Description
	category := existing.Category
	estimatedMins := existing.EstimatedMins
	status := existing.Status

	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: *req.Description != ""}
	}
	if req.Category != nil {
		category = sql.NullString{String: *req.Category, Valid: *req.Category != ""}
	}
	if req.EstimatedMins != nil {
		estimatedMins = *req.EstimatedMins
	}
	if req.Status != nil {
		status = *req.Status
	}

	err = h.queries.UpdateTopic(r.Context(), generated.UpdateTopicParams{
		Title:         title,
		Description:   description,
		Category:      category,
		EstimatedMins: estimatedMins,
		Status:        status,
		ID:            id,
	})
	if err != nil {
		jsonError(w, "Fehler beim Aktualisieren", http.StatusInternalServerError)
		return
	}

	updated, _ := h.queries.GetTopic(r.Context(), id)
	jsonOK(w, topicResponse(&updated, false))
}

func (h *TopicsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	// Check topic exists and user is submitter or admin
	topic, err := h.queries.GetTopic(r.Context(), id)
	if err != nil {
		jsonError(w, "Thema nicht gefunden", http.StatusNotFound)
		return
	}

	user := UserFromContext(r.Context())
	if topic.SubmittedBy != user.ID && user.Role != "admin" {
		jsonError(w, "Keine Berechtigung", http.StatusForbidden)
		return
	}

	if err := h.queries.DeleteTopic(r.Context(), id); err != nil {
		jsonError(w, "Fehler beim Löschen", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "ok"})
}

// Vote adds an upvote (with denormalized vote_count update in a transaction)
func (h *TopicsHandler) Vote(w http.ResponseWriter, r *http.Request) {
	topicID, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	user := UserFromContext(r.Context())

	// Check if already voted
	_, err = h.queries.GetVote(r.Context(), generated.GetVoteParams{
		TopicID: topicID,
		UserID:  user.ID,
	})
	if err == nil {
		jsonError(w, "Bereits abgestimmt", http.StatusConflict)
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)
	err = qtx.CreateVote(r.Context(), generated.CreateVoteParams{
		TopicID: topicID,
		UserID:  user.ID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Abstimmen", http.StatusInternalServerError)
		return
	}

	err = qtx.IncrementVoteCount(r.Context(), topicID)
	if err != nil {
		jsonError(w, "Fehler beim Abstimmen", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	topic, _ := h.queries.GetTopic(r.Context(), topicID)
	jsonOK(w, topicResponse(&topic, true))
}

// Unvote removes an upvote (with denormalized vote_count update in a transaction)
func (h *TopicsHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	topicID, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	user := UserFromContext(r.Context())

	// Check if actually voted
	_, err = h.queries.GetVote(r.Context(), generated.GetVoteParams{
		TopicID: topicID,
		UserID:  user.ID,
	})
	if err != nil {
		jsonError(w, "Noch nicht abgestimmt", http.StatusConflict)
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)
	err = qtx.DeleteVote(r.Context(), generated.DeleteVoteParams{
		TopicID: topicID,
		UserID:  user.ID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Entfernen", http.StatusInternalServerError)
		return
	}

	err = qtx.DecrementVoteCount(r.Context(), topicID)
	if err != nil {
		jsonError(w, "Fehler beim Entfernen", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, "Interner Fehler", http.StatusInternalServerError)
		return
	}

	topic, _ := h.queries.GetTopic(r.Context(), topicID)
	jsonOK(w, topicResponse(&topic, false))
}

func topicResponse(t *generated.Topic, voted bool) map[string]any {
	resp := map[string]any{
		"id":             t.ID,
		"title":          t.Title,
		"submitted_by":   t.SubmittedBy,
		"estimated_mins": t.EstimatedMins,
		"status":         t.Status,
		"vote_count":     t.VoteCount,
		"is_recurring":   t.IsRecurring == 1,
		"voted":          voted,
		"created_at":     t.CreatedAt.Format(time.RFC3339),
	}
	if t.MeetingID.Valid {
		resp["meeting_id"] = t.MeetingID.Int64
	}
	if t.Description.Valid {
		resp["description"] = t.Description.String
	} else {
		resp["description"] = nil
	}
	if t.Category.Valid {
		resp["category"] = t.Category.String
	} else {
		resp["category"] = nil
	}
	return resp
}
