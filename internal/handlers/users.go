package handlers

import (
    "database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/lib/pq"
)

type User struct {
	ID int `json:"id"`
	Username string `json:"username"`
}

// User creation endpoint
func CreateUser(c *gin.Context, db *sql.DB) {
	// Parse JSON request body into User struct
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the username already exists
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username availability"})
        return
    }

    if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        return
    }

	// Insert user into database and return the inserted ID
    var id int
    err = db.QueryRow("INSERT INTO users (username) VALUES ($1) RETURNING id", user.Username).Scan(&id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	// Return ID of newly inserted user
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Login endpoint
func Login(c *gin.Context, db *sql.DB) {
	// Parse JSON request body into User struct
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query database for user
	row := db.QueryRow("SELECT id FROM users WHERE username = $1", user.Username)

	// Get ID of user
	var id int
	err := row.Scan(&id)
	if err != nil {
		// Check if user was not found
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
			return
		}
		// Return error if other error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// Return ID of user
	c.JSON(http.StatusOK, gin.H{"id": id})
}