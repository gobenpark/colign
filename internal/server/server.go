package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gobenpark/CoSpec/internal/auth"
	"github.com/gobenpark/CoSpec/internal/config"
	"github.com/gobenpark/CoSpec/internal/database"
	"github.com/gobenpark/CoSpec/internal/middleware"
	"github.com/uptrace/bun"

	"github.com/gobenpark/CoSpec/gen/proto/auth/v1/authv1connect"
	"github.com/gobenpark/CoSpec/gen/proto/project/v1/projectv1connect"
	"github.com/gobenpark/CoSpec/internal/project"
)

type Server struct {
	router     *gin.Engine
	db         *bun.DB
	jwtManager *auth.JWTManager
}

func New(cfg *config.Config) (*Server, error) {
	db := database.New(cfg.DatabaseURL, cfg.Debug)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	s := &Server{
		router:     gin.Default(),
		db:         db,
		jwtManager: jwtManager,
	}

	s.setupMiddleware()
	s.setupRoutes(cfg)

	return s, nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Connect-Protocol-Version"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func (s *Server) setupRoutes(cfg *config.Config) {
	s.router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth service (Connect)
	authService := auth.NewService(s.db, s.jwtManager)
	oauthService := auth.NewOAuthService(s.db, s.jwtManager, auth.OAuthConfig{
		GitHubClientID:     cfg.GitHubClientID,
		GitHubClientSecret: cfg.GitHubClientSecret,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		RedirectBaseURL:    cfg.RedirectBaseURL,
	})

	authConnectHandler := auth.NewConnectHandler(authService, oauthService)
	path, handler := authv1connect.NewAuthServiceHandler(authConnectHandler)
	s.router.Any(path+"/*path", gin.WrapH(handler))

	// REST auth routes (OAuth redirects)
	api := s.router.Group("/api")
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(api)

	oauthHandler := auth.NewOAuthHandler(oauthService, cfg.FrontendURL)
	oauthHandler.RegisterRoutes(api)

	// Project service (Connect)
	projectService := project.NewService(s.db)
	projectConnectHandler := project.NewConnectHandler(projectService)
	projectPath, projectHandler := projectv1connect.NewProjectServiceHandler(projectConnectHandler)
	s.router.Any(projectPath+"/*path", gin.WrapH(projectHandler))

	// Protected routes group
	_ = s.router.Group("/api").Use(middleware.JWTAuth(s.jwtManager))
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Close() error {
	return s.db.Close()
}
