# Mortgage Calculator App

This is a mortgage calculator application built with Golang, HTMX, TailwindCSS, and PostgreSQL DB. The application allows users to calculate mortgage payments based on their input parameters.

## Features

- Calculate mortgage payments based on principal, interest rate, loan term, down payment, annual taxes, insurance, and monthly HOA fees.
- Store mortgage information in a PostgreSQL database.
- Utilizes HTMX for dynamic and interactive web features.
- Styled with TailwindCSS for a modern and responsive design.

## Nice to Haves & Future Improvements

- **Styling for Mobile**: Enhance the application's responsiveness by implementing mobile-friendly styling.
- **API Routes for Real-time Property Tax & Interest Rates**: Integrate API routes to retrieve real-time local property tax information and average interest rates offered. This feature will provide users with more accurate and up-to-date calculations.
- **Integration with Courier for Automated Notifications**: Utilize Courier, a service for sending out mortgage information to users. This will streamline the communication process and enhance user experience.

## Technologies Used

- **Golang**: Backend development and logic.
- **HTMX**: Frontend interaction and dynamic content.
- **TailwindCSS**: Styling and design.
- **PostgreSQL**: Database management and storage.
- **Courier**: Service for sending out mortgage information to users.

## Usage

To run the application locally, make sure you have Go installed along with PostgreSQL. Clone the repository and configure the necessary environment variables. Then, run the application using the following command:

```bash
go run main.go
```

## Contributing

Contributions are welcome! If you have any ideas for improvements or new features, feel free to open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
