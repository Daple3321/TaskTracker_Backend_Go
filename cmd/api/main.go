package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Daple3321/TaskTracker/internal/handlers"
	"github.com/Daple3321/TaskTracker/internal/middleware"
	"github.com/Daple3321/TaskTracker/internal/repositories"
	"github.com/joho/godotenv"

	"github.com/redis/go-redis/v9"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if err := run(); err != nil {
		// Ensure there's always something visible in container logs even if
		// logging setup fails or the error happens very early.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logFile, err := SetupLogger()
	if err != nil {
		return fmt.Errorf("setup logger: %w", err)
	}
	defer logFile.Close()

	envPath := filepath.Join("..", "..", "configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		slog.Warn("no .env file found, using process environment", "envPath", envPath)
	}

	if err := ValidateEnvVars(); err != nil {
		slog.Error("error validating env vars", "err", err)
		return fmt.Errorf("validate env vars: %w", err)
	}

	// -- DB Setup --
	db, err := SetupDB()
	if err != nil {
		return fmt.Errorf("setup db: %w", err)
	}

	dir, err := resolveMigrationsDir()
	if err != nil {
		slog.Error("migrations path", "err", err)
		return fmt.Errorf("resolve migrations dir: %w", err)
	}
	sourceURL, err := migrationSourceURL(dir)
	if err != nil {
		slog.Error("migrations file URL", "err", err)
		return fmt.Errorf("build migrations URL: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		slog.Error("error creating mysql driver", "err", err)
		return fmt.Errorf("create mysql migrate driver: %w", err)
	}
	migrations, err := migrate.NewWithDatabaseInstance(sourceURL, "tasks", driver)
	if err != nil {
		slog.Error("error creating migrations", "err", err)
		return fmt.Errorf("create migrations instance: %w", err)
	}
	defer migrations.Close()

	err = migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("error running migrations", "err", err)
		return fmt.Errorf("run migrations: %w", err)
	}
	// ------------------

	// -- Redis setup --
	redisCfgPath := getEnv("REDIS_CFG_PATH", "/app/configs/redisCfg.yml")
	redisCfg, err := repositories.LoadRedisConfig(redisCfgPath)
	if err != nil {
		slog.Error("error loading redis cfg", "err", err)
		return fmt.Errorf("error loading redis cfg: %w", err)
	}

	rdb, err := SetupRedis(ctx, redisCfg)
	if err != nil {
		slog.Error("error setting up redis", "err", err)
		return fmt.Errorf("error setting up redis: %w", err)
	}
	// -------------------

	go middleware.LimitTimeoutRoutine(ctx)

	usersHandler := handlers.NewUsersHandler(db)
	authRouter := usersHandler.RegisterRoutes()

	tasksHandler := handlers.NewTaskHandler(db, rdb)
	tasksRouter := tasksHandler.RegisterRoutes()

	router := http.NewServeMux()
	router.Handle("/tasks/", http.StripPrefix("/tasks", tasksRouter))

	router.Handle("/auth/", http.StripPrefix("/auth", authRouter))

	handler := corsMiddleware(router)

	serverIP := getEnv("SERVERIP", "0.0.0.0")
	serverPort := os.Getenv("SERVERPORT")
	slog.Info("Listening on:", "ip", serverIP, "port", serverPort)
	if err := http.ListenAndServe(serverIP+":"+serverPort, handler); err != nil {
		slog.Error("error starting http server", "err", err)
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

func ValidateEnvVars() error {
	vars := []string{
		"SERVERPORT",
		"TASKDB_USERNAME",
		"TASKDB_PASSWORD",
		"JWT_SECRET_KEY",
	}

	for _, v := range vars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("env var %s not set", v)
		}
	}

	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", getEnv("FRONTEND_ORIGIN", "http://localhost:5173"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// resolveMigrationsDir returns MIGRATIONS_PATH if set, otherwise the first
// existing internal/migrations directory relative to the current working directory.
func resolveMigrationsDir() (string, error) {
	if d := os.Getenv("MIGRATIONS_PATH"); d != "" {
		return d, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(wd, "internal", "migrations"),
		filepath.Join(wd, "..", "internal", "migrations"),
		filepath.Join(wd, "..", "..", "internal", "migrations"),
	}
	for _, c := range candidates {
		fi, err := os.Stat(c)
		if err == nil && fi.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("migrations dir not found under %q; set MIGRATIONS_PATH", wd)
}

// migrationSourceURL builds a file:// URL golang-migrate accepts on Unix and Windows.
func migrationSourceURL(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	p := filepath.ToSlash(abs)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p, nil
}

func SetupDB() (*sql.DB, error) {
	dbHost := getEnv("TASKDB_HOST", "localhost")
	newDb, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/tasks?parseTime=true&multiStatements=true",
			os.Getenv("TASKDB_USERNAME"),
			os.Getenv("TASKDB_PASSWORD"),
			dbHost))
	if err != nil {
		slog.Error("error opening database", "err", err)
		return nil, err
	}

	pingErr := newDb.Ping()
	if pingErr != nil {
		slog.Error("error while pinging DB", "err", pingErr)
		return nil, pingErr
	}

	return newDb, nil
}

func SetupRedis(ctx context.Context, cfg repositories.RedisConfig) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		//Username:     cfg.User,
		//Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to redis server", "err", err)
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func SetupLogger() (*os.File, error) {
	workDir, _ := os.Getwd()
	logPath := path.Join(workDir, "server.log")
	//os.WriteFile(logPath, []byte{}, os.ModeAppend)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "err", err)
		return nil, err
	}

	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	w := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewJSONHandler(w, &opts))
	slog.SetDefault(logger)

	return logFile, nil
}
