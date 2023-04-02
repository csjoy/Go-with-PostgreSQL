package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	db := sqlx.MustConnect("pgx", os.Getenv("DATABASE_URL"))

	r := gin.Default()

	r.GET("/get", func(ctx *gin.Context) {
		rows, err := db.Query("SELECT * FROM posts")
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var posts []Post
		for rows.Next() {
			var newPost Post
			err = rows.Scan(&newPost.ID, &newPost.Title, &newPost.Content, &newPost.CreatedAt)
			if err != nil {
				panic(err)
			}
			posts = append(posts, newPost)
		}
		ctx.IndentedJSON(http.StatusOK, posts)
	})

	r.GET("/getx", func(ctx *gin.Context) {
		rows, err := db.Queryx("SELECT * FROM posts")
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var posts []Post
		for rows.Next() {
			var newPost Post
			err = rows.StructScan(&newPost)
			if err != nil {
				panic(err)
			}
			posts = append(posts, newPost)
		}

		if err != nil {
			panic(err)
		}
		ctx.IndentedJSON(http.StatusOK, posts)
	})

	r.GET("/getxx", func(ctx *gin.Context) {
		var posts []Post
		err := db.Select(&posts, "SELECT * FROM posts")
		if err != nil {
			panic(err)
		}

		if err != nil {
			panic(err)
		}
		ctx.IndentedJSON(http.StatusOK, posts)
	})

	r.POST("/", func(ctx *gin.Context) {
		var newPost Post
		if err := ctx.BindJSON(&newPost); err != nil {
			log.Fatal("Error binding json", err)
		}
		res := db.MustExec(`INSERT INTO posts(title, content, created_at) VALUES($1, $2, $3)`, newPost.Title, newPost.Content, time.Now())
		fmt.Println(res.RowsAffected())

		ctx.IndentedJSON(http.StatusCreated, newPost)

	})

	if err := r.Run(":5000"); err != nil {
		log.Fatal("Error initilizing server", err)
	}
}
