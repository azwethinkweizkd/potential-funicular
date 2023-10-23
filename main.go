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
		principal, _ := strconv.ParseFloat(r.PostFormValue("purchasePrice"), 64) 
		lengthOfMortgageInMonths, _ := strconv.ParseFloat(r.PostFormValue("mortgageTerm"), 64) 
		downPayment, _ := strconv.ParseFloat(r.PostFormValue("downPayment"), 64)  
		annualTaxes, _ := strconv.ParseFloat(r.PostFormValue("annualTaxes"), 64)  
		interestRate, _ := strconv.ParseFloat(r.PostFormValue("interestRate"), 64) 
		annualInsurance, _ := strconv.ParseFloat(r.PostFormValue("annualInsurance"), 64) 
		monthlyHoa,  _ := strconv.ParseFloat(r.PostFormValue("monthlyHoa"), 64) 
		
		//Maths
		principalMinusDownPayment := principal - downPayment
		monthlyPrincipal := principalMinusDownPayment / lengthOfMortgageInMonths
		if math.IsNaN(monthlyPrincipal) {
    		monthlyPrincipal = 0
		}
		
		monthlyTaxes := (principal/12) * (annualTaxes/100)
		monthlyInsurance := annualInsurance/12
		monthlyInterestRate := (interestRate/100)/12
		monthlyInterestPayment := monthlyPrincipal * monthlyInterestRate
		monthlyPrincipalPlusInterest := monthlyPrincipal + monthlyInterestPayment
		plusOneMonthlyInterestRate := 1 + monthlyInterestRate
		exponentialByMortgageLength := math.Pow(plusOneMonthlyInterestRate, lengthOfMortgageInMonths)

		numerator := monthlyInterestRate * exponentialByMortgageLength
		denominator := exponentialByMortgageLength - 1
		division :=(principalMinusDownPayment*numerator)/denominator

		monthlyMortgagePayment := division + monthlyInsurance + monthlyHoa + monthlyTaxes

		if math.IsNaN(monthlyMortgagePayment) {
    		monthlyMortgagePayment = 0
		}

		response := fmt.Sprintf(`
				<h3 class="text-white text-6xl text-center m-auto mb-2" >
					$%.2f
				</h3>
				<div class="flex-grow">
					<div class="grid grid-cols-2 gap-y-6">
            			<p class="text-xl">Principal & Interest</p>
            			<p class="text-right text-xl">$%.2f</p>
            			<p class="text-xl">Monthly Taxes</p>
            			<p class="text-right text-xl">$%.2f</p>
            			<p class="text-xl">Monthly Insurance</p>
            			<p class="text-right text-xl">$%.2f</p>
            			<p class="text-xl">HOA</p>
            			<p class="text-right text-xl">$%.2f</p>
						<p class="text-center text-sm italic col-span-2">
						Lorem ipsum dolor sit amet consectetur, adipisicing elit. Officia id aperiam quasi in deleniti voluptate rerum suscipit neque tenetur dignissimos voluptatibus doloremque beatae commodi corrupti, mollitia repudiandae blanditiis perferendis voluptas asperiores pariatur aliquam impedit cumque itaque? Distinctio perspiciatis sed voluptas!
					  </p>
					</div>
				</div>`, monthlyMortgagePayment, monthlyPrincipalPlusInterest, monthlyTaxes, monthlyInsurance, monthlyHoa)

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


