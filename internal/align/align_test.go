package align

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	qverrors "github.com/antlabs/quickvalidate/pkg/errors"
)

// The whole point of this package: for every value, the statically generated
// validators must report exactly the same failures as go-playground/validator.
//
// The oracle is validator.New() (the default configuration, which is what the
// generator targets).

var oracle = validator.New()

// compared counts the differential comparisons performed by this run.
var compared int64

func TestMain(m *testing.M) {
	code := m.Run()
	fmt.Printf("differential comparisons: %d (oracle panics skipped: %d)\n",
		atomic.LoadInt64(&compared), atomic.LoadInt64(&skipped))
	os.Exit(code)
}

var skipped int64

// failures normalises both error types into a sorted list of
// "namespace|tag|param|actualTag|structNamespace|kind".
func failures(err error) []string {
	var out []string
	switch e := err.(type) {
	case nil:
		return nil
	case validator.ValidationErrors:
		for _, fe := range e {
			out = append(out, fmt.Sprintf("%s|%s|%s|%s|%s|%v",
				fe.Namespace(), fe.Tag(), fe.Param(),
				fe.ActualTag(), fe.StructNamespace(), fe.Kind()))
		}
	case qverrors.ValidationErrors:
		for _, fe := range e {
			out = append(out, fmt.Sprintf("%s|%s|%s|%s|%s|%v",
				fe.Namespace, fe.Tag, fe.Param,
				fe.ActualTag, fe.StructNamespace, fe.Kind))
		}
	default:
		out = append(out, "unexpected error type: "+reflect.TypeOf(err).String())
	}
	sort.Strings(out)
	return out
}

// oraclePanics runs the reference implementation, reporting whether it panicked
// (go-playground/validator has a few genuine crash paths, e.g. nil pointers
// inside a slice validated with `unique`; those inputs are not comparable).
func oraclePanics(v interface{}) (err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	return oracle.Struct(v), false
}

// compare runs one value through both implementations.
func compare(t *testing.T, name string, v interface{}) {
	t.Helper()

	wantErr, panicked := oraclePanics(v)
	if panicked {
		atomic.AddInt64(&skipped, 1)
		t.Logf("%s: validator panicked, skipping (not comparable)", name)
		return
	}
	atomic.AddInt64(&compared, 1)
	want := failures(wantErr)

	validatable, ok := v.(interface{ Validate() error })
	if !ok {
		t.Fatalf("%s: %T has no generated Validate method", name, v)
	}
	got := failures(validatable.Validate())

	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Errorf("%s\n  value: %s\n  validator: %v\n  generated: %v",
			name, describe(v), want, got)
	}
}

func describe(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	var parts []string
	for i := 0; i < rv.NumField(); i++ {
		parts = append(parts, fmt.Sprintf("%s=%v", rv.Type().Field(i).Name, rv.Field(i).Interface()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// ---------------------------------------------------------------------------
// curated cases: the boundary paths get explicit coverage
// ---------------------------------------------------------------------------

func TestCurated(t *testing.T) {
	one := "ab"
	empty := ""
	long := "abcdef"
	num := 7
	small := 3

	cases := []struct {
		name  string
		value interface{}
	}{
		{"scalars/zero", &Scalars{}},
		{"scalars/valid", &Scalars{
			Req: "x", Email: "a@b.com", MinS: "abc", MaxS: "abcde", LenS: "abcd",
			GtI: 6, GteI: 5, LtU: 4, LteU: 5, EqI: 5, NeI: 6, EqS: "abc", NeS: "abd",
			EqF: 1.5, GtF: 1.6, B: true, OmitMin: "abc", OmitPtr: &one,
			PlainPtr: &one, NilPtr: &num, SliceLen: []int{1, 2}, MapLen: map[string]int{"a": 1, "b": 2},
		}},
		{"scalars/boundaries", &Scalars{
			Req: "", Email: "a@b", MinS: "ab", MaxS: "abcdef", LenS: "abc",
			GtI: 5, GteI: 4, LtU: 5, LteU: 6, EqI: 4, NeI: 5, EqS: "ab", NeS: "abc",
			EqF: 1.6, GtF: 1.5, B: false, OmitMin: "", OmitPtr: nil,
			PlainPtr: nil, NilPtr: nil, SliceLen: []int{1}, MapLen: map[string]int{"a": 1},
		}},
		{"unicode/ascii-run-count", &Unicode{Min: "中文", Gt: "中文", Max: "中文", Len: "中文"}},
		{"unicode/valid", &Unicode{Min: "abc", Gt: "abc", Max: "abc", Len: "ab", Alpha: "abc", Alnum: "ab1"}},
		{"unicode/mixed", &Unicode{Min: "a中b", Gt: "a中", Max: "a中b", Len: "中文", Alpha: "a中b", Alnum: "a中b"}},
		{"unicode/empty", &Unicode{}},

		{"formats/all-empty", &Formats{}},
		{"formats/all-valid", &Formats{
			Email: "user@example.com", URL: "https://example.com/x", HTTPURL: "http://example.com",
			UUID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", UUID4: "6ba7b810-9dad-41d1-80b4-00c04fd430c8",
			ULID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", HexColor: "#fff", RGB: "rgb(1,2,3)",
			RGBA: "rgba(1,2,3,0.5)", HSL: "hsl(120,50%,50%)", Color: "#abc",
			ColorOr: "rgb(1,2,3)", IP: "127.0.0.1", IPv4: "127.0.0.1", IPv6: "::1",
			CIDR: "192.168.0.0/24", MAC: "01:23:45:67:89:ab", Hostname: "example",
			FQDN: "example.com", Port: 8080, HostPort: "example.com:8080",
			JSON: `{"a":1}`, JWT: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.abc",
			Base64: "aGVsbG8=", Base32: "MZXW6===", DataURI: "data:text/plain;base64,aGVsbG8=",
			Lat: "45.0", Lng: "-122.0", SSN: "123-45-6789", E164: "+15551234567",
			Lower: "abc", Upper: "ABC", ASCII: "abc", Multi: "中文",
			Hex: "deadbeef", Numeric: "123", Number: "1.5", Semver: "1.2.3",
			BIC: "DEUTDEFF", MongoID: "507f1f77bcf86cd799439011",
			MD5: strings.Repeat("a", 32), SHA256: strings.Repeat("b", 64),
			ISBN: "0-306-40615-2", ISBN13: "978-0-306-40615-7", ISSN: "0378-5955",
			Eth:     "0x323b7db6c8d3f0b1c3f8f0b0f9a0b1c2d3e4f5a6",
			BTC:     "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",
			Country: "CN", Currency: "USD", Lang: "zh-Hans-CN",
			Card: "4111111111111111", TZ: "Asia/Shanghai", DT: "2024-01-02",
		}},
		{"formats/all-invalid", &Formats{
			Email: "nope", URL: "not a url", HTTPURL: "ftp://x", UUID: "nope",
			Color: "nope", ColorOr: "nope", IP: "999.1.1.1", CIDR: "1.2.3.4",
			JSON: "{", Base64: "!!!", Numeric: "12a", Number: "1.2.3",
			Country: "XX", Currency: "XXX", DT: "02/01/2024", TZ: "Mars/Olympus",
		}},

		{"content/valid", &Content{Contains: "aafoob", ContainsA: "zzz", ContainsR: "axb",
			Excludes: "bar", ExcludesA: "zzz", ExcludesR: "zzz", Starts: "abc", Ends: "xyz",
			NotStarts: "zz", NotEnds: "zz"}},
		{"content/invalid", &Content{Contains: "bar", ContainsA: "zzz2", ContainsR: "bbb",
			Excludes: "foobar", ExcludesA: "cat", ExcludesR: "xxx", Starts: "zz", Ends: "zz",
			NotStarts: "abc", NotEnds: "xyz"}},
		{"content/empty", &Content{}},

		{"oneof/valid", &OneOfs{S: "red", CI: "RED", I: 2, U: 3}},
		{"oneof/invalid", &OneOfs{S: "pink", CI: "pink", I: 9, U: 9}},
		{"oneof/empty", &OneOfs{}},

		{"uniques/unique", &Uniques{Str: []string{"a", "b"}, Int: []int{1, 2},
			Ptrs: []*string{&one, &long}, Items: []uniqueItem{{Name: "a"}, {Name: "b"}},
			Map: map[string]int{"a": 1, "b": 2}, Single: "x", Other: "y"}},
		{"uniques/dupes", &Uniques{Str: []string{"a", "a"}, Int: []int{1, 1},
			Ptrs: []*string{&one, &one}, Items: []uniqueItem{{Name: "a"}, {Name: "a"}},
			Map: map[string]int{"a": 1, "b": 1}, Single: "x", Other: "x"}},
		{"uniques/nil", &Uniques{}},

		{"containers/valid", &Containers{
			Items: []string{"ab", "cd"}, Nums: []int{1, 2},
			Ptrs: []*string{&one, &long}, Structs: []uniqueItem{{Name: "a"}},
			M: map[string]string{"ab": "x"}, MV: map[string][]int{"a": {1, 2}},
			Matrix: [][]int{{1, 2}, {3}},
		}},
		{"containers/invalid", &Containers{
			Items: []string{"", "a"}, Nums: []int{0, -1},
			Ptrs: []*string{nil, &empty}, Structs: []uniqueItem{{}},
			M: map[string]string{"a": ""}, MV: map[string][]int{"a": {0}},
			Matrix: [][]int{{0}, nil},
		}},
		{"containers/nil", &Containers{}},

		{"cross/valid", &Cross{A: 2, B: 1, C: "same", D: "same", E: "x", F: "y",
			G: "abcdef", H: "cd", I: "abcdef", Leave: &small}},
		{"cross/invalid", &Cross{A: 1, B: 2, C: "x", D: "y", E: "x", F: "x",
			G: "abc", H: "zz", I: "abc", Leave: &small}},
		{"cross/nil", &Cross{}},

		{"cond/all-true", &Conditionals{Type: "a", A: "x", B: "x", C: "x", D: "x", N: 1, M: "x"}},
		{"cond/all-false", &Conditionals{Type: "b"}},
		{"cond/mixed", &Conditionals{Type: "b", A: "x", C: "x"}},
		{"cond/excluded", &Conditionals{Type: "a", G: "x", H: "x", I: "x", J: "x"}},

		{"nested/zero", &Nested{}},
		{"nested/valid", &Nested{Inner: Inner{Street: "s", Zip: "12345"},
			Ptr: &Inner{Street: "s", Zip: "12345"}, NoTag: Inner{Street: "s", Zip: "12345"},
			Deep: Deep{In: Inner{Street: "s", Zip: "12345"}}}},
		{"nested/invalid", &Nested{Inner: Inner{}, Ptr: &Inner{}, NoTag: Inner{}, Deep: Deep{}}},
		{"nested/nil-ptr", &Nested{Inner: Inner{Street: "s", Zip: "12345"}}},

		{"times/future", &Times{T: time.Now().Add(time.Hour), Past: time.Now().Add(-time.Hour), Dur: time.Second, DurMax: time.Minute}},
		{"times/past", &Times{T: time.Now().Add(-time.Hour), Past: time.Now().Add(time.Hour), Dur: 0, DurMax: 3 * time.Hour}},

		{"ptrs/nil", &Ptrs{}},
		{"ptrs/empty-targets", &Ptrs{Required: &empty, OmitNil: &empty, OmitZero: &empty, Plain: &small}},

		{"byfield/zero", &ByField{}},
		{"byfield/set", &ByField{N: &num, E: &one, Z: &small, M: &one}},

		{"durations/valid", &Durations{D: 2 * time.Hour, E: time.Minute}},
		{"durations/invalid", &Durations{D: time.Minute, E: 2 * time.Hour}},

		{"eqignore/valid", &EqIgnore{A: "golang", B: "java"}},
		{"eqignore/invalid", &EqIgnore{A: "java", B: "GOLANG"}},
	}

	for _, c := range cases {
		compare(t, c.name, c.value)
	}
}

// ---------------------------------------------------------------------------
// randomised differential testing
// ---------------------------------------------------------------------------

var (
	stringPool = []string{
		"", "a", "ab", "abc", "abcd", "abcde", "foo", "foobar", "aafoob", "bar",
		"中文", "a中b", "🙂", "RED", "red", "Red", " x ", "x,y", "1.5", "123",
		"true", "false", "a@b.com", "user@example.com", "not-an-email",
		"http://example.com", "not a url", "127.0.0.1", "::1", "999.1.1.1",
		"192.168.0.0/24", "01:23:45:67:89:ab", "#fff", "#abcdef", "rgb(1,2,3)",
		"rgba(1,2,3,0.5)", "hsl(120,50%,50%)", "nope",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8", "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"{\"a\":1}", "{", "data:text/plain;base64,aGVsbG8=", "aGVsbG8=",
		"45.0", "-122.0", "123-45-6789", "+15551234567", "ABC", "CN", "XX",
		"USD", "zh-Hans-CN", "2024-01-02", "Asia/Shanghai", "1.2.3",
		"4111111111111111", strings.Repeat("a", 32), strings.Repeat("b", 64),
		"abcdefghij", "abcdefghijklmnop",
	}
	intPool   = []int64{0, 1, -1, 2, 5, 6, 100, -100}
	uintPool  = []uint64{0, 1, 2, 3, 5, 255, 65535}
	floatPool = []float64{0, 1.5, -1.5, 1.6, 100.25, -0.5}
	boolPool  = []bool{true, false}
	timePool  = []time.Time{{}, time.Now().Add(-time.Hour), time.Now().Add(time.Hour)}
	durPool   = []time.Duration{0, time.Second, time.Minute, 2 * time.Hour, -time.Hour}
)

func fill(rv reflect.Value, rnd *rand.Rand, depth int) {
	switch rv.Kind() {
	case reflect.String:
		rv.SetString(stringPool[rnd.Intn(len(stringPool))])
	case reflect.Bool:
		rv.SetBool(boolPool[rnd.Intn(len(boolPool))])
	case reflect.Interface:
		if rnd.Intn(3) == 0 {
			return // leave it nil
		}
		rv.Set(reflect.ValueOf(stringPool[rnd.Intn(len(stringPool))]))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv.SetInt(intPool[rnd.Intn(len(intPool))])
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv.SetUint(uintPool[rnd.Intn(len(uintPool))])
	case reflect.Float32, reflect.Float64:
		rv.SetFloat(floatPool[rnd.Intn(len(floatPool))])
	case reflect.Slice:
		switch rnd.Intn(4) {
		case 0:
			rv.Set(reflect.Zero(rv.Type()))
		default:
			n := rnd.Intn(4)
			s := reflect.MakeSlice(rv.Type(), n, n)
			for i := 0; i < n; i++ {
				fill(s.Index(i), rnd, depth+1)
			}
			rv.Set(s)
		}
	case reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			fill(rv.Index(i), rnd, depth+1)
		}
	case reflect.Map:
		if rnd.Intn(4) == 0 {
			rv.Set(reflect.Zero(rv.Type()))
			return
		}
		m := reflect.MakeMap(rv.Type())
		for i := 0; i < rnd.Intn(3); i++ {
			k := reflect.New(rv.Type().Key()).Elem()
			fill(k, rnd, depth+1)
			v := reflect.New(rv.Type().Elem()).Elem()
			fill(v, rnd, depth+1)
			m.SetMapIndex(k, v)
		}
		rv.Set(m)
	case reflect.Ptr:
		if rnd.Intn(3) == 0 {
			rv.Set(reflect.Zero(rv.Type()))
			return
		}
		p := reflect.New(rv.Type().Elem())
		fill(p.Elem(), rnd, depth+1)
		rv.Set(p)
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			rv.Set(reflect.ValueOf(timePool[rnd.Intn(len(timePool))]))
			return
		}
		if rv.Type() == reflect.TypeOf(time.Duration(0)) {
			rv.Set(reflect.ValueOf(durPool[rnd.Intn(len(durPool))]))
			return
		}
		if depth > 4 {
			return
		}
		for i := 0; i < rv.NumField(); i++ {
			if rv.Field(i).CanSet() {
				fill(rv.Field(i), rnd, depth+1)
			}
		}
	}
}

func fuzzType(t *testing.T, makeValue func() interface{}, iterations int) {
	t.Helper()
	rnd := rand.New(rand.NewSource(20240921))
	for i := 0; i < iterations; i++ {
		v := makeValue()
		fill(reflect.ValueOf(v).Elem(), rnd, 0)
		compare(t, fmt.Sprintf("%T#%d", v, i), v)
	}
}

func TestFuzzScalars(t *testing.T) { fuzzType(t, func() interface{} { return &Scalars{} }, 300) }
func TestFuzzUnicode(t *testing.T) { fuzzType(t, func() interface{} { return &Unicode{} }, 200) }
func TestFuzzFormats(t *testing.T) { fuzzType(t, func() interface{} { return &Formats{} }, 300) }
func TestFuzzContent(t *testing.T) { fuzzType(t, func() interface{} { return &Content{} }, 200) }
func TestFuzzOneOfs(t *testing.T)  { fuzzType(t, func() interface{} { return &OneOfs{} }, 200) }
func TestFuzzUniques(t *testing.T) { fuzzType(t, func() interface{} { return &Uniques{} }, 200) }
func TestFuzzContainers(t *testing.T) {
	fuzzType(t, func() interface{} { return &Containers{} }, 200)
}
func TestFuzzCross(t *testing.T)   { fuzzType(t, func() interface{} { return &Cross{} }, 200) }
func TestFuzzCond(t *testing.T)    { fuzzType(t, func() interface{} { return &Conditionals{} }, 300) }
func TestFuzzNested(t *testing.T)  { fuzzType(t, func() interface{} { return &Nested{} }, 200) }
func TestFuzzTimes(t *testing.T)   { fuzzType(t, func() interface{} { return &Times{} }, 100) }
func TestFuzzPtrs(t *testing.T)    { fuzzType(t, func() interface{} { return &Ptrs{} }, 300) }
func TestFuzzByField(t *testing.T) { fuzzType(t, func() interface{} { return &ByField{} }, 200) }
func TestFuzzDurations(t *testing.T) {
	fuzzType(t, func() interface{} { return &Durations{} }, 100)
}
func TestFuzzEqIgnore(t *testing.T) { fuzzType(t, func() interface{} { return &EqIgnore{} }, 100) }

func TestCurated2(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{"formats2/valid", &Formats2{
			URI: "https://example.com/x", Urn: "urn:isbn:0451450523",
			B64URL: "aGVsbG8=", B64Raw: "aGVsbG8", PrintAS: "abc",
			BoolStr: "true", ISBN10: "0306406152",
			EthChk:  "0x52908400098527886E0F7030069857D2E4169EE7",
			Btc32:   "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4",
			MongoCS: "mongodb://localhost:27017", CVE: "CVE-2020-1234",
			DNSLabel: "example", Luhn: "4111111111111111", Cron: "0 0 * * *",
			SpiceDB: "user123", Country3: "CHN", CountryN: "156", Subdiv: "CN-BJ",
			CurrNum: 156, CCAlias: "CN", EUAlias: "DE",
			Postcode: "SW1A 1AA", Post2: "SW1A 1AA", PostCountry: "GB",
			IP4Addr: "127.0.0.1", IPAddr: "127.0.0.1", Unix: "/tmp/x.sock",
		}},
		{"formats2/invalid", &Formats2{
			URI: ":://", Urn: "urn:", B64URL: "!!!", B64Raw: "a", PrintAS: "中文",
			BoolStr: "yes", ISBN10: "123", EthChk: "0xabc", Btc32: "bc1qqq",
			MongoCS: "nope", CVE: "CVE-1", DNSLabel: "Example", Luhn: "1234567890123456",
			Cron: "nope", SpiceDB: "!!", Country3: "XXXX", CountryN: "999", Subdiv: "ZZ-ZZ",
			CurrNum: 999, CCAlias: "XX", EUAlias: "CN", Postcode: "ZZZZ", Post2: "ZZZZ",
			PostCountry: "GB", IP4Addr: "not-an-ip", IPAddr: "nope", Unix: "nope",
		}},
		{"formats2/empty", &Formats2{}},

		{"structs/dive-keys-values", &Structs{
			Keys:  map[string]map[string]int{"ab": {"cd": 1}},
			Array: [2]string{"abc", "abcd"}, Only: Inner{Street: "s", Zip: "12345"},
			NoLvl: Inner{}, Def: "", OE: []string{"ab"},
		}},
		{"structs/dive-invalid", &Structs{
			Keys:  map[string]map[string]int{"ab": {"cd": 0}, "": {"": 1}},
			Array: [2]string{"ab", ""}, Only: Inner{}, NoLvl: Inner{Street: "s", Zip: "12345"},
			Def: "x", OE: []string{""},
		}},
		{"structs/zero", &Structs{}},

		{"crossstruct/valid", &CrossStruct{Ref: Inner{Street: "s", Zip: "12345"},
			Same: "s", Diff: "t", Num: 5, NumGte: 5, NumLt: 0, NumLte: 1, Other: OtherNested{Num: 1}}},
		{"crossstruct/invalid", &CrossStruct{Ref: Inner{Street: "s", Zip: "12345"},
			Same: "t", Diff: "s", Num: 1, NumGte: 0, NumLt: 2, NumLte: 2, Other: OtherNested{Num: 2}}},
		{"crossstruct/zero", &CrossStruct{}},

		{"formats3/valid", &Formats3{
			AlphaUni: "abc中文", AlnumUni: "ab1中文",
			CIDRv4: "192.168.0.0/24", CIDRv6: "2001:db8::/32",
			Host1123: "example.com", HSLA: "hsla(120,50%,50%,0.3)",
			HTML: "<p>x</p>", HTMLEnc: "&lt;p&gt;", URLEnc: "a%20b",
			IP6Addr: "::1", TCP4Addr: "127.0.0.1:80", TCP6Addr: "[::1]:80", TCPAddr: "127.0.0.1:80",
			UDP4Addr: "127.0.0.1:80", UDP6Addr: "[::1]:80", UDPAddr: "127.0.0.1:80",
			MD4: strings.Repeat("a", 32), RIPEMD128: strings.Repeat("b", 32),
			RIPEMD160: strings.Repeat("c", 40), SHA384: strings.Repeat("d", 96),
			SHA512: strings.Repeat("e", 128), TIGER128: strings.Repeat("f", 32),
			TIGER160: strings.Repeat("a", 40), TIGER192: strings.Repeat("b", 48),
			UUID3: "f47ac10b-58cc-3372-a567-0e02b2c3d479", UUID3Rfc: "f47ac10b-58cc-3372-a567-0e02b2c3d479",
			UUID4Rfc: "6ba7b810-9dad-41d1-80b4-00c04fd430c8", UUID5: "886313e1-3b8a-5372-9b90-0c9aee199e5d",
			UUID5Rfc: "886313e1-3b8a-5372-9b90-0c9aee199e5d", UUIDRfc: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			EU2: "DE", EU3: "DEU",
		}},
		{"formats3/invalid", &Formats3{
			AlphaUni: "-", AlnumUni: "-", CIDRv4: "nope", CIDRv6: "nope", Host1123: "-bad",
			HSLA: "nope", HTML: "nope", HTMLEnc: "nope", URLEnc: "a%zz",
			IP6Addr: "[", TCP4Addr: "999.1.1.1:99999", TCP6Addr: "[", TCPAddr: "999.1.1.1:99999",
			UDP4Addr: "999.1.1.1:99999", UDP6Addr: "[", UDPAddr: "999.1.1.1:99999",
			MD4: "nope", RIPEMD128: "nope", RIPEMD160: "nope", SHA384: "nope", SHA512: "nope",
			TIGER128: "nope", TIGER160: "nope", TIGER192: "nope",
			UUID3: "nope", UUID3Rfc: "nope", UUID4Rfc: "nope", UUID5: "nope",
			UUID5Rfc: "nope", UUIDRfc: "nope", EU2: "CN", EU3: "CHN",
		}},
		{"formats3/empty", &Formats3{}},

		{"ptrpath/nil", &PointerPaths{}},
		{"ptrpath/set", &PointerPaths{Inner: &Inner{Street: "same"},
			Same: "same", Diff: "other", Longer: "longer", Other: "x"}},
		{"ptrpath/differ", &PointerPaths{Inner: &Inner{Street: "same"},
			Same: "other", Diff: "same", Longer: "ab", Other: "x", Miss: "y"}},

		{"jsons/valid", &JSONs{Str: `{"a":1}`, Doc: []byte(`[1,2]`), Raw: NamedDoc(`"x"`)}},
		{"jsons/invalid", &JSONs{Str: `{`, Doc: []byte(`{`), Raw: NamedDoc(`{`)}},
		{"jsons/empty", &JSONs{}},

		{"embed/zero", &Embed{}},
		{"embed/valid", &Embed{Base: Base{Common: "x"}, PtrBase: &PtrBase{Note: "abc"},
			Named: Base{Common: "y"}}},
		{"embed/invalid", &Embed{Base: Base{}, PtrBase: &PtrBase{Note: "ab"},
			Named: Base{Common: ""}}},

		{"ifaces/nil", &Ifaces{}},
		{"ifaces/set", &Ifaces{Req: 5, Omit: "x", Def: "x", Plain: 1}},

		{"named/valid", &Named{Email: "user@example.com", Age: 30, Code: 156, Ratio: 50.5,
			MinS: "abcd", EqS: "abc", Pick: "red", Post: "SW1A 1AA", When: "2024-01-02",
			Sub: "same", Other: "same", Emails: []NamedStr{"user@example.com"}}},
		{"named/invalid", &Named{Email: "nope", Age: 200, Code: 999, Ratio: 200,
			MinS: "ab", EqS: "zzz", Pick: "pink", Post: "ZZZZ", When: "02/01/2024",
			Sub: "a", Other: "b", Emails: []NamedStr{"nope"}}},
		{"named/empty", &Named{}},
		// One side named, the other the builtin: eqfield compares the values,
		// while unique compares interface{} values and so never finds two
		// different types equal.
		{"named/mixed-same", &Named{Plain: "same", MixEq: "same", MixNe: "same", MixGt: "abcdef",
			PlainNum: 5, MixNum: 5, MixNumGt: 6, MixUniq: "same", MixWidth: 5, Wide: 5}},
		{"named/mixed-diff", &Named{Plain: "same", MixEq: "other", MixNe: "other", MixGt: "ab",
			PlainNum: 5, MixNum: 6, MixNumGt: 5, MixUniq: "same"}},

		{"numerics/valid", &Numerics{CodeStr: "156", CodeInt: 156, CodeInt8: 4, CodeUint: 156,
			CodeEUInt: 276, CodeEUU64: 276, CurrInt: 156, CurrUint: 840, CurrU64: 392}},
		{"numerics/invalid", &Numerics{CodeStr: "abc", CodeInt: 999, CodeInt8: -1, CodeUint: 999,
			CodeEUInt: 840, CodeEUU64: 840, CurrInt: 999, CurrUint: 999, CurrU64: 999}},
		// the reference reduces country codes modulo 1000 and truncates large
		// unsigned currency codes straight down to int.
		{"numerics/modulo", &Numerics{CodeInt: 10156, CodeUint: 10156, CodeEUInt: 10276,
			CodeEUU64: 10276, CurrU64: math.MaxUint64}},
		// unsigned wrap-around: the two conversions disagree on these values, so
		// they pin down modulo-before-cast (country codes) versus cast-only
		// (currency codes).
		{"numerics/uint-wrap", &Numerics{CodeU64: math.MaxUint64 - 459, CurrU64: math.MaxUint64 - 459,
			CodeEUU64: math.MaxUint64 - 339}},
		{"numerics/empty", &Numerics{}},
	}
	for _, c := range cases {
		compare(t, c.name, c.value)
	}
}

func TestFuzzFormats2(t *testing.T) { fuzzType(t, func() interface{} { return &Formats2{} }, 300) }
func TestFuzzFormats3(t *testing.T) { fuzzType(t, func() interface{} { return &Formats3{} }, 200) }
func TestFuzzNumerics(t *testing.T) { fuzzType(t, func() interface{} { return &Numerics{} }, 200) }
func TestFuzzNamed(t *testing.T)    { fuzzType(t, func() interface{} { return &Named{} }, 200) }
func TestFuzzEmbed(t *testing.T)    { fuzzType(t, func() interface{} { return &Embed{} }, 200) }
func TestFuzzPtrPaths(t *testing.T) {
	fuzzType(t, func() interface{} { return &PointerPaths{} }, 200)
}
func TestFuzzIfaces(t *testing.T)   { fuzzType(t, func() interface{} { return &Ifaces{} }, 200) }
func TestFuzzStructs(t *testing.T)  { fuzzType(t, func() interface{} { return &Structs{} }, 200) }
func TestFuzzCrossStruct(t *testing.T) {
	fuzzType(t, func() interface{} { return &CrossStruct{} }, 200)
}

// TestFilesystem exercises the file/dir/image tags against real paths.
func TestFilesystem(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	png := filepath.Join(dir, "pic.png")
	// 1x1 transparent PNG
	if err := os.WriteFile(png, []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}, 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		value interface{}
	}{
		{"files/valid", &Files{File: file, Dir: sub, FilePath: file, DirPath: sub}},
		{"files/missing", &Files{File: dir, Dir: file, FilePath: "/nope/nope", DirPath: "/nope/nope"}},
		{"files/empty", &Files{}},
	}
	for _, c := range cases {
		compare(t, c.name, c.value)
	}

	// image takes byte content, which v10.25.0 always rejects (string only).
	data, err := os.ReadFile(png)
	if err != nil {
		t.Fatal(err)
	}
	compare(t, "files/image-bytes", &Files{Image: data})
	compare(t, "files/image-empty", &Files{})
}
