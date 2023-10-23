package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

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
		floatPrincipal, _ := strconv.ParseFloat(principal, 64)  
		lengthOfMortgageInMonths := r.PostFormValue("mortgageTerm")
		floatMortgageLength, _ := strconv.ParseFloat(lengthOfMortgageInMonths, 64)  
		downPayment := r.PostFormValue("downPayment")
		floatDownPayment, _ := strconv.ParseFloat(downPayment, 64)  
		annualTaxes := r.PostFormValue("annualTaxes")
		floatAnnualTaxes, _ := strconv.ParseFloat(annualTaxes, 64)  
		interestRate := r.PostFormValue("interestRate")
		floatInterestRate, _ := strconv.ParseFloat(interestRate, 64)  
		annualInsurance := r.PostFormValue("annualInsurance")
		floatAnnualInsurance, _ := strconv.ParseFloat(annualInsurance, 64)  
		monthlyHoa := r.PostFormValue("monthlyHoa")
		floatMonthlyHoa, _ := strconv.ParseFloat(monthlyHoa, 64)
		
		//Maths
		principalMinusDownPayment := floatPrincipal - floatDownPayment
		monthlyTaxes := (floatPrincipal/12) * (floatAnnualTaxes/100)
		monthlyInsurance := floatAnnualInsurance/12
		monthlyInterestRate := (floatInterestRate/100)/12
		plusOneMonthlyInterestRate := 1 + monthlyInterestRate
		exponentialByMortgageLength := math.Pow(plusOneMonthlyInterestRate, floatMortgageLength)

		numerator := monthlyInterestRate * exponentialByMortgageLength
		denominator := exponentialByMortgageLength - 1
		division :=(principalMinusDownPayment*numerator)/denominator

		monthlyMortgagePayment := division + monthlyInsurance + floatMonthlyHoa + monthlyTaxes

		
response := fmt.Sprintf(`
<div class="col-span-1">
	<h3 class="text-white text-6xl text-center m-auto" id="monthlyPayment">
		$%.2f
	</h3>
</div>`, monthlyMortgagePayment)

fmt.Fprint(w, response)
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


