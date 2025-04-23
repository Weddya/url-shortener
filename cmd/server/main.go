package main

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"url-shortener/internal/handler"
	"url-shortener/internal/middleware"
	"url-shortener/internal/repository"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort int
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	cfg := loadConfig()

	fmt.Println(cfg)
	conn, err := pgxpool.New(context.Background(), buildConnString(cfg))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err = conn.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	repo := repository.NewPostgresRepo(conn)
	metrics := middleware.NewMetrics("url_shortener")
	urlHandler := handler.NewURLHandler(repo)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	e.Use(middleware.PrometheusMiddleware(metrics))

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.POST("/create", urlHandler.CreateShortURL)
	e.GET("/:code", urlHandler.RedirectByCode)

	// Graceful shutdown
	go func() {
		addr := fmt.Sprintf(":%d", cfg.ServerPort)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func loadConfig() Config {
	_ = godotenv.Load()

	port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "url_shortener"),
		ServerPort: port,
	}
}

func buildConnString(cfg Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
