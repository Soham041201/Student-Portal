package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := "postgres"

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	fmt.Println(connectionString)
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}
	router := gin.Default()
	router.Use(dbMiddleware(db))
	router.GET("/albums", getAlbumbs)
	router.POST("/album", createAlbum)
	router.Run(":8080")
}

func dbMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func getAlbumbs(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, _ := db.Query("SELECT * FROM albums")
	results := []album{}

	for rows.Next() {
		var id string
		var title string
		var artist string
		var price float64

		err := rows.Scan(&id, &title, &artist, &price)
		if err != nil {
			panic(err)
		}
		results = append(results, album{id, title, artist, price})
	}

	c.IndentedJSON(http.StatusOK, results)
}

func createAlbum(c *gin.Context) {
	var newAlbum album
	db := c.MustGet("db").(*sql.DB)

	_, dbErr := db.Exec(`CREATE TABLE IF NOT EXISTS albums (
		id VARCHAR(255) PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		price DECIMAL(10,2) NOT NULL
	  )`)

	if dbErr != nil {
		c.AbortWithError(http.StatusInternalServerError, dbErr)
		return
	}

	insertQuery := `INSERT INTO albums(id,title,artist,price) VALUES ($1,$2,$3,$4)`

	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	albums = append(albums, newAlbum)

	_, err := db.Exec(insertQuery, newAlbum.ID, newAlbum.Title, newAlbum.Artist, newAlbum.Price)

	if err != nil {
		fmt.Print(err)
		c.IndentedJSON(http.StatusInternalServerError, err)
	}
}
