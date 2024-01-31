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

	"github.com/gotailwindcss/tailwind/twembed"
	"github.com/gotailwindcss/tailwind/twhandler"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

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

    // Send the message
    requestID, err := client.SendMessage(context.Background(), request)
    if err != nil {
        log.Printf("Failed to send email: %v", err)
        return "", err
    }

    log.Printf("Email sent successfully, request ID: %s", requestID)
    return requestID, nil
}

func getLoanDescription(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDBName := os.Getenv("POSTGRES_DB_NAME")

	var (
		host     = "localhost"
		port     = 5432
		user     = postgresUser
		password = postgresPassword
		dbName   = postgresDBName
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbName)

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

	loanType := r.URL.Query().Get("loanType")
    query := "SELECT description FROM loans WHERE loan_type = $1"
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

		err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDBName := os.Getenv("POSTGRES_DB_NAME")

	var (
		host     = "localhost"
		port     = 5432
		user     = postgresUser
		password = postgresPassword
		dbName   = postgresDBName
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbName)

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        log.Fatal(err)
    }

	defer db.Close()

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

	insertQuery := "INSERT INTO MortgageInfo (monthly_mortgage_payment, principal, mortgage_term, annual_taxes, down_payment, interest_rate, annual_insurance, monthly_hoa, email, date_created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	_, err = db.Exec(insertQuery, info.MonthlyMortgagePayment, info.Principal, info.MortgageTerm, info.AnnualTaxes, info.DownPayment, info.InterestRate, info.AnnualInsurance, info.MonthlyHOA, info.Email, info.DateCreated)

	if err != nil {
		panic(err)
	}

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
	yearlyTaxes := multiply(principal, annualTaxes)/100
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
type Loans struct {
    id    int
    loan_type  string
    description string
}

func createTables(db *sql.DB) error {
	sqlStmtDropLoans := `DROP TABLE IF EXISTS loans`
	_, errLoansDrop := db.Exec(sqlStmtDropLoans)
	if errLoansDrop != nil {
		return errLoansDrop
	}
	fmt.Println("Loans table dropped successfully")
	sqlStmtAddLoans := `
		CREATE TABLE IF NOT EXISTS loans (
			id SERIAL PRIMARY KEY,
			loan_type TEXT,
			description TEXT
		)
	`

	// Execute the SQL statement
	_, errLoansAdd := db.Exec(sqlStmtAddLoans)
	if errLoansAdd != nil {
		return errLoansAdd
	}
	fmt.Println("Loans table created successfully")

	sqlStmtAddInfo := `
		CREATE TABLE IF NOT EXISTS MortgageInfo (
			id SERIAL PRIMARY KEY,
			monthly_mortgage_payment TEXT,
			principal TEXT,
			mortgage_term TEXT,
			annual_taxes TEXT,
			down_payment TEXT,
			interest_rate TEXT,
			annual_insurance TEXT,
			monthly_hoa TEXT,
			email TEXT,
			date_created DATE
		)
	`

	// Execute the SQL statement
	_, errInfo := db.Exec(sqlStmtAddInfo)
	if errInfo != nil {
		return errInfo
	}
	fmt.Println("Mortgage Info table created successfully")
	return nil
}

func seedDb() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDBName := os.Getenv("POSTGRES_DB_NAME")

	var (
		host     = "localhost"
		port     = 5432
		user     = postgresUser
		password = postgresPassword
		dbName   = postgresDBName
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbName)

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Verify the connection
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Successfully connected to the database!")
	if err := createTables(db); err != nil {
		log.Fatal("Error creating items table:", err)
	}
    // Seed data into the database
    seedData := []Loans{
        {id: 1, loan_type: "conventional", description: "A conventional mortgage loan is a type of home financing that isn't government-backed and is offered by private lenders like banks and credit unions. Borrowers typically need a solid credit history and financial stability to qualify, adhering to lender underwriting guidelines. These loans can have either fixed or adjustable interest rates, and they may require private mortgage insurance (PMI) if the down payment is below a certain threshold."},
        {id: 2, loan_type: "fha", description: "An FHA loan is a government-backed mortgage insured by the Federal Housing Administration. It's known for its lower down payment requirement, often as low as 3.5%, making it an option for those with modest savings. These loans have flexible credit requirements and can have fixed or adjustable interest rates. FHA loans are offered by FHA-approved lenders and aim to help more people become homeowners while providing extra protection for lenders with government insurance."},
        {id: 3, loan_type: "jumbo", description: "A jumbo loan is a mortgage for higher-priced properties that exceed standard loan limits. It's offered by private lenders and may have different interest rate options. Qualifying for a jumbo loan typically requires meeting stricter credit and income criteria. This option is suitable for financing luxury homes or properties in pricey real estate markets. If you're interested, we can discuss how a jumbo loan could work for your home purchase."},
        {id: 4, loan_type: "refinance", description: "A refinance loan allows you to replace your existing mortgage with a new one, often to secure a better interest rate or change the loan terms. It's a way to adjust your current home loan to better suit your financial goals and circumstances. Refinancing can potentially reduce your monthly payments or help you access the equity in your home for other purposes, like home improvements or debt consolidation. If you're considering a refinance, we can explore the options available to you and help you make an informed decision."},
        {id: 5, loan_type: "vaLoan", description: "A VA loan is a special mortgage program for eligible veterans, active-duty service members, and select National Guard and Reserve members, along with qualifying spouses. Backed by the Department of Veterans Affairs (VA), it requires no down payment and offers competitive interest rates. Provided by private lenders, VA loans aim to make home buying easier for military personnel. If you qualify, we can assist you in considering this advantageous option for buying or refinancing your home."},
        
    }

    for _, loans := range seedData {
        _, err := db.Exec("INSERT INTO loans (id, loan_type, description) VALUES ($1, $2, $3)", loans.id, loans.loan_type, loans.description)
        if err != nil {
            log.Fatal(err)
        }
    }

    fmt.Println("Seed data inserted successfully!")
}


func setupServer() *http.Server {
    mux := http.NewServeMux()
    mux.Handle("/", http.FileServer(http.Dir("static")))
    mux.Handle("/css/", twhandler.New(http.Dir("css"), "/css", twembed.New()))

    mux.HandleFunc("/getLoanDescription", getLoanDescription)
    mux.HandleFunc("/postMonthlyPayment", postMonthlyPayment)
    mux.HandleFunc("/postSendEmailAndSaveInDb", postSendEmailAndSaveInDb)

    return &http.Server{Addr: ":" + getPort(), Handler: mux}
}

func getPort() string {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    port := os.Getenv("PORT")
    if port == "" {
        log.Fatal("$PORT must be set")
    }
    return port
}


func main() {
	fmt.Println("Mortgage Calculator")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/css/", twhandler.New(http.Dir("css"), "/css", twembed.New()))

    seedDb()

    server := setupServer()
    log.Fatal(server.ListenAndServe())
}