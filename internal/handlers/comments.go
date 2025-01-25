package handlers

import (
    "database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

type Comment struct {
	ID int `json:"id"`
	ThreadID int `json:"thread_id"`
	UserID int `json:"user_id"`
	UserName string `json:"user_name"`
	Text string `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// Comment creation endpoint
func CreateComment(c *gin.Context, db *sql.DB) {
	// Parse JSON request body into Comment struct
	var comment Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		log.Printf("Error inserting comment: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert comment into database
	result, err := db.Exec("INSERT INTO comments (thread_id, user_id, text) VALUES (?, ?, ?)", comment.ThreadID, comment.UserID, comment.Text)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get ID of newly inserted comment
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch the created_at timestamp
    var createdAt time.Time
    err = db.QueryRow("SELECT created_at FROM comments WHERE id = ?", id).Scan(&createdAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch created_at"})
        return
    }

	// Return the inserted comment
    c.JSON(http.StatusOK, gin.H{
        "id":         id,
        "thread_id":  comment.ThreadID,
        "user_id":    comment.UserID,
        "text":       comment.Text,
        "created_at": createdAt,
    })
}

// Comment listing endpoint
func ListComments(c *gin.Context, db *sql.DB) {
	// Parse thread ID from URL
	threadID, err := strconv.Atoi(c.DefaultQuery("threadID", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse optional limit query parameter from URL
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		// Set limit to 100 if not provided
		limit = 100
	}

	// Parse last comment ID query parameter from URL. This is used to get comments after a certain comment.
	lastCommentID, err := strconv.Atoi(c.DefaultQuery("lastCommentID", "0"))
	if err != nil {
		// Set last comment ID to 0 if not provided
		lastCommentID = 0
	}

	// Query database for comment
	rows, err := db.Query("SELECT m.id, thread_id, user_id, u.username AS user_name, m.text, m.created_at FROM comments m LEFT JOIN users u ON u.id = m.user_id WHERE thread_id = ? AND m.id > ? ORDER BY m.id ASC LIMIT ?", threadID, lastCommentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create slice of comments
	var comments []Comment

	// Iterate over rows
	for rows.Next() {
		// Create new comment
		var comment Comment

		// Scan row into comment
		err := rows.Scan(&comment.ID, &comment.ThreadID, &comment.UserID, &comment.UserName, &comment.Text, &comment.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Append comment to slice
		comments = append(comments, comment)
	}

	// Return slice of comments
	c.JSON(http.StatusOK, comments)
}

// Delete a comment by ID
func DeleteComment(c *gin.Context, db *sql.DB) {
	// Parse the comment ID from the URL parameter
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	// Execute SQL to delete the comment
	result, err := db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify deletion"})
		return
	}

	// If no rows were affected, the comment does not exist
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// Update a comment by ID
func UpdateComment(c *gin.Context, db *sql.DB) {
    // Parse the comment ID from the URL parameter
    commentID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
        return
    }

    // Parse the new comment text from the request body
    var input struct {
        Text string `json:"text"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // Ensure the comment text is not empty
    if input.Text == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Comment text cannot be empty"})
        return
    }

    // Execute SQL to update the comment
    result, err := db.Exec("UPDATE comments SET text = ? WHERE id = ?", input.Text, commentID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
        return
    }

    // Check if any rows were affected
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify update"})
        return
    }

    // If no rows were affected, the comment does not exist
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
        return
    }

    // Return success message
    c.JSON(http.StatusOK, gin.H{"message": "Comment updated successfully"})
}