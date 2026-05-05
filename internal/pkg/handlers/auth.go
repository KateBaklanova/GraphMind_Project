package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kate/knowledge-graph/internal/pkg/auth"
	"github.com/kate/knowledge-graph/internal/pkg/models"
	"github.com/kate/knowledge-graph/internal/pkg/repository"
	"github.com/kate/knowledge-graph/internal/pkg/utils"
)

type AuthHandler struct {
	repo       *repository.PostgresRepository
	jwtManager *auth.JWTManager
}

func NewAuthHandler(repo *repository.PostgresRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

// Регистрация
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация
	if req.Username == "" || len(req.Username) < 3 {
		ValidationError(w, "username", "username must be at least 3 characters")
		return
	}

	if !utils.ValidateEmail(req.Email) {
		ValidationError(w, "email", "invalid email format")
		return
	}

	if valid, msg := utils.ValidatePassword(req.Password); !valid {
		ValidationError(w, "password", msg)
		return
	}

	// Хэш пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Новый пользователь
	user := &models.User{
		ID:        models.NewUserID(),
		Username:  utils.SanitizeInput(req.Username),
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.repo.CreateUser(user); err != nil {
		Error(w, http.StatusConflict, "user with this email already exists")
		return
	}

	// Генерация токена
	token, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	user.Password = ""
	Success(w, http.StatusCreated, "user registered successfully", models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Авторизация
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "email and password required")
		return
	}

	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	user.Password = ""
	Success(w, http.StatusOK, "login successful", models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Получить текущего пользователя
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.RequireAuth(w, r)
	if !ok {
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil {
		Error(w, http.StatusNotFound, "user not found")
		return
	}

	user.Password = ""
	Success(w, http.StatusOK, "user retrieved", user)
}
