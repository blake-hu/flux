package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pgvector/pgvector-go"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var schema = `
CREATE TABLE IF NOT EXISTS person (
	first_name text,
	last_name text,
	email text,
	embedding vector(3)
);`

type Person struct {
	FirstName string          `db:"first_name"`
	LastName  string          `db:"last_name"`
	Email     string          `db:"email"`
	Embedding pgvector.Vector `db:"embedding"`
}

const db_user = "user"
const db_password = "password"
const db_name = "mydb"
const port = "8080"

type App struct {
	db *sqlx.DB
}

func main() {
	db, err := sqlx.Connect("postgres", fmt.Sprintf("host=db user=%s dbname=%s password=%s sslmode=disable", db_user, db_name, db_password))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s", err.Error())
	}
	defer db.Close()

	app := &App{db: db}
	db.MustExec("CREATE EXTENSION IF NOT EXISTS vector;")
	db.MustExec(schema)

	server := http.Server{
		Addr: ":" + port,
	}
	http.HandleFunc("/", app.indexHandler)

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Failed to start server")
		os.Exit(1)
	}

	fmt.Printf("Sever listening on port %s\n", port)
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("hit index handler\n")
	people := []Person{
		Person{FirstName: "Max", LastName: "Glass", Email: "glass@u.northwestern.edu", Embedding: pgvector.NewVector([]float32{1, 1, 1})},
		Person{FirstName: "Blake", LastName: "Hu", Email: "email", Embedding: pgvector.NewVector([]float32{2, 2, 2})},
	}
	_, err := app.db.NamedExec("INSERT INTO person (first_name, last_name, email, embedding) VALUES (:first_name, :last_name, :email, :embedding)", people)
	if err != nil {
		log.Fatalf("failed to insert %s\n", err.Error())
	}

	var selected_people []Person
	app.db.Select(&selected_people, "SELECT * FROM person ORDER BY embedding <-> $1", pgvector.NewVector([]float32{1, 1, 1}))
	fmt.Printf("people: %+v\n", selected_people)
}
