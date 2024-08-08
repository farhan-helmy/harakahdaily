package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type Article struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Article    string `json:"article"`
	ArticleUrl string `json:"articleUrl"`
}

func main() {

	// Set up logging to a file
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	// Create a new logger that writes to the file
	fileLogger := log.New(logFile, "", log.LstdFlags)

	e := echo.New()

	// Add logging middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: logFile,
	}))

	e.GET("/harakahdaily", getHarakahDailyArticles(fileLogger))

	e.Logger.Fatal(e.Start(":1323"))
}

func getHarakahDailyArticles(fileLogger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		db, err := sql.Open("sqlite3", "./harakahdaily.db")
		if err != nil {
			fileLogger.Printf("Database connection failed: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
		}
		defer db.Close()

		rows, err := db.Query("SELECT id, title, article, articleUrl FROM harakahdaily")
		if err != nil {
			fileLogger.Printf("Query execution failed: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Query execution failed"})
		}
		defer rows.Close()

		var articles []Article
		for rows.Next() {
			var article Article
			err = rows.Scan(&article.ID, &article.Title, &article.Article, &article.ArticleUrl)
			if err != nil {
				fileLogger.Printf("Error scanning row: %v", err)
				continue
			}
			articles = append(articles, article)
		}

		if err = rows.Err(); err != nil {
			fileLogger.Printf("Error iterating over rows: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error iterating over rows"})
		}

		duration := time.Since(start)
		fileLogger.Printf("Request processed. IP: %s, Duration: %v, Articles returned: %d", c.RealIP(), duration, len(articles))

		return c.JSON(http.StatusOK, articles)
	}
}
