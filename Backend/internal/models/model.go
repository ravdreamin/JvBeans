package models

import "gorm.io/gorm"

// User represents a user in the database
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}

// Category represents a category of problems
type Category struct {
	gorm.Model
	Name     string    `json:"name"`
	Problems []Problem `json:"problems"`
}

// Problem represents a coding problem
type Problem struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
	CategoryID  uint   `json:"category_id"`
}

// Solution represents a solution to a problem
type Solution struct {
	gorm.Model
	Content   string `json:"content"`
	ProblemID uint   `json:"problem_id"`
}
