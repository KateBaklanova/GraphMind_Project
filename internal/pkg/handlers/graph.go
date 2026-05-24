package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/kate/knowledge-graph/internal/pkg/auth"
	"github.com/kate/knowledge-graph/internal/pkg/models"
	"github.com/kate/knowledge-graph/internal/pkg/repository"
	"github.com/kate/knowledge-graph/internal/pkg/utils"
)

type GraphHandler struct {
	repo *repository.PostgresRepository
}

func NewGraphHandler(repo *repository.PostgresRepository) *GraphHandler {
	return &GraphHandler{
		repo: repo,
	}
}

// Создание графа
func (h *GraphHandler) CreateGraph(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	var req models.CreateGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if valid, msg := utils.ValidateGraphName(req.Name); !valid {
		ValidationError(w, "name", msg)
		return
	}

	graph := &models.Graph{
		ID:          models.NewGraphID(),
		UserID:      claims.UserID,
		Name:        utils.SanitizeInput(req.Name),
		Description: utils.SanitizeInput(req.Description),
		IsPublic:    req.IsPublic,
		ViewCount:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.CreateGraph(graph); err != nil {

		log.Printf(
			"CreateGraph repo error: %v",
			err,
		)

		Error(w, http.StatusInternalServerError, err.Error())

		return
	}

	Success(w, http.StatusCreated, "graph created", graph)
}

// Все графы пользователя
func (h *GraphHandler) GetUserGraphs(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	graphs, err := h.repo.GetGraphsByUserID(claims.UserID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get graphs")
		return
	}

	Success(w, http.StatusOK, "graphs retrieved", graphs)
}

// Получить граф по ID
func (h *GraphHandler) GetGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	graph, err := h.repo.GetGraphByID(graphID)
	if err != nil {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	// Проверка доступа
	claims, _ := auth.GetUserFromContext(r.Context())
	if !graph.IsPublic && (claims == nil || claims.UserID != graph.UserID) {
		Error(w, http.StatusForbidden, "access denied")
		return
	}

	go h.repo.IncrementGraphViews(graphID)

	Success(w, http.StatusOK, "graph retrieved", graph)
}

// Обновление графа
func (h *GraphHandler) UpdateGraph(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	graphID := vars["id"]

	existing, err := h.repo.GetGraphByID(graphID)
	if err != nil || existing.UserID != claims.UserID {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	var req models.UpdateGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != "" {
		if valid, msg := utils.ValidateGraphName(req.Name); !valid {
			ValidationError(w, "name", msg)
			return
		}
		existing.Name = utils.SanitizeInput(req.Name)
	}

	if req.Description != "" {
		existing.Description = utils.SanitizeInput(req.Description)
	}

	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}

	existing.UpdatedAt = time.Now()

	if err := h.repo.UpdateGraph(existing); err != nil {
		Error(w, http.StatusInternalServerError, "failed to update graph")
		return
	}

	log.Println("UPDATED:", graphID)

	Success(w, http.StatusOK, "graph updated", existing)
}

// DeleteGraph - удаление графа
func (h *GraphHandler) DeleteGraph(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	graphID := vars["id"]

	if err := h.repo.DeleteGraph(graphID, claims.UserID); err != nil {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	log.Println("DELETED:", graphID)

	Success(w, http.StatusOK, "graph deleted", nil)
}

// Публичные графы
func (h *GraphHandler) GetPublicGraphs(w http.ResponseWriter, r *http.Request) {
	graphs, err := h.repo.GetPublicGraphs(20, 0)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get public graphs")
		return
	}

	Success(w, http.StatusOK, "public graphs retrieved", graphs)
}

// Сохранение данных графа (ноды и ребра)
func normalizeID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func (h *GraphHandler) SaveGraphData(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	graphID := vars["id"]

	start := time.Now()

	log.Printf(
		"[SaveGraphData] START graph=%s",
		graphID,
	)

	graph, err := h.repo.GetGraphByID(graphID)
	if err != nil || graph.UserID != claims.UserID {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	var req models.GraphDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	nodeMap := make(map[string]bool)

	// Создаем ноды
	nodes := make([]models.Node, 0)

	for _, n := range req.Nodes {
		id := normalizeID(n.ID)

		nodeMap[id] = true

		nodes = append(nodes, models.Node{
			ID:        id,
			GraphID:   graphID,
			Label:     utils.SanitizeInput(n.Label),
			Type:      n.Type,
			CreatedAt: time.Now(),
		})
	}

	// Добавление недостающих нод
	for _, e := range req.Edges {
		from := normalizeID(e.From)
		to := normalizeID(e.To)

		if from == "" || to == "" {
			Error(w, http.StatusBadRequest, "edge has empty from/to")
			return
		}

		if from == to {
			Error(w, http.StatusBadRequest, "self-loop not allowed")
			return
		}

		if !nodeMap[from] {
			nodeMap[from] = true
			nodes = append(nodes, models.Node{
				ID:        from,
				GraphID:   graphID,
				Label:     from,
				Type:      "auto",
				CreatedAt: time.Now(),
			})
		}

		if !nodeMap[to] {
			nodeMap[to] = true
			nodes = append(nodes, models.Node{
				ID:        to,
				GraphID:   graphID,
				Label:     to,
				Type:      "auto",
				CreatedAt: time.Now(),
			})
		}
	}

	// Сохранение нод
	log.Println("NODES:", len(nodes))
	log.Println("EDGES (before save):", len(req.Edges))
	if err := h.repo.AddNodesBatch(nodes); err != nil {
		log.Println("nodes error:", err)
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Создание ребер
	edges := make([]models.Edge, 0)

	for _, e := range req.Edges {
		from := normalizeID(e.From)
		to := normalizeID(e.To)

		edges = append(edges, models.Edge{
			ID:        models.NewEdgeID(),
			GraphID:   graphID,
			FromID:    from,
			ToID:      to,
			Label:     utils.SanitizeInput(e.Label),
			Weight:    1.0,
			CreatedAt: time.Now(),
		})
	}

	// Сохранение ребер
	if err := h.repo.AddEdgesBatch(edges); err != nil {
		log.Println("edges error:", err)
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	defer func() {
		log.Printf(
			"[SaveGraphData] DONE graph=%s duration=%s",
			graphID,
			time.Since(start),
		)
	}()

	Success(w, http.StatusOK, "graph data saved", map[string]interface{}{
		"nodes_count": len(nodes),
		"edges_count": len(edges),
	})
}

// Статистика графа
// Убрать ???
func (h *GraphHandler) GetGraphStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	graph, err := h.repo.GetGraphByID(graphID)
	if err != nil {
		Error(w, http.StatusNotFound, "graph not found")
		return
	}

	stats := map[string]interface{}{
		"nodes_count":     len(graph.Nodes),
		"edges_count":     len(graph.Edges),
		"documents_count": len(graph.Documents),
		"view_count":      graph.ViewCount,
		"created_at":      graph.CreatedAt,
		"updated_at":      graph.UpdatedAt,
	}

	Success(w, http.StatusOK, "graph statistics", stats)
}
