package main

import (
	"database/sql"
	"errors"
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

func division(n, d float64) (float64, error) {
	if d == 0 {
		return 0, errors.New("Can't divide by zero")
	}
	div := n / d
	return div, nil
}

func multiply(n, d float64) float64 {
	return n * d
}

func getLoanDescription(w http.ResponseWriter, r *http.Request) {
    loanType := r.URL.Query().Get("loanType")
	dbPassword := os.Getenv("DB_PASSWORD")
    db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/saved_user_mortgagedb", dbPassword))
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer db.Close()

    query := "SELECT description FROM loans WHERE loan_type = ?"
    var description string

    err = db.QueryRow(query, loanType).Scan(&description)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    fmt.Fprint(w, description)
}

func postMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	principal, _ := strconv.ParseFloat(r.PostFormValue("purchasePrice"), 64) 
	lengthOfMortgageInMonths, _ := strconv.ParseFloat(r.PostFormValue("mortgageTerm"), 64) 
	downPayment, _ := strconv.ParseFloat(r.PostFormValue("downPayment"), 64)  
	annualTaxes, _ := strconv.ParseFloat(r.PostFormValue("annualTaxes"), 64)  
	interestRate, _ := strconv.ParseFloat(r.PostFormValue("interestRate"), 64) 
	annualInsurance, _ := strconv.ParseFloat(r.PostFormValue("annualInsurance"), 64) 
	monthlyHoa,  _ := strconv.ParseFloat(r.PostFormValue("monthlyHoa"), 64) 
	
	//Maths
	principalMinusDownPayment := principal - downPayment
	monthlyPrincipal, _ := division(principalMinusDownPayment, lengthOfMortgageInMonths)
	taxPercent, _ := division(annualTaxes, 100)
	yearlyTaxes := multiply(principal, taxPercent)
	monthlyTaxes, _ := division(yearlyTaxes, 12)
	monthlyInsurance, _ := division(annualInsurance, 12)
	interestRatePercentage, _ := division(interestRate, 100)
	monthlyInterestRate, _ := division(interestRatePercentage, 12)
	monthlyInterestPayment := multiply(monthlyPrincipal, monthlyInterestRate)
	monthlyPrincipalPlusInterest := monthlyPrincipal + monthlyInterestPayment
	plusOneMonthlyInterestRate := 1 + monthlyInterestRate
	exponentialByMortgageLength := math.Pow(plusOneMonthlyInterestRate, lengthOfMortgageInMonths)

	numerator1 := multiply(monthlyInterestRate, exponentialByMortgageLength)
	numerator2 := multiply(principalMinusDownPayment, numerator1)
	denominator := exponentialByMortgageLength - 1
	division, _ := division(numerator2,denominator)

	monthlyMortgagePayment := division + monthlyInsurance + monthlyHoa + monthlyTaxes

	if math.IsNaN(monthlyMortgagePayment) {
		monthlyMortgagePayment = 0
	}

	response := fmt.Sprintf(`
			<h3 class="text-white text-6xl text-center m-auto mb-8" >
				$%.2f
			</h3>
			<div class="flex-grow">
				<div class="grid grid-cols-2 gap-y-8">
					<p class="text-xl">Principal & Interest</p>
					<p class="text-right text-xl">$%.2f</p>
					<p class="text-xl">Monthly Taxes</p>
					<p class="text-right text-xl">$%.2f</p>
					<p class="text-xl">Monthly Insurance</p>
					<p class="text-right text-xl">$%.2f</p>
					<p class="text-xl">HOA</p>
					<p class="text-right text-xl">$%.2f</p>
					<p class="text-center text-sm italic col-span-2">
					Please note that the mortgage calculator on our website provides estimates for general informational purposes only. For personalized guidance and accurate loan information, we recommend reaching out to our expert loan officers who can assist you with your specific mortgage needs.
				  </p>
				</div>
			</div>`, monthlyMortgagePayment, monthlyPrincipalPlusInterest, monthlyTaxes, monthlyInsurance, monthlyHoa)

	fmt.Fprint(w, response)
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

	mux.HandleFunc("/getLoanDescription", getLoanDescription)
	mux.HandleFunc("/postMonthlyPayment", postMonthlyPayment)

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