package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// Models
type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Response struct {
	Success   bool        `json:"id"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

var (
	articles = []Article{
		{ID: 1, Title: "Getting Started with Go", Content: "Go is a programming language...", Author: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "Web Development with Gin", Content: "Gin is a web framework...", Author: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	nextId     = 3
	articleMux sync.Mutex
)

// Main program
func main() {
	r := gin.New()
	r.Use(
		ErrorHandlerMiddleware(),
		RequestIDMiddleware(),
		LoggingMiddleware(),
		CORSMiddleware(),
		RateLimitMiddleware(),
		ContentTypeMiddleware(),
	)
	// public routes
	public := r.Group("/")
	{
		public.GET("/ping", ping)
		public.GET("/articles", getArticles)
		public.GET("/articles/:id", getArticleById)
	}

	//protected routes
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/articles", createArticle)
		protected.PUT("/articles/:id", updateArticle)
		protected.DELETE("/article/:id", deleteArticle)
		protected.GET("/admin/stats", getStats)
	}

	log.Println("Server running on :8080")
	r.Run(":8080")
}

// essential middlewares

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Response{
			Success:   false,
			Error:     "interal server error",
			RequestID: c.GetString("request_id"),
		})
	})
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID, _ := c.Get("request_id")
		duration := time.Since(start)

		log.Printf(
			"[%s] %s %s %d %s %s %s",
			reqID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
			c.Request.UserAgent(),
		)
	}
}

func AuthMiddleware() gin.HandlerFunc {
	apiKeys := map[string]string{
		"admin-key":    "admin",
		"user-key-456": "user",
	}

	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		role, ok := apiKeys[key]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, Response{
				Success:   false,
				Error:     "invalid or missing API key",
				RequestID: c.GetString("request_id"),
			})
		}
		c.Set("role", role)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"https://myblog.com":    true,
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-type, X-API-Key, X-Request-ID")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	var visitors = make(map[string]*rate.Limiter)
	var mu sync.Mutex

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		limiter, exists := visitors[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 100)
			visitors[ip] = limiter
		}
		return limiter
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		c.Header("X-RateLimit-Limit", "100")

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, Response{
				Success:   false,
				Error:     "too many requests, limit exceeded",
				RequestID: c.GetString("request_id"),
			})
			return
		}
		c.Next()
	}
}

func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			if !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, Response{
					Success:   false,
					Error:     "Content type must be application/json",
					RequestID: c.GetString("request_id"),
				})
				return
			}
		}
		c.Next()
	}
}

// Creating the handlers

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   "pong",
		RequestID: c.GetString("request_id"),
	})
}

func getArticles(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      articles,
		RequestID: c.GetString("request_id"),
	})
}

func getArticleById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	article, _ := findArticleByID(id)

	if article == nil {
		c.JSON(http.StatusNotFound, Response{
			Success:   false,
			Error:     "article not found",
			RequestID: c.GetString("request_id"),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      article,
		RequestID: c.GetString("request_id"),
	})
}

func createArticle(c *gin.Context) {
	var input Article
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Error: err.Error()})
		return
	}

	if err := validateArticle(input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Success: false, Error: err.Error()})
		return
	}

	articleMux.Lock()
	defer articleMux.Unlock()

	input.ID = nextId
	nextId++
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()
	articles = append(articles, input)

	c.JSON(http.StatusCreated, Response{Success: true, Data: input})
}

func updateArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	article, index := findArticleByID(id)

	if article == nil {
		c.JSON(http.StatusNotFound, Response{Success: false, Error: "article not found"})
		return
	}

	var input Article
	c.ShouldBindJSON(&input)

	articles[index].Title = input.Title
	articles[index].Content = input.Content
	articles[index].Author = input.Author
	articles[index].UpdatedAt = time.Now()

	c.JSON(http.StatusOK, Response{Success: true, Data: articles[index]})
}

func deleteArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, index := findArticleByID(id)

	if index == -1 {
		c.JSON(http.StatusNotFound, Response{Success: false, Error: "article not found"})
		return
	}

	articles = append(articles[:index], articles[index+1:]...)
	c.JSON(http.StatusOK, Response{Success: true, Message: "article deleted"})
}

func getStats(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, Response{Success: false, Error: "admin access required"})
		return
	}

	stats := map[string]interface{}{
		"total_articles": len(articles),
		"uptime":         "24h",
	}

	c.JSON(http.StatusOK, Response{Success: true, Data: stats})
}

func findArticleByID(id int) (*Article, int) {
	for i, a := range articles {
		if a.ID == id {
			return &a, i
		}
	}
	return nil, -1
}

func validateArticle(article Article) error {
	if article.Title == "" || article.Content == "" || article.Author == "" {
		return gin.Error{Err: http.ErrMissingFile}
	}
	return nil
}
