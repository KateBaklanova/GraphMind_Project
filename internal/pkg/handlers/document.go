package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/kate/knowledge-graph/internal/pkg/auth"
	"github.com/kate/knowledge-graph/internal/pkg/models"
	"github.com/kate/knowledge-graph/internal/pkg/repository"
	"github.com/kate/knowledge-graph/internal/pkg/utils"
)

type DocumentHandler struct {
	repo *repository.PostgresRepository
}

func NewDocumentHandler(repo *repository.PostgresRepository) *DocumentHandler {
	return &DocumentHandler{repo: repo}
}

// Добавить документ к графу
func (h *DocumentHandler) AddDocument(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	graphID := vars["graph_id"]

	// Граф принадлежит пользователю
	graph, err := h.repo.GetGraphByID(graphID)
	if err != nil || graph.UserID != claims.UserID {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	var req struct {
		Name    string `json:"name"`
		Content string `json:"content"`
		Type    string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация
	if req.Name == "" {
		ValidationError(w, "name", "document name is required")
		return
	}
	if req.Content == "" {
		ValidationError(w, "content", "document content is required")
		return
	}

	doc := &models.Doc{
		ID:        models.NewDocID(),
		GraphID:   graphID,
		Name:      utils.SanitizeInput(req.Name),
		Type:      req.Type,
		Content:   req.Content,
		Size:      int64(len(req.Content)),
		CreatedAt: time.Now(),
	}

	if doc.Type == "" {
		doc.Type = "text"
	}

	if err := h.repo.AddDocument(doc); err != nil {
		Error(w, http.StatusInternalServerError, "failed to add document")
		return
	}

	Success(w, http.StatusCreated, "document added", map[string]interface{}{
		"id":         doc.ID,
		"name":       doc.Name,
		"type":       doc.Type,
		"size":       doc.Size,
		"created_at": doc.CreatedAt,
	})
}

// Получить все документы графа
func (h *DocumentHandler) GetGraphDocuments(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	graphID := vars["graph_id"]

	// Проверка доступа
	graph, err := h.repo.GetGraphByID(graphID)
	if err != nil {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	if !graph.IsPublic && claims.UserID != graph.UserID {
		Error(w, http.StatusForbidden, "access denied")
		return
	}

	documents, err := h.repo.GetDocumentsByGraphID(graphID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get documents")
		return
	}

	Success(w, http.StatusOK, "documents retrieved", documents)
}

// Удалить документ
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	graphID := vars["graph_id"]
	docID := vars["doc_id"]

	if err := h.repo.DeleteDocument(docID, graphID, claims.UserID); err != nil {
		Error(w, http.StatusNotFound, "document not found")
		return
	}

	Success(w, http.StatusOK, "document deleted", nil)
}
