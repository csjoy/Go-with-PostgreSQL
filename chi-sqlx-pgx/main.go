package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type Article struct {
	ArticleID     int64     `json:"id"`
	PublishedYear string    `json:"year"`
	Title         string    `json:"title"`
	Contents      string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
}

var createTable = `
	CREATE TABLE IF NOT EXISTS articles(
		article_id serial primary key,
		published_year varchar(4),
		title text,
		contents text,
		created_at timestamp default now()
	);
`

var DBClient *sqlx.DB

func getArticle(date, slug string) (string, error) {
	var article string
	err := DBClient.Get(&article, `SELECT contents FROM articles WHERE published_year=$1 AND title=$2`, date, slug)
	if err != nil {
		return "", err
	}
	return article, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	DBClient = sqlx.MustConnect("pgx", os.Getenv("DATABASE_URL"))

	DBClient.MustExec(createTable)
	DBClient.MustExec(`INSERT INTO articles(published_year, title, contents) VALUES($1, $2, $3)`, "2023", "Mousepad", "You should fall in love with a girl.")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Get("/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		fmt.Println(userID)
	})

	r.Get("/admin/*", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "*")
		fmt.Println(slug)
	})

	apiRouter := chi.NewRouter()

	apiRouter.Get("/articles/{date}-{slug}", func(w http.ResponseWriter, r *http.Request) {
		dataParam := chi.URLParam(r, "date")
		slugParam := chi.URLParam(r, "slug")
		article, err := getArticle(dataParam, slugParam)
		if err != nil {
			w.WriteHeader(422)
			w.Write([]byte(fmt.Sprintf("error fetching article %s-%s: %v", dataParam, slugParam, err)))
			return
		}
		if article == "" {
			w.WriteHeader(404)
			w.Write([]byte("article not found"))
			return
		}
		w.Write([]byte(article))
	})
	r.Mount("/api", apiRouter)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("route does not exist"))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("method is not valid"))
	})

	http.ListenAndServe(":3000", r)
}
