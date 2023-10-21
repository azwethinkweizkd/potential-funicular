package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type SaveUserMortgage struct {
	ID        	 int      `json:"id"`
	principal 	 string   `json:"principal"`
	mortgageTerm string   `json:"mortgageTerm"`
}

func init() {
    // loads values from .env into the system
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func main() {
	fmt.Println("Mortgage Calulator")
	dbPassword := os.Getenv("DB_PASSWORD")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/css/", twhandler.New(http.Dir("css"), "/css", twembed.New()))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins in development
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	s := &http.Server{Addr: ":8001", Handler: c.Handler(mux)}
	
	h1 := func(w http.ResponseWriter, r *http.Request) {
		principal := r.PostFormValue("purchasePrice")
		lengthOfMortgageInMonths := r.PostFormValue("mortgageTerm")

		log.Println(principal, lengthOfMortgageInMonths)
	}

	h2 := func(w http.ResponseWriter, r *http.Request) { 
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		test := r.PostFormValue("test")
		log.Println(test)
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, test)
	}
	http.HandleFunc("/test", h2)
	http.HandleFunc("/postMonthlyPayment", h1)

	
    db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/saved_user_mortgagedb", dbPassword))
	if err != nil {
        panic(err.Error())
    }
    defer db.Close()
    //insert, err := db.Query("INSERT INTO testtable2 VALUES('24')")
    if err !=nil {
        panic(err.Error())
    }
    //defer insert.Close()
	fmt.Println("Now listening on: http://localhost:8001")
	log.Fatal(s.ListenAndServe())
}


