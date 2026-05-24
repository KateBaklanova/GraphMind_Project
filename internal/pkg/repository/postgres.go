package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/kate/knowledge-graph/internal/pkg/models"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

// Создание через докер (папка migration)
func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// User
func (r *PostgresRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.Password,
		user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Graph
func (r *PostgresRepository) CreateGraph(graph *models.Graph) error {
	query := `INSERT INTO graphs (id, user_id, name, description, summary, is_public, view_count, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query, graph.ID, graph.UserID, graph.Name, graph.Description,
		graph.Summary, graph.IsPublic, graph.ViewCount, graph.CreatedAt, graph.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetGraphsByUserID(userID string) ([]models.Graph, error) {
	query := `SELECT id, user_id, name, description, summary, is_public, view_count, created_at, updated_at 
			  FROM graphs WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var graphs []models.Graph
	for rows.Next() {
		var g models.Graph
		err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.Description, &g.Summary,
			&g.IsPublic, &g.ViewCount, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		graphs = append(graphs, g)
	}
	return graphs, nil
}

func (r *PostgresRepository) IncrementGraphViews(id string) error {
	_, err := r.db.Exec(`
		UPDATE graphs
		SET view_count = view_count + 1
		WHERE id = $1
	`, id)

	return err
}

func (r *PostgresRepository) GetGraphByID(id string) (*models.Graph, error) {

	var g models.Graph
	query := `SELECT id, user_id, name, description, summary, is_public, view_count, created_at, updated_at 
			  FROM graphs WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&g.ID, &g.UserID, &g.Name, &g.Description,
		&g.Summary, &g.IsPublic, &g.ViewCount, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query("SELECT id, label, type, metadata, created_at FROM nodes WHERE graph_id = $1", id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var n models.Node
			var metadata []byte
			rows.Scan(&n.ID, &n.Label, &n.Type, &metadata, &n.CreatedAt)
			if len(metadata) > 0 {
				json.Unmarshal(metadata, &n.Metadata)
			}
			g.Nodes = append(g.Nodes, n)
		}
	}

	// Загружаем связи
	edgeRows, err := r.db.Query("SELECT id, from_id, to_id, label, weight, metadata, created_at FROM edges WHERE graph_id = $1", id)
	if err == nil {
		defer edgeRows.Close()
		for edgeRows.Next() {
			var e models.Edge
			var metadata []byte
			edgeRows.Scan(&e.ID, &e.FromID, &e.ToID, &e.Label, &e.Weight, &metadata, &e.CreatedAt)
			if len(metadata) > 0 {
				json.Unmarshal(metadata, &e.Metadata)
			}
			g.Edges = append(g.Edges, e)
		}
	}

	return &g, nil
}

func (r *PostgresRepository) UpdateGraph(graph *models.Graph) error {
	query := `UPDATE graphs SET name = $1, description = $2, summary = $3, is_public = $4, updated_at = $5 
			  WHERE id = $6 AND user_id = $7`
	_, err := r.db.Exec(query, graph.Name, graph.Description, graph.Summary,
		graph.IsPublic, graph.UpdatedAt, graph.ID, graph.UserID)
	return err
}

func (r *PostgresRepository) DeleteGraph(id string, userID string) error {
	result, err := r.db.Exec("DELETE FROM graphs WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresRepository) GetPublicGraphs(limit, offset int) ([]models.Graph, error) {
	query := `SELECT id, user_id, name, description, summary, view_count, created_at 
			  FROM graphs WHERE is_public = true 
			  ORDER BY view_count DESC, created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var graphs []models.Graph
	for rows.Next() {
		var g models.Graph
		err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.Description, &g.Summary,
			&g.ViewCount, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		graphs = append(graphs, g)
	}
	return graphs, nil
}

// Node methods
func (r *PostgresRepository) AddNodesBatch(nodes []models.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, node := range nodes {
		metadata, _ := json.Marshal(node.Metadata)

		_, err := tx.ExecContext(ctx, `
			INSERT INTO nodes (
				id, graph_id, label, type, metadata, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (graph_id, id)
			DO UPDATE SET
				label = EXCLUDED.label,
				type = EXCLUDED.type,
				metadata = EXCLUDED.metadata
		`,
			node.ID,
			node.GraphID,
			node.Label,
			node.Type,
			metadata,
			node.CreatedAt,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Edge
func (r *PostgresRepository) AddEdgesBatch(edges []models.Edge) error {
	ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, edge := range edges {
		metadata, _ := json.Marshal(edge.Metadata)

		_, err := tx.ExecContext(ctx, `
			INSERT INTO edges (
				id, graph_id, from_id, to_id,
				label, weight, metadata, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id)
			DO NOTHING
		`,
			edge.ID,
			edge.GraphID,
			edge.FromID,
			edge.ToID,
			edge.Label,
			edge.Weight,
			metadata,
			edge.CreatedAt,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Document
func (r *PostgresRepository) AddDocument(doc *models.Doc) error {
	query := `
		INSERT INTO documents (id, graph_id, name, type, content, size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query, doc.ID, doc.GraphID, doc.Name, doc.Type, doc.Content, doc.Size, doc.CreatedAt)
	return err
}

func (r *PostgresRepository) GetDocumentsByGraphID(graphID string) ([]models.Doc, error) {
	query := `
		SELECT id, graph_id, name, type, content, size, created_at
		FROM documents
		WHERE graph_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []models.Doc
	for rows.Next() {
		var doc models.Doc
		err := rows.Scan(&doc.ID, &doc.GraphID, &doc.Name, &doc.Type, &doc.Content, &doc.Size, &doc.CreatedAt)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *PostgresRepository) DeleteDocument(docID, graphID, userID string) error {
	var graphUserID string
	err := r.db.QueryRow("SELECT user_id FROM graphs WHERE id = $1", graphID).Scan(&graphUserID)
	if err != nil {
		return err
	}
	if graphUserID != userID {
		return sql.ErrNoRows
	}

	result, err := r.db.Exec("DELETE FROM documents WHERE id = $1 AND graph_id = $2", docID, graphID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
