# QuickValidate

[English](README.md) | [中文](README_zh.md)

QuickValidate is a static code generation version of the popular [go-playground/validator](https://github.com/go-playground/validator) library. It generates Go code for validation at build time instead of using reflection at runtime, which significantly improves performance.

## Features

- Generate static validation code from struct tags
- Compatible with go-playground/validator's tag syntax
- Significantly faster than reflection-based validation
- No runtime reflection overhead
- Type-safe validation
- Compile-time validation errors
- Supports common validation tags: required, email, url, oneof, min/max, etc.

## Installation

```bash
go install github.com/antlabs/quickvalidate/cmd/quickvalidate@latest
```

## Usage

1. Define your structs with validation tags (same as go-playground/validator)

```go
type User struct {
    FirstName      string     `validate:"required"`
    LastName       string     `validate:"required"`
    Age            uint8      `validate:"gte=0,lte=130"`
    Email          string     `validate:"required,email"`
    Gender         string     `validate:"oneof=male female prefer_not_to"`
    FavouriteColor string     `validate:"iscolor"`
    Website        string     `validate:"url"`
    Addresses      []*Address `validate:"required,dive,required"`
}

type Address struct {
    Street string `validate:"required"`
    City   string `validate:"required"`
    Planet string `validate:"required"`
    Phone  string `validate:"required"`
}
```

2. Run the quickvalidate command to generate validation code

```bash
quickvalidate -i ./path/to/your/package -o ./path/to/output/validators.go
```

3. Use the generated validation code

```go
user := &User{...}
err := user.Validate() // Generated method
if err != nil {
    // Handle validation errors
}
```

## Performance

QuickValidate is significantly faster than reflection-based validation. Benchmarks show it can be up to 10-20x faster for complex structs.

## License

MIT
