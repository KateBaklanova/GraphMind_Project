package models

import (
	"time"

	"github.com/google/uuid"
)

// User - пользователь
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // не возвращаем в JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Graph - граф знаний
type Graph struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	IsPublic    bool      `json:"is_public"`
	ViewCount   int       `json:"view_count"`
	Documents   []Doc     `json:"documents,omitempty"`
	Nodes       []Node    `json:"nodes,omitempty"`
	Edges       []Edge    `json:"edges,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Node - узел
type Node struct {
	ID        string                 `json:"id"`
	GraphID   string                 `json:"graph_id,omitempty"`
	Label     string                 `json:"label"`
	Type      string                 `json:"type,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// Edge - связь
type Edge struct {
	ID        string                 `json:"id"`
	GraphID   string                 `json:"graph_id,omitempty"`
	FromID    string                 `json:"from"`
	ToID      string                 `json:"to"`
	Label     string                 `json:"label"`
	Weight    float64                `json:"weight,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// Doc - документ
type Doc struct {
	ID        string    `json:"id"`
	GraphID   string    `json:"graph_id,omitempty"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Content   string    `json:"content,omitempty"`
	Size      int64     `json:"size,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Claims - JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateGraphRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

type UpdateGraphRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsPublic    *bool  `json:"is_public,omitempty"`
}

type GraphDataRequest struct {
	Nodes []NodeInput `json:"nodes"`
	Edges []EdgeInput `json:"edges"`
}

type NodeInput struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type,omitempty"`
}

type EdgeInput struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

// ID generator
func NewUserID() string  { return uuid.New().String() }
func NewGraphID() string { return uuid.New().String() }
func NewNodeID() string  { return uuid.New().String() }
func NewEdgeID() string  { return uuid.New().String() }
func NewDocID() string   { return uuid.New().String() }
