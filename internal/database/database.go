package database

import (
    "database/sql"
    "fmt"
)

// Create the database file and initialize the necessary tables and seed data
func InitializeDatabase(dbPath string) error {
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return fmt.Errorf("failed to open database: %v", err)
    }
    defer db.Close()

    // Create tables
    tableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        username TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS threads (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        user_id INT DEFAULT 0 NOT NULL
    );

    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        thread_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        text TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (thread_id) REFERENCES threads(id),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS thread_tags (
        thread_id INT REFERENCES threads(id) ON DELETE CASCADE,
        tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
        PRIMARY KEY (thread_id, tag_id)
    );
    `
    _, err = db.Exec(tableSQL)
    if err != nil {
        return fmt.Errorf("failed to create tables: %v", err)
    }

    // Seed tags
    seedTagsSQL := `
    INSERT OR IGNORE INTO tags (name) VALUES
    ('School'),
    ('Work'),
    ('Interests and Hobbies'),
    ('Miscellaneous');
    `
    _, err = db.Exec(seedTagsSQL)
    if err != nil {
        return fmt.Errorf("failed to seed tags: %v", err)
    }

    fmt.Println("Database initialized successfully.")
    return nil
}