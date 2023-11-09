// seeds/seed.go

package seeds

import (
	"database/sql"
)

func SeedDatabase(db *sql.DB) error {
    schemaStatements := []string{
        `CREATE DATABASE IF NOT EXISTS saved_user_mortgagedb;`,  
        `USE saved_user_mortgagedb;`,

        `CREATE TABLE IF NOT EXISTS loans (
            id serial PRIMARY KEY,
            loan_type varchar(255) NOT NULL,
            description varchar(255),
        );`,

        `CREATE TABLE IF NOT EXISTS mortgageinfo (
            id serial PRIMARY KEY,
            monthly_mortgage_payment varchar(255) NOT NULL,
            principal varchar(255) NOT NULL,
            mortgage_term varchar(255) NOT NULL,
            annual_taxes varchar(255) NOT NULL,
            down_payment varchar(255) NOT NULL,
            interest_rate varchar(255) NOT NULL,
            annual_insurance varchar(255) NOT NULL,
            monthly_hoa varchar(255) NOT NULL,
            email varchar(255) NOT NULL,
            date_created DATETIME
        );`,
    }

    dataStatements := []string{
        `INSERT INTO loans (loan_type, description)
		VALUES ('conventional', 'A conventional mortgage isn''t government-backed, provided by private lenders like banks. Qualification involves a strong credit history and financial stability, following lender guidelines. These loans feature fixed or adjustable rates and might require PMI with lower down payments.');
		`,
        `INSERT INTO loans (loan_type, description)
		VALUES ('fha', 'FHA loans are government-backed mortgages insured by the Federal Housing Administration. They feature a low 3.5%% down payment, flexible credit criteria, and fixed or adjustable interest rates. These loans, available through FHA-approved lenders, promote homeownership and provide lender protection through government insurance.');
		`,
        `INSERT INTO loans (loan_type, description)
		VALUES ('jumbo', 'A jumbo loan is designed for high-end properties exceeding typical loan limits, offering variable interest rates. Qualification often entails stringent credit and income criteria. It's an attractive choice for luxury real estate or properties in upscale markets. If you're considering a jumbo loan, let's discuss how it can be tailored to your specific home purchase needs.');
		`,
        `INSERT INTO loans (loan_type, description)
		VALUES ('refinance', 'Refinancing replaces your current mortgage for better rates or terms, aligning with your financial goals. It may reduce payments or unlock home equity for improvements or debt consolidation. Contact us to explore refinancing options and make a well-informed decision tailored to your needs.');
		`,
        `INSERT INTO loans (loan_type, description)
		VALUES ('vaLoan', 'A VA loan is a government-backed mortgage for eligible veterans, active-duty service members, and their qualified spouses. It offers benefits like no down payment and competitive interest rates to make homeownership easier for military personnel. Contact us to explore VA loan options.');
		`,

    }

    for _, sqlStatement := range schemaStatements {
        _, err := db.Exec(sqlStatement)
        if err != nil {
            return err
        }
    }

    for _, sqlStatement := range dataStatements {
        _, err := db.Exec(sqlStatement)
        if err != nil {
            return err
        }
    }

    return nil
}
