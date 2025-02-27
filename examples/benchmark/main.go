package main

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// User contains user information
type User struct {
	FirstName      string     `validate:"required"`
	LastName       string     `validate:"required"`
	Age            uint8      `validate:"gte=0,lte=130"`
	Email          string     `validate:"required,email"`
	Gender         string     `validate:"oneof=male female prefer_not_to"`
	FavouriteColor string     `validate:"iscolor"`
	Addresses      []*Address `validate:"required,dive,required"`
}

// Address houses a users address information
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

func main() {
	// Create a valid user for benchmarking
	addresses := make([]*Address, 0, 100)
	for i := 0; i < 100; i++ {
		addresses = append(addresses, &Address{
			Street: fmt.Sprintf("Street %d", i),
			City:   fmt.Sprintf("City %d", i),
			Planet: fmt.Sprintf("Planet %d", i),
			Phone:  fmt.Sprintf("Phone %d", i),
		})
	}

	user := &User{
		FirstName:      "John",
		LastName:       "Doe",
		Age:            30,
		Email:          "john.doe@example.com",
		Gender:         "male",
		FavouriteColor: "#00ff00",
		Addresses:      addresses,
	}

	// Benchmark go-playground/validator
	validate := validator.New()
	iterations := 10000

	fmt.Println("Benchmarking go-playground/validator...")
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = validate.Struct(user)
	}
	duration := time.Since(start)
	fmt.Printf("Time taken for %d iterations: %v\n", iterations, duration)
	fmt.Printf("Average time per validation: %v\n", duration/time.Duration(iterations))

	// Benchmark QuickValidate
	// Note: This will use the generated Validate() method
	fmt.Println("\nBenchmarking QuickValidate...")
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_ = user.Validate()
	}
	duration = time.Since(start)
	fmt.Printf("Time taken for %d iterations: %v\n", iterations, duration)
	fmt.Printf("Average time per validation: %v\n", duration/time.Duration(iterations))
}
