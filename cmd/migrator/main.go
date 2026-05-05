package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// Команды
var (
	up      = flag.Bool("up", false, "Apply all pending migrations")
	down    = flag.Bool("down", false, "Rollback last migration")
	initDB  = flag.Bool("init", false, "Initialize database (create migrations table)")
	version = flag.Bool("version", false, "Show current migration version")
)

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

func main() {
	flag.Parse()

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "graph_user")
	dbPassword := getEnv("DB_PASSWORD", "graph_password")
	dbName := getEnv("DB_NAME", "graphdb")
	migrationsDir := getEnv("MIGRATIONS_DIR", "./migrations")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database")

	if *initDB {
		if err := initMigrationTable(db); err != nil {
			log.Fatal("Failed to init migration table:", err)
		}
		log.Println("Migration table initialized")
		return
	}

	if *version {
		currentVersion, err := getCurrentVersion(db)
		if err != nil {
			log.Fatal("Failed to get current version:", err)
		}
		log.Printf("Current migration version: %d", currentVersion)
		return
	}

	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		log.Fatal("Failed to load migrations:", err)
	}

	if len(migrations) == 0 {
		log.Println("No migrations found")
		return
	}

	if *up {
		if err := applyUpMigrations(db, migrations); err != nil {
			log.Fatal("Failed to apply migrations:", err)
		}
		log.Println("All migrations applied successfully")
	}

	if *down {
		if err := rollbackLastMigration(db, migrations); err != nil {
			log.Fatal("Failed to rollback migration:", err)
		}
		log.Println("Migration rolled back successfully")
	}
}

// Создает таблицу для отслеживания миграций
func initMigrationTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
			down_sql TEXT
		);
	`
	_, err := db.Exec(query)
	return err
}

// Получает текущую версию миграции
func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	query := `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	err := db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// Загружает миграции из файлов
func loadMigrations(dir string) ([]Migration, error) {
	// .up.sql файлы
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	var migrations []Migration

	for _, upFile := range files {
		base := filepath.Base(upFile)
		parts := strings.Split(base, "_")
		if len(parts) < 2 {
			log.Printf("Skipping invalid migration file: %s", base)
			continue
		}

		version := 0
		fmt.Sscanf(parts[0], "%d", &version)

		upSQL, err := os.ReadFile(upFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", upFile, err)
		}

		downFile := filepath.Join(dir, fmt.Sprintf("%d_%s.down.sql", version, strings.Join(parts[1:], "_")))
		downSQL := []byte{}
		if _, err := os.Stat(downFile); err == nil {
			downSQL, _ = os.ReadFile(downFile)
		}

		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Применяет все ожидающие миграции
func applyUpMigrations(db *sql.DB, migrations []Migration) error {
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, m := range migrations {
		if m.Version <= currentVersion {
			log.Printf("⏭️ Skipping migration %d (%s) - already applied", m.Version, m.Name)
			continue
		}

		log.Printf("🔄 Applying migration %d (%s)...", m.Version, m.Name)

		if _, err := tx.Exec(m.UpSQL); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", m.Version, err)
		}

		insertQuery := `
			INSERT INTO schema_migrations (version, down_sql) 
			VALUES ($1, $2)
		`
		if _, err := tx.Exec(insertQuery, m.Version, m.DownSQL); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}

		log.Printf("Migration %d applied", m.Version)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Откат последней миграции
func rollbackLastMigration(db *sql.DB, migrations []Migration) error {
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	if currentVersion == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	var lastMigration Migration
	for _, m := range migrations {
		if m.Version == currentVersion {
			lastMigration = m
			break
		}
	}

	if lastMigration.Version == 0 {
		return fmt.Errorf("migration %d not found", currentVersion)
	}

	log.Printf("Rolling back migration %d (%s)...", lastMigration.Version, lastMigration.Name)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if lastMigration.DownSQL != "" {
		if _, err := tx.Exec(lastMigration.DownSQL); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", lastMigration.Version, err)
		}
	}

	deleteQuery := `DELETE FROM schema_migrations WHERE version = $1`
	if _, err := tx.Exec(deleteQuery, lastMigration.Version); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Migration %d rolled back", lastMigration.Version)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
