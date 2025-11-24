package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "auth/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	DB *pgxpool.Pool
}

// @title           Markdown auth
// @version         1.0
// @description     Auth Server for Markdown-editor

// @host            localhost:8080
// @BasePath        /
func main() {
	var host, port string
	flag.StringVar(&host, "host", "", "Host to bind")
	flag.StringVar(&port, "port", "", "Port to bind")
	flag.Parse()

	dsn := os.Getenv("AUTH_DATABASE_URL")

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer db.Close()

	app := &App{DB: db}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/docs", "./docs")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))
	r.GET("/health", healthHandler)
	r.GET("/v1/check_auth", app.checkAuthHandler)
	r.POST("/v1/register", app.registerHandler)
	r.POST("/v1/login", app.loginHandler)
	r.POST("/v1/refresh", app.refreshHandler)
	r.POST("/v1/logout", app.logoutHandler)

	err = app.DB.Ping(context.Background())
	if err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	serverAddr := fmt.Sprintf("%s:%s", host, port)
	if err := r.Run(serverAddr); err != nil {
		panic(fmt.Sprintf("Failed to run server: %v", err))
	}
	log.Printf("Server started on %s\n", serverAddr)
}
