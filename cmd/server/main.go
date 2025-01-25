package main

import (
    "database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	"github.com/gin-contrib/cors"
	"github.com/CVWO/sample-go-app/internal/handlers"
)

func main() {
	// Set Gin to Release Mode
    gin.SetMode(gin.ReleaseMode)
	
	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// Print the working directory
	fmt.Println("Working directory:", wd)

	// Open the SQLite database file
	db, err := sql.Open("sqlite", wd+"/internal/database/database.db")

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	// Create the Gin router
	r := gin.Default()

	if err != nil {
		log.Fatal(err)
	}

	// Enable CORS
    r.Use(cors.Default())

	// Creation endpoints
	r.POST("/users", func(c *gin.Context) { handlers.CreateUser(c, db) })
	r.POST("/threads", func(c *gin.Context) { handlers.CreateThread(c, db) })
	r.POST("/comments", func(c *gin.Context) { handlers.CreateComment(c, db) })

	// Listing endpoints
	r.GET("/threads", func(c *gin.Context) {
		if tagsQuery := c.Query("tags"); tagsQuery != "" {
			handlers.GetThreadsByTags(c, db) // Handle filtered query
		} else {
			handlers.ListThreads(c, db) // Handle unfiltered query
		}
	})
	r.GET("/comments", func(c *gin.Context) { handlers.ListComments(c, db) })
	r.GET("/tags", func(c *gin.Context) { handlers.ListTags(c, db) })

	// Login endpoint
	r.POST("/login", func(c *gin.Context) { handlers.Login(c, db) })

	// Deletion endpoints
	r.DELETE("/comments/:id", func(c *gin.Context) { handlers.DeleteComment(c, db) })
	r.DELETE("/threads/:id", func(c *gin.Context) { handlers.DeleteThread(c, db) })

	// Update endpoints
	r.PATCH("/comments/:id", func(c *gin.Context) { handlers.UpdateComment(c, db) })
	r.PATCH("/threads/:id", func(c *gin.Context) { handlers.UpdateThread(c, db) })

	// Bind to the port specified by the PORT environment variable
    port := os.Getenv("PORT")
    if port == "" {
        port = "10000"  // Default Render port
    }

    fmt.Println("Using port:", port)
    if err := r.Run("0.0.0.0:" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}