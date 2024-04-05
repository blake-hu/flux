package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var schema = `
CREATE TABLE IF NOT EXISTS person (
	first_name text,
	last_name text,
	email text
);`

type Person struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
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
	app.db.Exec("INSERT INTO person (first_name, last_name, email) VALUES ($1, $2, $3)", "Max", "Glass", "glass@u.northwestern.edu")
	fmt.Printf("inserted into db\n")
	people := []Person{}
	app.db.Select(&people, "SELECT * FROM person")
	fmt.Printf("people: %+v\n", people)
}
