package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type MonthlyPayment struct {
	MontlyPayment  string
}

func init() {
    // loads values from .env into the system
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func Atoi (s string) int {
    var (
        n uint64
        i int
        v byte
    )   
    for ; i < len(s); i++ {
        d := s[i]
        if '0' <= d && d <= '9' {
            v = d - '0'
        } else if 'a' <= d && d <= 'z' {
            v = d - 'a' + 10
        } else if 'A' <= d && d <= 'Z' {
            v = d - 'A' + 10
        } else {
            n = 0; break        
        }
        n *= uint64(10) 
        n += uint64(v)
    }
    return int(n)
}


func Atof(s string) float64 {
	s = strings.TrimSpace(s) // Remove leading/trailing spaces

    if s == "" {
        return 0.00
    }

    var result float64
    var decimalPlace float64

    for _, ch := range s {
        if ch == '.' {
            decimalPlace = 1.0
        } else if ch < '0' || ch > '9' {
            return 0.00 // Invalid character in the string
        } else if decimalPlace == 0 {
            digit := float64(ch - '0')
            result = result*10 + digit
        } else {
            digit := float64(ch - '0')
            result = result + digit/decimalPlace
            decimalPlace *= 10
        }
    }

    return result
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
		principal := Atof(r.PostFormValue("purchasePrice"))
		lengthOfMortgageInMonths := Atoi(r.PostFormValue("mortgageTerm"))
		downPayment := Atof(r.PostFormValue("downPayment"))
		annualTaxes := Atof(r.PostFormValue("annualTaxes"))
		interestRate := Atof(r.PostFormValue("interestRate"))
		annualInsurance := Atof(r.PostFormValue("annualInsurance"))
		monthlyHoa := Atof(r.PostFormValue("monthlyHoa"))


		fmt.Printf("Principal: %.2f, Length: %d, Downpayment: %.2f, Annual Taxes: %.2f, Interest Rate: %.2f, Annual Insurance: %.2f, Monthly HOA: %.2f \n", principal, lengthOfMortgageInMonths, downPayment, annualTaxes,interestRate, annualInsurance, monthlyHoa)
	}

	mux.HandleFunc("/postMonthlyPayment", h1)

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


