package app

import (
	"errors"
	"time"

	"github.com/kate/knowledge-graph/internal/pkg/models"
	"github.com/kate/knowledge-graph/internal/pkg/repository"
	"github.com/kate/knowledge-graph/internal/pkg/utils"
)

// Сервис для работы с графами (бизнес-логика)
type GraphService struct {
	repo *repository.PostgresRepository
}

// Новый сервис
func NewGraphService(repo *repository.PostgresRepository) *GraphService {
	return &GraphService{
		repo: repo,
	}
}

// Новый граф
func (s *GraphService) CreateGraph(userID, name, description string, isPublic bool) (*models.Graph, error) {
	if valid, msg := utils.ValidateGraphName(name); !valid {
		return nil, errors.New(msg)
	}

	graph := &models.Graph{
		ID:          models.NewGraphID(),
		UserID:      userID,
		Name:        utils.SanitizeInput(name),
		Description: utils.SanitizeInput(description),
		IsPublic:    isPublic,
		ViewCount:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateGraph(graph); err != nil {
		return nil, errors.New("failed to create graph")
	}

	return graph, nil
}

// Все графы пользователя
func (s *GraphService) GetUserGraphs(userID string) ([]models.Graph, error) {
	return s.repo.GetGraphsByUserID(userID)
}

// Граф по ID с проверкой доступа
func (s *GraphService) GetGraphByID(graphID, userID string) (*models.Graph, error) {
	graph, err := s.repo.GetGraphByID(graphID)
	if err != nil {
		return nil, errors.New("graph not found")
	}

	if !graph.IsPublic && userID != graph.UserID {
		return nil, errors.New("access denied")
	}

	return graph, nil
}

// Обновляет граф
func (s *GraphService) UpdateGraph(graphID, userID, name, description string, isPublic *bool) (*models.Graph, error) {
	graph, err := s.repo.GetGraphByID(graphID)
	if err != nil || graph.UserID != userID {
		return nil, errors.New("graph not found")
	}

	if name != "" {
		if valid, msg := utils.ValidateGraphName(name); !valid {
			return nil, errors.New(msg)
		}
		graph.Name = utils.SanitizeInput(name)
	}

	if description != "" {
		graph.Description = utils.SanitizeInput(description)
	}

	if isPublic != nil {
		graph.IsPublic = *isPublic
	}

	graph.UpdatedAt = time.Now()

	if err := s.repo.UpdateGraph(graph); err != nil {
		return nil, errors.New("failed to update graph")
	}

	return graph, nil
}

// Удаляет граф
func (s *GraphService) DeleteGraph(graphID, userID string) error {
	return s.repo.DeleteGraph(graphID, userID)
}

// Получает публичные графы
func (s *GraphService) GetPublicGraphs(limit, offset int) ([]models.Graph, error) {
	return s.repo.GetPublicGraphs(limit, offset)
}

// Сохраняет ноды и ребра графа
func (s *GraphService) SaveGraphData(graphID, userID string, nodes []models.NodeInput, edges []models.EdgeInput) error {
	graph, err := s.repo.GetGraphByID(graphID)
	if err != nil || graph.UserID != userID {
		return errors.New("graph not found")
	}

	graphNodes := make([]models.Node, len(nodes))
	for i, n := range nodes {
		graphNodes[i] = models.Node{
			ID:        n.ID,
			GraphID:   graphID,
			Label:     utils.SanitizeInput(n.Label),
			Type:      n.Type,
			CreatedAt: time.Now(),
		}
	}

	if err := s.repo.AddNodesBatch(graphNodes); err != nil {
		return errors.New("failed to save nodes")
	}

	graphEdges := make([]models.Edge, len(edges))
	for i, e := range edges {
		graphEdges[i] = models.Edge{
			ID:        models.NewEdgeID(),
			GraphID:   graphID,
			FromID:    e.From,
			ToID:      e.To,
			Label:     utils.SanitizeInput(e.Label),
			Weight:    1.0,
			CreatedAt: time.Now(),
		}
	}

	if err := s.repo.AddEdgesBatch(graphEdges); err != nil {
		return errors.New("failed to save edges")
	}

	return nil
}

// Получает статистику графа
func (s *GraphService) GetGraphStats(graphID string) (map[string]interface{}, error) {
	graph, err := s.repo.GetGraphByID(graphID) // дубли
	if err != nil {
		return nil, errors.New("graph not found")
	}

	return map[string]interface{}{
		"nodes_count":     len(graph.Nodes),
		"edges_count":     len(graph.Edges),
		"documents_count": len(graph.Documents),
		"view_count":      graph.ViewCount,
		"created_at":      graph.CreatedAt,
		"updated_at":      graph.UpdatedAt,
	}, nil
}
