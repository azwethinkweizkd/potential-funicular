package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"
	"github.com/trycourier/courier-go/v2"
)

type SavedMortgageInfo struct {
    ID                  int
    MonthlyMortgagePayment string
    Principal            string
    MortgageTerm         string
    AnnualTaxes          string
    AnnualTaxPercentage  string
    DownPayment          string
    InterestRate         string
    AnnualInsurance      string
    MonthlyHOA           string
    Email               string
    DateCreated          time.Time
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

func sendEmail(email string, i SavedMortgageInfo) (string, error) {
	courierApiKey := os.Getenv("COURIER_API_KEY")
	courierTemplateId := os.Getenv("COURIER_TEMPLATE_ID")
	client := courier.CreateClient(courierApiKey, nil)

	request := courier.SendMessageRequestBody{
		Message: map[string]interface{}{
			"to": map[string]string{
				"email": email,
			},
			"template": courierTemplateId,
			"data": map[string]interface{}{
				"monthlyMortgagePayment": i.MonthlyMortgagePayment,
				"principal":              i.Principal,
				"interestRate":           i.InterestRate,
				"downPayment":            i.DownPayment,
				"mortgageTerm":           i.MortgageTerm,
				"annualTaxes":            i.AnnualTaxes,
				"annualInsurance":        i.AnnualInsurance,
				"monthlyHOA":             i.MonthlyHOA,
			},
		},
	}

	requestID, err := client.SendMessage(context.Background(), request)

	if err != nil {
		log.Println(err)
		return "Failed to send email", err
	}

	return requestID, nil
}

func getLoanDescription(w http.ResponseWriter, r *http.Request) {
    loanType := r.URL.Query().Get("loanType")
    dbURL := os.Getenv("JAWSDB_URL")

    db, err := sql.Open("mysql", dbURL)
    if err != nil {
        log.Fatal(err)
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

func postSendEmailAndSaveInDb(w http.ResponseWriter, r *http.Request) {
	principal, _ := strconv.ParseFloat(r.PostFormValue("purchasePrice"), 64) 
	lengthOfMortgageInMonths, _ := strconv.ParseFloat(r.PostFormValue("mortgageTerm"), 64) 
	downPayment, _ := strconv.ParseFloat(r.PostFormValue("downPayment"), 64)  
	annualTaxes, _ := strconv.ParseFloat(r.PostFormValue("annualTaxes"), 64)  
	interestRate, _ := strconv.ParseFloat(r.PostFormValue("interestRate"), 64) 
	annualInsurance, _ := strconv.ParseFloat(r.PostFormValue("annualInsurance"), 64) 
	monthlyHoa, _ := strconv.ParseFloat(r.PostFormValue("monthlyHoa"), 64) 
	email := r.PostFormValue("email")

	principalMinusDownPayment := principal - downPayment
	taxPercent, _ := division(annualTaxes, 100)
	yearlyTaxes := multiply(principal, taxPercent)
	monthlyTaxes, _ := division(yearlyTaxes, 12)
	monthlyInsurance, _ := division(annualInsurance, 12)
	interestRatePercentage, _ := division(interestRate, 100)
	monthlyInterestRate, _ := division(interestRatePercentage, 12)
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

	dbURL := os.Getenv("JAWSDB_URL")

    db, err := sql.Open("mysql", dbURL)
    if err != nil {
        log.Fatal(err)
    }

	principalStr := strconv.FormatFloat(principal, 'f', 2, 64)
	mortgageTermStr := strconv.FormatFloat(lengthOfMortgageInMonths/12, 'f', -1, 64)
	downPaymentStr := strconv.FormatFloat(downPayment, 'f', 2, 64)
	annualTaxesStr := strconv.FormatFloat(monthlyTaxes * 12, 'f', 2, 64)
	interestRateStr := strconv.FormatFloat(interestRate, 'f', 3, 64)
	annualInsuranceStr := strconv.FormatFloat(annualInsurance, 'f', 2, 64)
	monthlyHoaStr := strconv.FormatFloat(monthlyHoa, 'f', 2, 64)
	monthlyMortgagePaymentStr := strconv.FormatFloat(monthlyMortgagePayment, 'f', 2, 64)

	info := SavedMortgageInfo{
		MonthlyMortgagePayment: monthlyMortgagePaymentStr,
		Principal:            	principalStr,
		MortgageTerm:         	mortgageTermStr,
		AnnualTaxes:          	annualTaxesStr,
		DownPayment:          	downPaymentStr,
		InterestRate:         	interestRateStr,
		AnnualInsurance:      	annualInsuranceStr,
		MonthlyHOA:           	monthlyHoaStr,
		Email:               	email,
		DateCreated:          	time.Now(),
	}

	insertQuery := "INSERT INTO MortgageInfo (monthly_mortgage_payment, principal, mortgage_term, annual_taxes, down_payment, interest_rate, annual_insurance, monthly_hoa, email, date_created) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = db.Exec(insertQuery, info.MonthlyMortgagePayment, info.Principal, info.MortgageTerm, info.AnnualTaxes, info.DownPayment, info.InterestRate, info.AnnualInsurance, info.MonthlyHOA, info.Email, info.DateCreated)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	reqId, err := sendEmail(email, info)

	if err != nil {
		log.Println("Email failed to send")
		message := fmt.Sprintf(`<p id="message" class="text-red-600 text-center">
		Email failed to send &#10060</p>`)
    	fmt.Fprint(w, message)
	} else {
		log.Println(reqId)
		message := fmt.Sprintf(`<p id="message" class="text-green-300 text-center">
		Email sent successfully &#x2713;</p>`)
    	fmt.Fprint(w, message)
	}
}

func postMonthlyPayment(w http.ResponseWriter, r *http.Request) {
	principal, _ := strconv.ParseFloat(r.PostFormValue("purchasePrice"), 64) 
	lengthOfMortgageInMonths, _ := strconv.ParseFloat(r.PostFormValue("mortgageTerm"), 64) 
	downPayment, _ := strconv.ParseFloat(r.PostFormValue("downPayment"), 64)  
	annualTaxes, _ := strconv.ParseFloat(r.PostFormValue("annualTaxes"), 64)  
	interestRate, _ := strconv.ParseFloat(r.PostFormValue("interestRate"), 64) 
	annualInsurance, _ := strconv.ParseFloat(r.PostFormValue("annualInsurance"), 64) 
	monthlyHoa,  _ := strconv.ParseFloat(r.PostFormValue("monthlyHoa"), 64) 

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
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/css/", twhandler.New(http.Dir("css"), "/css", twembed.New()))

	port := os.Getenv("PORT")
		if port == "" {
    	log.Fatal("$PORT must be set")
	}

	s := &http.Server{Addr: ":" + port, Handler: mux}

	mux.HandleFunc("/getLoanDescription", getLoanDescription)
	mux.HandleFunc("/postMonthlyPayment", postMonthlyPayment)
	mux.HandleFunc("/postSendEmailAndSaveInDb", postSendEmailAndSaveInDb)
    
	fmt.Println("Now listening on: http://localhost:" + port)
	log.Fatal(s.ListenAndServe())
}