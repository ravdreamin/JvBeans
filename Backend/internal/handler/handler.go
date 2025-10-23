package handler

import (
	"net/http"

	"jvbeans/internal/auth"
	"jvbeans/internal/models"
	"jvbeans/internal/runner"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Handler holds the database connection
type Handler struct {
	DB *gorm.DB
}

// NewHandler creates a new Handler instance
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// --- Auth Handlers ---

type AuthInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterUser creates a new user
func (h *Handler) RegisterUser(c *gin.Context) {
	var input AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{Username: input.Username, Password: string(hashedPassword)}
	result := h.DB.Create(&user)
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LoginUser logs in a user and returns a JWT
func (h *Handler) LoginUser(c *gin.Context) {
	var input AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Error generating JWT for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// --- Problem Handlers ---

// GetCategories returns all categories with their problems
func (h *Handler) GetCategories(c *gin.Context) {
	var categories []models.Category
	// Preload "Problems" to nest them under each category
	if err := h.DB.Preload("Problems").Find(&categories).Error; err != nil {
		log.Printf("Error fetching categories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetProblem returns a single problem by ID
func (h *Handler) GetProblem(c *gin.Context) {
	id := c.Param("id")
	var problem models.Problem
	if err := h.DB.First(&problem, id).Error; err != nil {
		log.Printf("Error fetching problem %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}
	c.JSON(http.StatusOK, problem)
}

// GetSolution returns the solution for a specific problem
func (h *Handler) GetSolution(c *gin.Context) {
	id := c.Param("id")
	var solution models.Solution
	if err := h.DB.Where("problem_id = ?", id).First(&solution).Error; err != nil {
		log.Printf("Error fetching solution for problem %s: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Solution not found for this problem"})
		return
	}
	c.JSON(http.StatusOK, solution)
}

// --- Runner Handler ---

type CodeRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// RunCode executes code using the Piston runner
func (h *Handler) RunCode(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// You can get the user ID from the context if needed
	// userID, _ := c.Get("userID")
	// log.Printf("User %d is running code", userID)

	result, err := runner.ExecuteCode(req.Language, req.Code)
	if err != nil {
		log.Printf("Error executing code for language %s: %v", req.Language, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Execution failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// --- Health Check ---

// HealthCheck provides a simple health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "API is running"})
}