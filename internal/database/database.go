package database

import (
    "database/sql"
    "fmt"
)

// Create the database file and initialize the necessary tables and seed data
func InitializeDatabase(db *sql.DB) error {
    // Create tables
    tableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS threads (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        user_id INT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS comments (
        id SERIAL PRIMARY KEY,
        thread_id INT NOT NULL,
        user_id INT NOT NULL,
        text TEXT NOT NULL,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (thread_id) REFERENCES threads(id),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS tags (
        id SERIAL PRIMARY KEY,
        name TEXT UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS thread_tags (
        thread_id INT REFERENCES threads(id) ON DELETE CASCADE,
        tag_id INT REFERENCES tags(id) ON DELETE CASCADE,
        PRIMARY KEY (thread_id, tag_id)
    );
    `

    _, err := db.Exec(tableSQL)
    if err != nil {
        return fmt.Errorf("failed to create tables: %v", err)
    }

    // Seed tags
    seedTagsSQL := `
    INSERT INTO tags (name) VALUES
    ('School'),
    ('Work'),
    ('Interests and Hobbies'),
    ('Miscellaneous')
    ON CONFLICT (name) DO NOTHING;
    `
    _, err = db.Exec(seedTagsSQL)
    if err != nil {
        return fmt.Errorf("failed to seed tags: %v", err)
    }

    fmt.Println("Database initialized successfully.")
    return nil
}