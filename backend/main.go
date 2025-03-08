package main

import (
	"car-rental/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Database connection settings
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "admin" // Change this to your actual password
	dbname   = "Car_Rental"
)

var db *sql.DB

// Connect to the PostgreSQL database
func connectDB() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("🚀 Connected to PostgreSQL database!")
}

// Get all cars
func getCars(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM cars")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var cars []models.Car
	for rows.Next() {
		var car models.Car
		if err := rows.Scan(&car.ID, &car.Brand, &car.Model, &car.PricePerDay, &car.Availability); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cars = append(cars, car)
	}
	c.JSON(http.StatusOK, cars)
}

// Add a new car
func addCar(c *gin.Context) {
	var car models.Car
	if err := json.NewDecoder(c.Request.Body).Decode(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("INSERT INTO cars (brand, model, price_per_day, availability) VALUES ($1, $2, $3, $4)",
		car.Brand, car.Model, car.PricePerDay, car.Availability)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Car added successfully!"})
}

// Rent a car with pickup & drop-off datetime
// Rent a car (Create a new rental record)
func rentCar(c *gin.Context) {
	var rental models.Rental
	if err := c.ShouldBindJSON(&rental); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure rental_date is set
	if rental.RentalDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rental_date is required"})
		return
	}

	_, err := db.Exec("INSERT INTO rentals (customer_id, car_id, rental_date, pickup_datetime, dropoff_datetime, status) VALUES ($1, $2, $3, $4, $5, $6)",
		rental.CustomerID, rental.CarID, rental.RentalDate, rental.PickupDatetime, rental.DropoffDatetime, rental.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Mark the car as unavailable
	_, _ = db.Exec("UPDATE cars SET availability = FALSE WHERE id = $1", rental.CarID)

	c.JSON(http.StatusCreated, gin.H{"message": "Car rented successfully!"})
}

// Return a car
func returnCar(c *gin.Context) {
	id := c.Param("id")

	// Set status to returned and update drop-off datetime
	_, err := db.Exec("UPDATE rentals SET status = 'returned', dropoff_datetime = NOW() WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set car availability to true
	_, _ = db.Exec("UPDATE cars SET availability = TRUE WHERE id = (SELECT car_id FROM rentals WHERE id = $1)", id)

	c.JSON(http.StatusOK, gin.H{"message": "Car returned successfully!"})
}

// Delete a car
func deleteCar(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM cars WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully!"})
}

// Get all customers
func getCustomers(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		customers = append(customers, customer)
	}
	c.JSON(http.StatusOK, customers)
}

// Get all rentals
func getRentals(c *gin.Context) {
	rows, err := db.Query("SELECT id, customer_id, car_id, pickup_datetime, dropoff_datetime, status FROM rentals")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var rentals []models.Rental
	for rows.Next() {
		var rental models.Rental
		var dropoffDatetime sql.NullString

		// Scan data into variables
		err := rows.Scan(&rental.ID, &rental.CustomerID, &rental.CarID, &rental.PickupDatetime, &dropoffDatetime, &rental.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Convert sql.NullString to *string (set NULL values to nil)
		if dropoffDatetime.Valid {
			rental.DropoffDatetime = &dropoffDatetime.String
		} else {
			rental.DropoffDatetime = nil
		}

		// Append formatted rental data
		rentals = append(rentals, rental)
	}

	c.JSON(http.StatusOK, rentals)
}

// Get a single customer by ID
func getCustomerByID(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT * FROM customers WHERE id = $1", id)

	var customer models.Customer
	err := row.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return

	}

	c.JSON(http.StatusOK, customer)
}

// Add a new customer
func addCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("INSERT INTO customers (name, email, phone) VALUES ($1, $2, $3)",
		customer.Name, customer.Email, customer.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Customer added successfully!"})
}

// Delete a customer by ID
func deleteCustomer(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully!"})
}

// Update an existing customer
func updateCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer

	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE customers SET name = $1, email = $2, phone = $3 WHERE id = $4",
		customer.Name, customer.Email, customer.Phone, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer updated successfully!"})
}

// Get a single car by ID
func getCarByID(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT * FROM cars WHERE id = $1", id)

	var car models.Car
	err := row.Scan(&car.ID, &car.Brand, &car.Model, &car.PricePerDay, &car.Availability)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, car)
}

// Update car details
func updateCar(c *gin.Context) {
	id := c.Param("id")
	var car models.Car

	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE cars SET brand = $1, model = $2, price_per_day = $3, availability = $4 WHERE id = $5",
		car.Brand, car.Model, car.PricePerDay, car.Availability, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Car updated successfully!"})
}

// Delete a rental record by ID
func deleteRental(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM rentals WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rental deleted successfully!"})
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	// Customers API
	router.GET("/customers", getCustomers)          // Get all customers
	router.GET("/customers/:id", getCustomerByID)   // Get single customer
	router.POST("/customers", addCustomer)          // Add a new customer
	router.PUT("/customers/:id", updateCustomer)    // Update customer
	router.DELETE("/customers/:id", deleteCustomer) // Delete customer

	// Cars API
	router.GET("/cars", getCars)          // Get all cars
	router.GET("/cars/:id", getCarByID)   // Get single car
	router.POST("/cars", addCar)          // Add a new car
	router.PUT("/cars/:id", updateCar)    // Update car
	router.DELETE("/cars/:id", deleteCar) // Delete car

	// Rentals API
	router.GET("/rentals", getRentals)          // Get all rentals
	router.POST("/rent", rentCar)               // Rent a car
	router.PUT("/return/:id", returnCar)        // Return a rented car
	router.DELETE("/rentals/:id", deleteRental) // Delete rental record

	return router
}

func main() {
	connectDB()
	defer db.Close()

	r := setupRouter()
	fmt.Println("🚀 Server running on port 8080")
	r.Run(":8080")
}
