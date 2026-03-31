package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/jaykay/vereinstool/backend/db/generated"
)

type TasksHandler struct {
	queries *generated.Queries
}

func NewTasks(queries *generated.Queries) *TasksHandler {
	return &TasksHandler{queries: queries}
}

func (h *TasksHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	assignedTo := query.Get("assigned_to")
	status := query.Get("status")
	meetingIDStr := query.Get("meeting_id")

	var tasks []generated.Task
	var err error

	if meetingIDStr != "" {
		meetingID, e := parseIntParam(meetingIDStr)
		if e != nil {
			jsonError(w, "Ungültige meeting_id", http.StatusBadRequest)
			return
		}
		tasks, err = h.queries.ListTasksByMeeting(r.Context(), sql.NullInt64{Int64: meetingID, Valid: true})
	} else if assignedTo == "me" {
		user := UserFromContext(r.Context())
		tasks, err = h.queries.ListTasksByAssignee(r.Context(), sql.NullInt64{Int64: user.ID, Valid: true})
	} else if status != "" {
		tasks, err = h.queries.ListTasksByStatus(r.Context(), status)
	} else {
		tasks, err = h.queries.ListTasks(r.Context())
	}

	if err != nil {
		jsonError(w, "Fehler beim Laden der Aufgaben", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, len(tasks))
	for i, t := range tasks {
		result[i] = taskResponse(&t)
	}
	jsonOK(w, result)
}

func (h *TasksHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TopicID     *int64 `json:"topic_id"`
		MeetingID   *int64 `json:"meeting_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		AssignedTo  *int64 `json:"assigned_to"`
		DueDate     string `json:"due_date"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		jsonError(w, "Titel erforderlich", http.StatusBadRequest)
		return
	}

	if req.TopicID == nil || *req.TopicID == 0 {
		jsonError(w, "Thema (topic_id) erforderlich", http.StatusBadRequest)
		return
	}

	var dueDate sql.NullTime
	if req.DueDate != "" {
		t, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			jsonError(w, "Ungültiges Datumsformat (YYYY-MM-DD erwartet)", http.StatusBadRequest)
			return
		}
		dueDate = sql.NullTime{Time: t, Valid: true}
	}

	user := UserFromContext(r.Context())
	task, err := h.queries.CreateTask(r.Context(), generated.CreateTaskParams{
		TopicID:     nullInt64(req.TopicID),
		MeetingID:   nullInt64(req.MeetingID),
		Title:       req.Title,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		AssignedTo:  nullInt64(req.AssignedTo),
		DueDate:     dueDate,
		CreatedBy:   user.ID,
	})
	if err != nil {
		jsonError(w, "Fehler beim Erstellen", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, taskResponse(&task))
}

func (h *TasksHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		jsonError(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	existing, err := h.queries.GetTask(r.Context(), id)
	if err != nil {
		jsonError(w, "Aufgabe nicht gefunden", http.StatusNotFound)
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		AssignedTo  *int64  `json:"assigned_to"`
		DueDate     *string `json:"due_date"`
		Status      *string `json:"status"`
	}
	if err := readJSON(r, &req); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	title := existing.Title
	description := existing.Description
	assignedTo := existing.AssignedTo
	dueDate := existing.DueDate
	status := existing.Status

	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: *req.Description != ""}
	}
	if req.AssignedTo != nil {
		assignedTo = sql.NullInt64{Int64: *req.AssignedTo, Valid: *req.AssignedTo > 0}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			dueDate = sql.NullTime{}
		} else {
			t, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				jsonError(w, "Ungültiges Datumsformat", http.StatusBadRequest)
				return
			}
			dueDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if req.Status != nil {
		if *req.Status != "open" && *req.Status != "done" && *req.Status != "cancelled" {
			jsonError(w, "Ungültiger Status", http.StatusBadRequest)
			return
		}
		status = *req.Status
	}

	err = h.queries.UpdateTask(r.Context(), generated.UpdateTaskParams{
		Title:       title,
		Description: description,
		AssignedTo:  assignedTo,
		DueDate:     dueDate,
		Status:      status,
		ID:          id,
	})
	if err != nil {
		jsonError(w, "Fehler beim Aktualisieren", http.StatusInternalServerError)
		return
	}

	updated, _ := h.queries.GetTask(r.Context(), id)
	jsonOK(w, taskResponse(&updated))
}

func taskResponse(t *generated.Task) map[string]any {
	resp := map[string]any{
		"id":         t.ID,
		"title":      t.Title,
		"status":     t.Status,
		"created_by": t.CreatedBy,
		"created_at": t.CreatedAt.Format(time.RFC3339),
	}
	if t.TopicID.Valid {
		resp["topic_id"] = t.TopicID.Int64
	}
	if t.MeetingID.Valid {
		resp["meeting_id"] = t.MeetingID.Int64
	}
	if t.Description.Valid {
		resp["description"] = t.Description.String
	} else {
		resp["description"] = nil
	}
	if t.AssignedTo.Valid {
		resp["assigned_to"] = t.AssignedTo.Int64
	} else {
		resp["assigned_to"] = nil
	}
	if t.DueDate.Valid {
		resp["due_date"] = t.DueDate.Time.Format("2006-01-02")
	} else {
		resp["due_date"] = nil
	}
	return resp
}
