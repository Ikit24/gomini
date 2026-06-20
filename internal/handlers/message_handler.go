package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Ikit24/gomini/internal/database"
	"github.com/Ikit24/gomini/internal/gemini"
	"github.com/google/uuid"
)

func (s *Server) HandleCreateMessage(w http.ResponseWriter, r *http.Request) {
	type messageParams struct {
		Content string `json:"content"`
	}

	var params messageParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	sessionIDString := r.PathValue("session_id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session_id format")
		return
	}

	session, err := s.DB.GetSessionByID(sessionID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "session not found")
		return
	}

	userMessage := database.Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		UserID:    session.UserID,
		Role:      database.UserRole,
		Content:   params.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.DB.CreateMessage(&userMessage)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't create message")
		return
	}

	dbMessages, err := s.DB.GetMessagesBySessionID(sessionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't retrieve message")
		return
	}

	geminiMessages := make([]gemini.Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		gMsg := gemini.Message{
			Role:    string(dbMsg.Role),
			Content: dbMsg.Content,
		}
		geminiMessages = append(geminiMessages, gMsg)
	}

	aiResponse, err := s.AI.GenerateChatResponse(r.Context(), geminiMessages, params.Content)
	if err != nil {
		fmt.Printf("AI Client error: %v\n", err)
		RespondWithError(w, http.StatusInternalServerError, "failed to get response from the AI")
		return
	}

	var builder strings.Builder

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	for text := range aiResponse {
		fmt.Fprint(w, text)
		flusher.Flush()
		builder.WriteString(text)
	}

	fmt.Fprint(w, "\n")

	aiMessage := database.Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		UserID:    session.UserID,
		Role:      database.ModelRole,
		Content:   builder.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.DB.CreateMessage(&aiMessage)
	if err != nil {
		fmt.Printf("database error: couldn't save AI response: %v\n", err)
		return
	}
}

func (s *Server) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	sessionIDString := r.PathValue("session_id")
	sessionID, err := uuid.Parse(sessionIDString)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid session id format")
		return
	}

	messages, err := s.DB.GetMessagesBySessionID(sessionID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "couldn't retrieve messages")
		return
	}

	RespondWithJSON(w, http.StatusOK, messages)
}
