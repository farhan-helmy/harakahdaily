package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"strings"

	// import Colly

	"github.com/gocolly/colly"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	Scrape()
}

func Scrape() {
	db, err := sql.Open("sqlite3", "./harakahdaily.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := colly.NewCollector()

	c.OnHTML("article", func(e *colly.HTMLElement) {
		title := e.ChildText("h1")
		fmt.Printf("Article Title: %s\n", title)

		paragraphs := []string{}
		e.ForEach("p", func(_ int, p *colly.HTMLElement) {
			paragraphText := strings.TrimSpace(p.Text)
			if paragraphText != "" {
				paragraphs = append(paragraphs, paragraphText)
			}
		})

		fmt.Print(paragraphs)

		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		stmt, err := tx.Prepare("insert into harakahdaily(title, article, articleUrl, date) values(?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(title, strings.Join(paragraphs, "\n"), e.Request.URL.String(), e.Request.URL.Path[1:11])
		if err != nil {
			log.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, e.Request.URL.Path) && !strings.ContainsAny(link, "?#%") && !strings.Contains(link, "page") {
			fmt.Printf("Link found: %s\n", link)
			e.Request.Visit(link)
		}
	})

	// Get current date
	currentDate := time.Now()

	// Calculate the start of last month
	startOfLastMonth := currentDate.AddDate(0, -1, 0).AddDate(0, 0, -currentDate.AddDate(0, -1, 0).Day()+1)

	// Loop backwards until start of last month
	for d := currentDate; d.After(startOfLastMonth) || d.Equal(startOfLastMonth); d = d.AddDate(0, 0, -1) {
		dateString := d.Format("2006/01/02")
		url := fmt.Sprintf("https://harakahdaily.net/index.php/%s/", dateString)
		fmt.Printf("Visiting: %s\n", url)
		c.Visit(url)
	}

}
