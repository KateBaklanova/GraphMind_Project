package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/kate/knowledge-graph/internal/pkg/auth"
	"github.com/kate/knowledge-graph/internal/pkg/handlers"
	"github.com/kate/knowledge-graph/internal/pkg/repository"
)

func main() {
	//.env
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "graph_user")
	dbPassword := getEnv("DB_PASSWORD", "graph_password")
	dbName := getEnv("DB_NAME", "graphdb")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-key-change-me")
	port := getEnv("PORT", "8081")

	// Подключение к PostgreSQL
	connStr := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser +
		" password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"

	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer repo.Close()
	log.Println("Connected to PostgreSQL")

	// токен на 24 часа
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	authHandler := handlers.NewAuthHandler(repo, jwtManager)
	graphHandler := handlers.NewGraphHandler(repo)
	documentHandler := handlers.NewDocumentHandler(repo)

	router := mux.NewRouter()

	router.Use(corsMiddleware)

	// Публичные
	router.HandleFunc("/health", healthCheck).Methods("GET")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/public/graphs", graphHandler.GetPublicGraphs).Methods("GET")

	// Защищенные
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(auth.AuthMiddleware(jwtManager))

	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET")
	protected.HandleFunc("/graphs", graphHandler.CreateGraph).Methods("POST")
	protected.HandleFunc("/graphs", graphHandler.GetUserGraphs).Methods("GET")
	protected.HandleFunc("/graphs/{id}", graphHandler.GetGraph).Methods("GET")
	protected.HandleFunc("/graphs/{id}", graphHandler.UpdateGraph).Methods("PUT")
	protected.HandleFunc("/graphs/{id}", graphHandler.DeleteGraph).Methods("DELETE")
	protected.HandleFunc("/graphs/{id}/data", graphHandler.SaveGraphData).Methods("POST")
	protected.HandleFunc("/graphs/{id}/stats", graphHandler.GetGraphStats).Methods("GET")
	protected.HandleFunc("/graphs/{graph_id}/documents", documentHandler.GetGraphDocuments).Methods("GET")
	protected.HandleFunc("/graphs/{graph_id}/documents", documentHandler.AddDocument).Methods("POST")
	protected.HandleFunc("/graphs/{graph_id}/documents/{doc_id}", documentHandler.DeleteDocument).Methods("DELETE")

	// Статический фронтенд
	router.PathPrefix("/").Handler(
		http.FileServer(http.Dir("./web/front")),
	)

	log.Printf("Go server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok", "service": "go-server"}`))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
