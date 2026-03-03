package handlers

import (
	"net/http"
	"time"

	"github.com/simplix/api/internal/repository"
)

type ReportHandler struct{ reports *repository.ReportsRepo }

func NewReportHandler(r *repository.ReportsRepo) *ReportHandler { return &ReportHandler{reports: r} }

func (h *ReportHandler) Overview(w http.ResponseWriter, r *http.Request) {
	stats, err := h.reports.Overview(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, stats)
}

func (h *ReportHandler) Contacts(w http.ResponseWriter, r *http.Request) {
	stats, err := h.reports.ContactStats(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, stats)
}

func (h *ReportHandler) Conversations(w http.ResponseWriter, r *http.Request) {
	stats, err := h.reports.ConversationStats(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, stats)
}

func (h *ReportHandler) Agents(w http.ResponseWriter, r *http.Request) {
	stats, err := h.reports.AgentStats(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, stats)
}

func (h *ReportHandler) TimeSeries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")
	groupBy := q.Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}

	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	rep, err := h.reports.ConversationTimeSeries(r.Context(), from, to, groupBy)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, rep)
}
