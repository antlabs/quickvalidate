package numeric

import (
	"fmt"
	"github.com/antlabs/quickvalidate/pkg/errors"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// Validate validates the Product struct
func (s *Product) Validate() error {
	var errs errors.ValidationErrors

	// Validate ID is greater than 0
	if !(s.ID > 0) {
		errs = append(errs, errors.ValidationError{
			Field:     "ID",
			Tag:       "gt",
			Param:     "0",
			Value:     s.ID,
			Namespace: "Product.ID",
		})
	}

	// Validate Price is greater than or equal to 0.01
	if !(s.Price >= 0.01) {
		errs = append(errs, errors.ValidationError{
			Field:     "Price",
			Tag:       "gte",
			Param:     "0.01",
			Value:     s.Price,
			Namespace: "Product.Price",
		})
	}

	// Validate Stock is greater than or equal to 0
	if !(s.Stock >= 0) {
		errs = append(errs, errors.ValidationError{
			Field:     "Stock",
			Tag:       "gte",
			Param:     "0",
			Value:     s.Stock,
			Namespace: "Product.Stock",
		})
	}

	// Validate Weight is greater than 0
	if !(s.Weight > 0) {
		errs = append(errs, errors.ValidationError{
			Field:     "Weight",
			Tag:       "gt",
			Param:     "0",
			Value:     s.Weight,
			Namespace: "Product.Weight",
		})
	}

	// Validate Weight is less than 1000
	if !(s.Weight < 1000) {
		errs = append(errs, errors.ValidationError{
			Field:     "Weight",
			Tag:       "lt",
			Param:     "1000",
			Value:     s.Weight,
			Namespace: "Product.Weight",
		})
	}

	// Validate Rating is greater than or equal to 0
	if !(s.Rating >= 0) {
		errs = append(errs, errors.ValidationError{
			Field:     "Rating",
			Tag:       "gte",
			Param:     "0",
			Value:     s.Rating,
			Namespace: "Product.Rating",
		})
	}

	// Validate Rating is less than or equal to 5
	if !(s.Rating <= 5) {
		errs = append(errs, errors.ValidationError{
			Field:     "Rating",
			Tag:       "lte",
			Param:     "5",
			Value:     s.Rating,
			Namespace: "Product.Rating",
		})
	}

	// Validate CategoryID is equal to 1
	if s.CategoryID != 1 {
		errs = append(errs, errors.ValidationError{
			Field:     "CategoryID",
			Tag:       "eq",
			Param:     "1",
			Value:     s.CategoryID,
			Namespace: "Product.CategoryID",
		})
	}

	// Validate MaxQuantity is equal to 10
	if s.MaxQuantity != 10 {
		errs = append(errs, errors.ValidationError{
			Field:     "MaxQuantity",
			Tag:       "eq",
			Param:     "10",
			Value:     s.MaxQuantity,
			Namespace: "Product.MaxQuantity",
		})
	}

	// Validate MinAge is less than or equal to 18
	if !(s.MinAge <= 18) {
		errs = append(errs, errors.ValidationError{
			Field:     "MinAge",
			Tag:       "lte",
			Param:     "18",
			Value:     s.MinAge,
			Namespace: "Product.MinAge",
		})
	}

	// Validate MaxAge is greater than or equal to 18
	if !(s.MaxAge >= 18) {
		errs = append(errs, errors.ValidationError{
			Field:     "MaxAge",
			Tag:       "gte",
			Param:     "18",
			Value:     s.MaxAge,
			Namespace: "Product.MaxAge",
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isOneOf(value string, allowedValues []string) bool {
	for _, v := range allowedValues {
		if value == v {
			return true
		}
	}
	return false
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// Validate validates the Order struct
func (s *Order) Validate() error {
	var errs errors.ValidationErrors

	// Validate OrderID is greater than 1000
	if !(s.OrderID > 1000) {
		errs = append(errs, errors.ValidationError{
			Field:     "OrderID",
			Tag:       "gt",
			Param:     "1000",
			Value:     s.OrderID,
			Namespace: "Order.OrderID",
		})
	}

	// Validate TotalAmount is greater than 0
	if !(s.TotalAmount > 0) {
		errs = append(errs, errors.ValidationError{
			Field:     "TotalAmount",
			Tag:       "gt",
			Param:     "0",
			Value:     s.TotalAmount,
			Namespace: "Order.TotalAmount",
		})
	}

	// Validate ItemCount is greater than 0
	if !(s.ItemCount > 0) {
		errs = append(errs, errors.ValidationError{
			Field:     "ItemCount",
			Tag:       "gt",
			Param:     "0",
			Value:     s.ItemCount,
			Namespace: "Order.ItemCount",
		})
	}

	// Validate ItemCount is less than or equal to 100
	if !(s.ItemCount <= 100) {
		errs = append(errs, errors.ValidationError{
			Field:     "ItemCount",
			Tag:       "lte",
			Param:     "100",
			Value:     s.ItemCount,
			Namespace: "Order.ItemCount",
		})
	}

	// Validate Status is equal to pending
	if s.Status != "pending" {
		errs = append(errs, errors.ValidationError{
			Field:     "Status",
			Tag:       "eq",
			Param:     "pending",
			Value:     s.Status,
			Namespace: "Order.Status",
		})
	}

	// Validate Priority is equal to 1
	if s.Priority != 1 {
		errs = append(errs, errors.ValidationError{
			Field:     "Priority",
			Tag:       "eq",
			Param:     "1",
			Value:     s.Priority,
			Namespace: "Order.Priority",
		})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isOneOf(value string, allowedValues []string) bool {
	for _, v := range allowedValues {
		if value == v {
			return true
		}
	}
	return false
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
