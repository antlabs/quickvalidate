package main

import (
	"fmt"
	"log"
)

// Product represents a product with various numeric validation rules
type Product struct {
	ID          int     `validate:"gt=0"`         // ID must be greater than 0
	Price       float64 `validate:"gte=0.01"`     // Price must be at least 0.01
	Stock       int     `validate:"gte=0"`        // Stock must be non-negative
	Weight      float64 `validate:"gt=0,lt=1000"` // Weight must be positive and less than 1000
	Rating      float64 `validate:"gte=0,lte=5"`  // Rating must be between 0 and 5
	CategoryID  int     `validate:"eq=1"`         // CategoryID must be exactly 1
	MaxQuantity int     `validate:"eq=10"`        // MaxQuantity must be exactly 10
	MinAge      int     `validate:"lte=18"`       // MinAge must be at most 18
	MaxAge      int     `validate:"gte=18"`       // MaxAge must be at least 18
}

// Order represents an order with validation rules
type Order struct {
	OrderID     int     `validate:"gt=1000"`      // OrderID must be greater than 1000
	TotalAmount float64 `validate:"gt=0"`         // TotalAmount must be positive
	ItemCount   int     `validate:"gt=0,lte=100"` // ItemCount must be positive and at most 100
	Status      string  `validate:"eq=pending"`   // Status must be exactly "pending"
	Priority    int     `validate:"eq=1"`         // Priority must be exactly 1
}

func main() {
	// Valid product
	validProduct := Product{
		ID:          1,
		Price:       9.99,
		Stock:       100,
		Weight:      5.5,
		Rating:      4.5,
		CategoryID:  1,
		MaxQuantity: 10,
		MinAge:      16,
		MaxAge:      18,
	}

	// Invalid product (fails multiple validations)
	invalidProduct := Product{
		ID:          0,    // Should be > 0
		Price:       0,    // Should be >= 0.01
		Stock:       -1,   // Should be >= 0
		Weight:      1500, // Should be < 1000
		Rating:      6,    // Should be <= 5
		CategoryID:  2,    // Should be == 1
		MaxQuantity: 15,   // Should be == 10
		MinAge:      21,   // Should be <= 18
		MaxAge:      16,   // Should be >= 18
	}

	// Valid order
	validOrder := Order{
		OrderID:     1001,
		TotalAmount: 150.75,
		ItemCount:   5,
		Status:      "pending",
		Priority:    1,
	}

	// Invalid order
	invalidOrder := Order{
		OrderID:     999,       // Should be > 1000
		TotalAmount: 0,         // Should be > 0
		ItemCount:   150,       // Should be <= 100
		Status:      "shipped", // Should be "pending"
		Priority:    2,         // Should be == 1
	}

	// Validate the valid product
	if err := validProduct.Validate(); err != nil {
		log.Fatalf("Valid product validation failed: %v", err)
	}
	fmt.Println("Valid product passed validation")

	// Validate the invalid product
	if err := invalidProduct.Validate(); err != nil {
		fmt.Println("Invalid product validation failed (as expected):")
		fmt.Printf("%v\n", err)
	} else {
		log.Fatal("Invalid product unexpectedly passed validation")
	}

	// Validate the valid order
	if err := validOrder.Validate(); err != nil {
		log.Fatalf("Valid order validation failed: %v", err)
	}
	fmt.Println("Valid order passed validation")

	// Validate the invalid order
	if err := invalidOrder.Validate(); err != nil {
		fmt.Println("Invalid order validation failed (as expected):")
		fmt.Printf("%v\n", err)
	} else {
		log.Fatal("Invalid order unexpectedly passed validation")
	}
}
