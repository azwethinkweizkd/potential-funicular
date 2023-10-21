package main
import (
    "fmt"
	"os"
	"log"
	"net/http"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func init() {
    // loads values from .env into the system
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func main() {
	fmt.Println("Mortgage Calulator")
	dbPassword, exists := os.LookupEnv("DB_PASSWORD")

    if exists {
		fmt.Println(dbPassword)
    }

	h1 := func(w http.ResponseWriter, r *http.Request) {

	}

	// h2 := func(w http.ResponseWriter, r *http.Request) {

	// }

	// define handlers
	http.HandleFunc("/", h1)

	
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

	log.Fatal(http.ListenAndServe(":3000", nil))
}

