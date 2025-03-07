package main

import (
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

	var cars []Car
	for rows.Next() {
		var car Car
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
	var car Car
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

// Rent a car
func rentCar(c *gin.Context) {
	var rental Rental
	if err := json.NewDecoder(c.Request.Body).Decode(&rental); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("INSERT INTO rentals (customer_id, car_id, rental_date, return_date, status) VALUES ($1, $2, $3, $4, 'active')",
		rental.CustomerID, rental.CarID, rental.RentalDate, rental.ReturnDate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, _ = db.Exec("UPDATE cars SET availability = FALSE WHERE id = $1", rental.CarID)

	c.JSON(http.StatusCreated, gin.H{"message": "Car rented successfully!"})
}

// Return a car
func returnCar(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("UPDATE rentals SET status = 'returned' WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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

	var customers []Customer
	for rows.Next() {
		var customer Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		customers = append(customers, customer)
	}
	c.JSON(http.StatusOK, customers)
}

// Get a single customer by ID
func getCustomerByID(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT * FROM customers WHERE id = $1", id)

	var customer Customer
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
	var customer Customer
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

func setupRouter() *gin.Engine {
	router := gin.Default()

	// Car routes
	router.GET("/cars", getCars)
	router.POST("/cars", addCar)
	router.DELETE("/cars/:id", deleteCar)

	// Rental routes
	router.POST("/rent", rentCar)
	router.PUT("/return/:id", returnCar)

	// Customer routes
	router.GET("/customers", getCustomers)
	router.GET("/customers/:id", getCustomerByID)
	router.POST("/customers", addCustomer)
	router.DELETE("/customers/:id", deleteCustomer)

	return router
}

func main() {
	connectDB()
	defer db.Close()

	r := setupRouter()
	fmt.Println("🚀 Server running on port 8080")
	r.Run(":8080")
}
