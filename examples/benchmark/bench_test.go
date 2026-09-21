package main

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
)

// The struct shapes below are the ones a generated validator is usually judged
// on: a flat struct with format checks, a nested one, and one where the depth is
// in a slice of pointers.

func benchUser(addresses int) *User {
	as := make([]*Address, 0, addresses)
	for i := 0; i < addresses; i++ {
		as = append(as, &Address{
			Street: fmt.Sprintf("Street %d", i),
			City:   fmt.Sprintf("City %d", i),
			Planet: fmt.Sprintf("Planet %d", i),
			Phone:  fmt.Sprintf("Phone %d", i),
		})
	}
	return &User{
		FirstName:      "John",
		LastName:       "Doe",
		Age:            30,
		Email:          "john.doe@example.com",
		Gender:         "male",
		FavouriteColor: "#00ff00",
		Addresses:      as,
	}
}

func BenchmarkGeneratedFlat(b *testing.B) {
	user := benchUser(0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := user.Validate(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReferenceFlat(b *testing.B) {
	user := benchUser(0)
	v := validator.New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := v.Struct(user); err != nil {
			b.Fatal(err)
		}
	}
}

// 100 addresses exercises `required,dive,required` over a slice of pointers.
func BenchmarkGeneratedNested100(b *testing.B) {
	user := benchUser(100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := user.Validate(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReferenceNested100(b *testing.B) {
	user := benchUser(100)
	v := validator.New()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := v.Struct(user); err != nil {
			b.Fatal(err)
		}
	}
}
