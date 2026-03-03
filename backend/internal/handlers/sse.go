package handlers

import (
	"net/http"

	"github.com/simplix/api/internal/middleware"
	"github.com/simplix/api/internal/service"
)

type SSEHandler struct {
	broker *service.SSEBroker
}

func NewSSEHandler(broker *service.SSEBroker) *SSEHandler {
	return &SSEHandler{broker: broker}
}

func (h *SSEHandler) Stream(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	h.broker.ServeSSE(userID, w, r)
}
