package handlers

import (
    "database/sql"
	"log"
	"net/http"
	"strings"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/lib/pq"
)

type Thread struct {
	ID int `json:"id"`
	Name string `json:"name"`
	UserID int `json:"user_id"`
	Tags []string `json:"tags"`
}

// Thread creation endpoint
func CreateThread(c *gin.Context, db *sql.DB) {
	// Predefined list of allowed tags
    predefinedTags := map[string]bool{
        "School":             true,
        "Work":               true,
        "Interests and Hobbies": true,
        "Miscellaneous":      true,
    }

	// Parse JSON request body into Channel struct
	var thread Thread
	if err := c.ShouldBindJSON(&thread); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert thread into database with RETURNING id
	var threadID int
	err := db.QueryRow("INSERT INTO threads (name, user_id) VALUES ($1, $2) RETURNING id", thread.Name, thread.UserID).Scan(&threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If there are tags, associate them with the thread
	var savedTags []string
	if len(thread.Tags) > 0 {
		for _, tag := range thread.Tags {
			// Validate the tag
			if !predefinedTags[tag] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag: " + tag})
				return
			}

			// Get the tag ID from the database
			var tagID int
			err := db.QueryRow("SELECT id FROM tags WHERE name = $1", tag).Scan(&tagID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Tag not found: " + tag})
				return
			}

			// Associate the tag with the thread in the thread_tags table
			log.Printf("Associating thread ID %d with tag ID %d", threadID, tagID)
			_, err = db.Exec("INSERT INTO thread_tags (thread_id, tag_id) VALUES ($1, $2)", threadID, tagID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Add the tag to the savedTags slice
			savedTags = append(savedTags, tag)
		}
	}

	// Return the added thread
    c.JSON(http.StatusOK, gin.H{
        "id":         threadID,
        "name":       thread.Name,
        "user_id":    thread.UserID,
		"tags":       savedTags,
    })
}

// Thread listing endpoint
func ListThreads(c *gin.Context, db *sql.DB) {
	// Query to get threads along with their associated tags
	query := `
	SELECT threads.id, threads.name, threads.user_id, string_agg(tags.name, ', ') AS tags
	FROM threads
	LEFT JOIN thread_tags ON threads.id = thread_tags.thread_id
	LEFT JOIN tags ON thread_tags.tag_id = tags.id
	GROUP BY threads.id
	`

	// Query database for threads
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create slice of threads
	var threads []Thread

	// Iterate over rows
	for rows.Next() {
		// Create new thread
		var thread Thread
		var tags sql.NullString // Use sql.NullString to handle NULL values

		// Scan row into thread
		err := rows.Scan(&thread.ID, &thread.Name, &thread.UserID, &tags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Check if tags is null, and if not, split the concatenated tags
        if tags.Valid && tags.String != "" {
            thread.Tags = strings.Split(tags.String, ",")
        } else {
            // If tags are null or empty, return an empty array
            thread.Tags = []string{}
        }

		// Append thread to slice
		threads = append(threads, thread)
	}

	// Return slice of threads
	c.JSON(http.StatusOK, threads)
}

// Delete a thread by ID
func DeleteThread(c *gin.Context, db *sql.DB) {
	// Parse the thread ID from the URL parameter
	threadID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Execute SQL to delete comments associated with the thread
	_, err = db.Exec("DELETE FROM comments WHERE thread_id = $1", threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated comments"})
		return
	}

	// Execute SQL to delete the thread
	result, err := db.Exec("DELETE FROM threads WHERE id = $1", threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify deletion"})
		return
	}

	// If no rows were affected, the thread does not exist
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted successfully"})
}

// Update a thread name and/or tags by ID
func UpdateThread(c *gin.Context, db *sql.DB) {
	// Predefined list of allowed tags
	predefinedTags := map[string]bool{
		"School":             true,
		"Work":               true,
		"Interests and Hobbies": true,
		"Miscellaneous":      true,
	}

    // Parse the thread ID from the URL parameter
    threadID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
        return
    }

    // Parse the request body
    var input struct {
        Name string `json:"name"`
		Tags []string `json:"tags"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

	// Start a transaction
    tx, err := db.Begin()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
        return
    }
    defer tx.Rollback()

    // Ensure the thread name is not empty
    if input.Name == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Thread name cannot be empty"})
        return
    }

    // Execute SQL to update the thread name
    result, err := db.Exec("UPDATE threads SET name = $1 WHERE id = $2", input.Name, threadID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread name"})
        return
    }

    // Check if any rows were affected
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify update"})
        return
    }

    // If no rows were affected, the thread does not exist
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found"})
        return
    }

	// Delete existing tags for the thread
    _, err = tx.Exec("DELETE FROM thread_tags WHERE thread_id = $1", threadID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing tags"})
        return
    }

	// Update the thread tags if provided
    if len(input.Tags) > 0 {
        // Validate tags
        for _, tag := range input.Tags {
            if !predefinedTags[tag] {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag: " + tag})
                return
            }
        }

		// Insert new tags
        for _, tag := range input.Tags {
            var tagID int
            err := tx.QueryRow("SELECT id FROM tags WHERE name = $1", tag).Scan(&tagID)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Tag not found in database: " + tag})
                return
            }

            _, err = tx.Exec("INSERT INTO thread_tags (thread_id, tag_id) VALUES ($1, $2)", threadID, tagID)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tags"})
                return
            }
        }
    }

	// Commit the transaction
    if err := tx.Commit(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    // Return success message
    c.JSON(http.StatusOK, gin.H{"message": "Thread updated successfully"})
}

// Tags listing endpoint
func ListTags(c *gin.Context, db *sql.DB) {
	// Query the database to get all tags
	rows, err := db.Query("SELECT name FROM tags")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tags"})
		return
	}
	defer rows.Close()

	// Create a slice to hold the tags
	var tags []string

	// Iterate over the rows to extract tag names
	for rows.Next() {
		var tagName string
		if err := rows.Scan(&tagName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan tags"})
			return
		}
		tags = append(tags, tagName)
	}

	// Return the list of tags
	c.JSON(http.StatusOK, tags)
}

// Threads listing by tags endpoint
func GetThreadsByTags(c *gin.Context, db *sql.DB) {
    // Get tags from query parameters
    tagsParam := c.Query("tags")
    if tagsParam == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Tags parameter is required"})
        return
    }

    // Split the tags into a slice
    tags := strings.Split(tagsParam, ",")
    placeholders := make([]string, len(tags))
    for i := range tags {
        placeholders[i] = "$" + strconv.Itoa(i+1)
    }

    // Create a query to get threads filtered by tags
    query := `
    SELECT threads.id, threads.name, threads.user_id, string_agg(tags.name, ', ') AS tags
    FROM threads
    LEFT JOIN thread_tags ON threads.id = thread_tags.thread_id
    LEFT JOIN tags ON thread_tags.tag_id = tags.id
    WHERE threads.id IN (
        SELECT thread_tags.thread_id
        FROM thread_tags
        LEFT JOIN tags ON thread_tags.tag_id = tags.id
        WHERE tags.name IN (` + strings.Join(placeholders, ", ") + `)
        GROUP BY thread_tags.thread_id
        HAVING COUNT(DISTINCT tags.name) = $` + strconv.Itoa(len(tags)+1) + `
    )
    GROUP BY threads.id
    `

    // Convert tags to []interface{} for use with Query
    queryArgs := make([]interface{}, len(tags) + 1)
    for i, tag := range tags {
        queryArgs[i] = tag
    }
	queryArgs[len(tags)] = len(tags)

    // Execute the query
    //args := append(tags, strconv.Itoa(len(tags)))
    rows, err := db.Query(query, queryArgs...)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
	defer rows.Close()

    // Create a slice to store the threads
    var threads []Thread

    // Iterate over rows
    for rows.Next() {
        var thread Thread
        var tags sql.NullString

        // Scan the row into thread and tags
        if err := rows.Scan(&thread.ID, &thread.Name, &thread.UserID, &tags); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // If tags are valid, split them into a slice
        if tags.Valid && tags.String != "" {
            thread.Tags = strings.Split(tags.String, ",")
        } else {
            thread.Tags = []string{}
        }

        // Append the thread to the slice
        threads = append(threads, thread)
    }

    // Return the filtered threads
    c.JSON(http.StatusOK, threads)
}