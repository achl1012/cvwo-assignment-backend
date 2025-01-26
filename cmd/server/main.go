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
	"github.com/CVWO/sample-go-app/internal/database"
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
    dbPath := wd + "/internal/database/database.db"
    // Check if the file exists
    if _, err := os.Stat(dbPath); os.IsNotExist(err) {
        fmt.Println("Database file does not exist. Creating new database...")
        err = database.InitializeDatabase(dbPath)
        if err != nil {
            log.Fatalf("Failed to initialize database: %v", err)
        }
    }

    fmt.Println("Using database path:", dbPath)
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

	// Create the Gin router
	r := gin.Default()

	if err != nil {
		log.Fatal(err)
	}

   	// Enable CORS with custom configuration
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{
			"https://forumflow-frontend.onrender.com",
			"http://localhost:10001",
		},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

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