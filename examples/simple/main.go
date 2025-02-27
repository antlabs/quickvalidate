package main

import (
	"fmt"
	"log"
	
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

// Validate validates the Address struct
func (a *Address) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

// Validate validates the User struct
func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

func main() {
	// Create a user with validation errors
	address := &Address{
		Street: "Eavesdown Docks",
		Planet: "Persphone",
		// City is missing - validation error
		Phone: "none",
	}

	user := &User{
		FirstName: "Badger",
		LastName:  "Smith",
		Age:       135, // Too old - validation error
		Gender:    "male",
		Email:     "Badger.Smith@gmail", // Invalid email - validation error
		// FavouriteColor is missing - validation error
		Addresses: []*Address{address},
	}

	// Validate the user
	err := user.Validate()
	if err != nil {
		log.Printf("Validation errors: %v\n", err)
	} else {
		fmt.Println("User is valid!")
	}

	// Fix the validation errors
	user.Age = 35
	user.Email = "Badger.Smith@gmail.com"
	user.FavouriteColor = "#00ff00"
	address.City = "New Mombasa"

	// Validate again
	err = user.Validate()
	if err != nil {
		log.Printf("Validation errors: %v\n", err)
	} else {
		fmt.Println("User is valid!")
	}
}
