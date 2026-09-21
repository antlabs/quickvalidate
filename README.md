# QuickValidate

[English](README.md) | [中文](README_zh.md)

QuickValidate is a **static code generator** for [go-playground/validator](https://github.com/go-playground/validator): it reads `validate` struct tags and emits validation code at build time, so validating a struct does not walk it through reflection.

The generated code is **behaviourally aligned with go-playground/validator v10.25.0** — same failing fields, tags, params and namespaces for the same input — and that alignment is checked by the differential suite in this repo.

## Features

- Same tag syntax and semantics as go-playground/validator: `required`, `omitempty`, `min/max`, `dive`, `keys/endkeys`, cross field, conditional presence, and **all 179 built-in tags** of v10.25.0 (`bakedInValidators`, the aliases, and the structural tags, checked one by one against the upstream source)
- Parameters are parsed and dispatched at generation time; the emitted code only compares and calls helpers
- Same-package named types (`type Email string`) work like their underlying type, as do pointer, slice and map elements
- Embedded structs are traversed exactly like the reference, including the namespace (`Outer.Base.Field`) and embedded pointers
- Type and parameter mistakes surface at generation time (validator panics at run time instead)
- Errors carry what the reference carries: `Namespace`, `StructNamespace`, `Field`, `Tag`, `ActualTag` (an alias reports what it expands to), `Param`, `Kind`, `Value`
- No reflection on the validation path: the emitted code compares values and calls one reflect-free helper package. `reflect` shows up only for its `Kind` constants in errors, for the presence check of struct and array fields (which is what the reference does), and for an interface field's error metadata
- Ships a **differential test suite** comparing generated output against the real library

## Install

```bash
go install github.com/antlabs/quickvalidate/cmd/quickvalidate@latest
```

## Usage

1. Write structs exactly as you would for go-playground/validator:

```go
type User struct {
    FirstName      string     `validate:"required"`
    Email          string     `validate:"required,email"`
    Age            uint8      `validate:"gte=0,lte=130"`
    FavouriteColor string     `validate:"iscolor"`
    Addresses      []*Address `validate:"required,dive,required"`
}
```

2. Generate:

```bash
quickvalidate -i ./path/to/your/package -o ./path/to/output/validate_gen.go
```

Or add `//go:generate quickvalidate -i . -o validate_gen.go` to the package.

3. Call the generated method:

```go
user := &User{...}
if err := user.Validate(); err != nil {
    var errs errors.ValidationErrors
    if errors.As(err, &errs) {
        for _, e := range errs {
            fmt.Println(e.Namespace, e.Tag, e.Param, e.Kind)
        }
    }
}
```

```
User.Email  email    ""    string
User.Color  iscolor  ""    string   // e.ActualTag is "hexcolor|rgb|rgba|hsl|hsla"
User.Age    lte      130   uint8
```

## Flags

- `-i` — input file or directory (directories are walked; `_test.go` and `*_gen.go` are skipped)
- `-o` — output file
- `-pkg` — package name (derived from the input by default)

## Coverage

Every baked-in tag of go-playground/validator v10.25.0 is supported, including the comparison/length family, `dive`/`keys`/`endkeys`, `unique`, cross-field and cross-struct comparisons, the conditional `required_*`/`excluded_*`/`skip_unless` family, the `contains*`/`excludes*`/`startswith` family, and the format tags (`email`, `url`, `uuid*`, `ulid`, `ip*`, `cidr*`, `mac`, `hostname`, `fqdn`, `json`, `jwt`, `base64*`, colour tags, `iso3166_*`, `iso4217*`, `bcp47_language_tag`, `cron`, `spicedb`, `credit_card`, `luhn_checksum`, `mongodb*`, `cve`, `semver`, `bic`, `dns_rfc1035_label`, `postcode_iso3166_*`, `eth_addr*`, `btc_addr*`, `isbn*`, `issn`, hashes, …) plus the `iscolor` / `country_code` / `eu_country_code` aliases and `a|b` or-branches.

Not supported:

- automatic recursion into structs from other packages (validator reflects through them; the generator can only recurse into same-package structs it generated). This is the one case that is skipped **silently**
- everything that is a *runtime* registration: custom validators (`RegisterValidation`), struct-level validation (`RegisterStructValidation`), translations, `RegisterCustomTypeFunc`, `RegisterTagNameFunc`, `RegisterAlias`, and the options `WithRequiredStructEnabled` / `WithPrivateFieldValidation` / `SetTagName`. A generated validator is fixed at build time, so there is no registry to configure at startup
- the runtime entry points `Var` / `VarWithValue` (they take a tag string) and `ValidateMap`
- `StructPartial` / `StructExcept` / `StructFiltered`, which select fields at run time

## Alignment notes

The generator targets `validator.New()` (default configuration):

- `required` on non-pointer struct fields is skipped, matching `requiredStructEnabled=false`
- `image` on `[]byte` is always false — v10.25.0's `isImage` only implements the string branch
- the number-reading tags (`iso3166_1_alpha_numeric*`, `iso4217_numeric`, `port`) dispatch on the field kind, mirroring the upstream `field.Kind()` switches one by one: country codes take strings, signed and unsigned integers (modulo 1000 before the cast), currency codes take integers only. Combinations the reference panics on (e.g. `iso4217_numeric` on a string field) are generation-time errors here
- `spicedb` only takes `permission`, `type`, `id` or no parameter; anything else is a generation-time error (validator panics at run time)
- on an `interface{}` field only the tags that ask whether the interface is set can be emitted (`required`, `isdefault`, the `required_*`/`excluded_*` family, `omitempty`/`omitnil`/`omitzero`). Tags needing the dynamic value (`min`, `email`, `dive`, …) are generation-time errors: the reference evaluates them against whatever the interface holds at run time, which static code cannot know
- environment-dependent tags (`timezone`, `bcp47_language_tag`, `*_addr` resolvers) behave like the reference in the same environment
- a string-only tag on a field that is not a string (`email` on an `int`, `uuid` on a struct) is a generation-time error. The reference reads such fields through reflect's `String()`, so it would report that field as failing for the rest of the program's life
- `json` is the one string tag that also reads a byte slice, and it accepts both (a named `[]byte` type included)
- inputs that make the reference implementation itself panic (e.g. `unique` over a slice holding nil pointers) are logged and skipped by the suite — currently 60 cases, with 5021 real comparisons per run

## Verification

`internal/align` is a differential suite: the same structs and the same values are run through both the real library and the generated code, and every failure is compared as `namespace|tag|param|actualTag|structNamespace|kind`.

```bash
make align     # or: go test ./internal/align/
```

It combines hand written boundary cases (empty, boundary, Unicode, nil pointers, embedded structs, interface fields, cross-field paths through pointers, maps, aliases, or-branches, named types, unsigned wrap-around) with randomised differential testing over per-type value pools.

All 179 built-in tags appear in the corpus. A run performs 5021 comparisons; the 60 cases the reference itself panics on are skipped and reported.

## Performance

`go test -run '^$' -bench . ./examples/benchmark/` (measured locally, Apple M4 Pro):

| struct | go-playground/validator | generated |
|---|---|---|
| flat, 6 format checks | 631 ns, 1 alloc | 480 ns, 0 allocs |
| nested, `required,dive,required` over 100 addresses | 10554 ns, 101 allocs | 2384 ns, 0 allocs |

The ratio depends on the struct; measure your own. `make benchmark` still prints the simple loop the original example used.

## Build & test

```bash
make build      # build the CLI
make test       # everything, including the alignment suite
make align      # the alignment suite only
make generate   # regenerate examples and corpus
make examples   # run the examples

go test -run '^$' -bench . ./examples/benchmark/   # generated vs the reference
```

## License

MIT. The validation logic and data tables in `pkg/validators` are ported from [go-playground/validator](https://github.com/go-playground/validator) (MIT, Copyright (c) 2015 Dean Karn); file headers note the origin.
