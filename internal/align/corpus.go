// Package align holds the differential test corpus: structs with `validate`
// tags plus the values they are exercised with. The generated validators must
// report exactly the same failures as go-playground/validator for every case.
package align

import "time"

//go:generate quickvalidate -i . -o validate_gen.go -pkg align

type Scalars struct {
	Req      string         `validate:"required"`
	Email    string         `validate:"email"`
	MinS     string         `validate:"min=3"`
	MaxS     string         `validate:"max=5"`
	LenS     string         `validate:"len=4"`
	GtI      int            `validate:"gt=5"`
	GteI     int8           `validate:"gte=5"`
	LtU      uint16         `validate:"lt=5"`
	LteU     uint           `validate:"lte=5"`
	EqI      int            `validate:"eq=5"`
	NeI      int            `validate:"ne=5"`
	EqS      string         `validate:"eq=abc"`
	NeS      string         `validate:"ne=abc"`
	EqF      float64        `validate:"eq=1.5"`
	GtF      float32        `validate:"gt=1.5"`
	B        bool           `validate:"eq=true"`
	OmitMin  string         `validate:"omitempty,min=3"`
	OmitPtr  *string        `validate:"omitempty,min=3"`
	PlainPtr *string        `validate:"min=3"`
	NilPtr   *int           `validate:"required"`
	SliceLen []int          `validate:"min=2,max=3"`
	MapLen   map[string]int `validate:"len=2"`
}

type Unicode struct {
	Min   string `validate:"min=3"`
	Gt    string `validate:"gt=2"`
	Max   string `validate:"max=3"`
	Len   string `validate:"len=2"`
	Alpha string `validate:"alpha"`
	Alnum string `validate:"alphanum"`
}

type Formats struct {
	Email    string `validate:"email"`
	URL      string `validate:"url"`
	HTTPURL  string `validate:"http_url"`
	UUID     string `validate:"uuid"`
	UUID4    string `validate:"uuid4"`
	ULID     string `validate:"ulid"`
	HexColor string `validate:"hexcolor"`
	RGB      string `validate:"rgb"`
	RGBA     string `validate:"rgba"`
	HSL      string `validate:"hsl"`
	Color    string `validate:"iscolor"`
	ColorOr  string `validate:"rgb|rgba"`
	IP       string `validate:"ip"`
	IPv4     string `validate:"ipv4"`
	IPv6     string `validate:"ipv6"`
	CIDR     string `validate:"cidr"`
	MAC      string `validate:"mac"`
	Hostname string `validate:"hostname"`
	FQDN     string `validate:"fqdn"`
	Port     uint16 `validate:"port"`
	HostPort string `validate:"hostname_port"`
	JSON     string `validate:"json"`
	JWT      string `validate:"jwt"`
	Base64   string `validate:"base64"`
	Base32   string `validate:"base32"`
	DataURI  string `validate:"datauri"`
	Lat      string `validate:"latitude"`
	Lng      string `validate:"longitude"`
	SSN      string `validate:"ssn"`
	E164     string `validate:"e164"`
	Lower    string `validate:"lowercase"`
	Upper    string `validate:"uppercase"`
	ASCII    string `validate:"ascii"`
	Multi    string `validate:"multibyte"`
	Hex      string `validate:"hexadecimal"`
	Numeric  string `validate:"numeric"`
	Number   string `validate:"number"`
	Semver   string `validate:"semver"`
	BIC      string `validate:"bic"`
	MongoID  string `validate:"mongodb"`
	MD5      string `validate:"md5"`
	SHA256   string `validate:"sha256"`
	ISBN     string `validate:"isbn"`
	ISBN13   string `validate:"isbn13"`
	ISSN     string `validate:"issn"`
	Eth      string `validate:"eth_addr"`
	BTC      string `validate:"btc_addr"`
	Country  string `validate:"iso3166_1_alpha2"`
	Currency string `validate:"iso4217"`
	Lang     string `validate:"bcp47_language_tag"`
	Card     string `validate:"credit_card"`
	TZ       string `validate:"timezone"`
	DT       string `validate:"datetime=2006-01-02"`
}

type Content struct {
	Contains  string `validate:"contains=foo"`
	ContainsA string `validate:"containsany=abc"`
	ContainsR string `validate:"containsrune=x"`
	Excludes  string `validate:"excludes=foo"`
	ExcludesA string `validate:"excludesall=abc"`
	ExcludesR string `validate:"excludesrune=x"`
	Starts    string `validate:"startswith=ab"`
	Ends      string `validate:"endswith=yz"`
	NotStarts string `validate:"startsnotwith=ab"`
	NotEnds   string `validate:"endsnotwith=yz"`
}

type OneOfs struct {
	S  string `validate:"oneof=red green blue"`
	CI string `validate:"oneofci=Red Green"`
	I  int    `validate:"oneof=1 2 3"`
	U  uint8  `validate:"oneof=1 2 3"`
}

type Uniques struct {
	Str    []string       `validate:"unique"`
	Int    []int          `validate:"unique"`
	Ptrs   []*string      `validate:"unique"`
	Items  []uniqueItem   `validate:"unique=Name"`
	Map    map[string]int `validate:"unique"`
	Single string         `validate:"unique=Other"`
	Other  string
}

type uniqueItem struct {
	Name string
	Tag  string
}

type Containers struct {
	Items   []string          `validate:"required,min=1,dive,required,min=2"`
	Nums    []int             `validate:"dive,gt=0"`
	Ptrs    []*string         `validate:"dive,min=2"`
	Structs []uniqueItem      `validate:"dive,structonly"`
	M       map[string]string `validate:"dive,keys,min=2,endkeys,required"`
	MV      map[string][]int  `validate:"dive,gt=0"`
	Matrix  [][]int           `validate:"dive,dive,gt=0"`
}

type Cross struct {
	A     int    `validate:"gtfield=B"`
	B     int    `validate:"ltfield=A"`
	C     string `validate:"eqfield=D"`
	D     string
	E     string `validate:"nefield=F"`
	F     string
	G     string `validate:"fieldcontains=H"`
	H     string
	I     string `validate:"fieldexcludes=H"`
	Leave *int   `validate:"omitempty,gtfield=A"`
	GteF  int    `validate:"gtefield=B"`
	LteF  int    `validate:"ltefield=B"`
}

type Conditionals struct {
	Type string `validate:"oneof=a b"`
	A    string `validate:"required_if=Type a"`
	B    string `validate:"required_unless=Type a"`
	C    string `validate:"required_with=A"`
	D    string `validate:"required_with_all=A B"`
	E    string `validate:"required_without=A"`
	F    string `validate:"required_without_all=A B"`
	G    string `validate:"excluded_if=Type a"`
	H    string `validate:"excluded_unless=Type a"`
	I    string `validate:"excluded_with=A"`
	J    string `validate:"excluded_with_all=A B"`
	K    string `validate:"excluded_without=A"`
	L    string `validate:"excluded_without_all=A B"`
	N    int    `validate:"required_if=Type a"`
	M    string `validate:"skip_unless=Type a"`
}

type Nested struct {
	Inner Inner  `validate:"required"`
	Ptr   *Inner `validate:"omitempty"`
	NoTag Inner
	Deep  Deep
}

type Inner struct {
	Street string `validate:"required"`
	Zip    string `validate:"numeric,len=5"`
}

type Deep struct {
	In Inner
}

type Times struct {
	T      time.Time     `validate:"gt"`
	Past   time.Time     `validate:"lt"`
	Dur    time.Duration `validate:"gte=1s"`
	DurMax time.Duration `validate:"lt=2h"`
}

type Ptrs struct {
	Required  *string  `validate:"required"`
	OmitNil   *string  `validate:"omitnil,min=2"`
	OmitZero  *string  `validate:"omitzero,min=2"`
	IsDefault *string  `validate:"isdefault"`
	Plain     *int     `validate:"gt=5"`
	OmitEmp   *int     `validate:"omitempty,gt=5"`
	Double    **string `validate:"omitempty,min=2"`
}

type ByField struct {
	N   *int    `validate:"required"`
	E   *string `validate:"omitempty,email"`
	Z   *int    `validate:"omitzero"`
	M   *string `validate:"omitnil"`
	Def string  `validate:"isdefault"`
}

type Durations struct {
	D time.Duration `validate:"gt=1h"`
	E time.Duration `validate:"lte=30m"`
}

type EqIgnore struct {
	A string `validate:"eq_ignore_case=GoLang"`
	B string `validate:"ne_ignore_case=GoLang"`
}

type Formats2 struct {
	URI         string `validate:"uri"`
	Urn         string `validate:"urn_rfc2141"`
	B64URL      string `validate:"base64url"`
	B64Raw      string `validate:"base64rawurl"`
	PrintAS     string `validate:"printascii"`
	BoolStr     string `validate:"boolean"`
	ISBN10      string `validate:"isbn10"`
	EthChk      string `validate:"eth_addr_checksum"`
	Btc32       string `validate:"btc_addr_bech32"`
	MongoCS     string `validate:"mongodb_connection_string"`
	CVE         string `validate:"cve"`
	DNSLabel    string `validate:"dns_rfc1035_label"`
	Luhn        string `validate:"luhn_checksum"`
	Cron        string `validate:"cron"`
	SpiceDB     string `validate:"spicedb=id"`
	Country3    string `validate:"iso3166_1_alpha3"`
	CountryN    string `validate:"iso3166_1_alpha_numeric"`
	Subdiv      string `validate:"iso3166_2"`
	CurrNum     int    `validate:"iso4217_numeric"`
	CCAlias     string `validate:"country_code"`
	EUAlias     string `validate:"eu_country_code"`
	Postcode    string `validate:"postcode_iso3166_alpha2=GB"`
	Post2       string `validate:"postcode_iso3166_alpha2_field=PostCountry"`
	PostCountry string
	IP4Addr     string `validate:"ip4_addr"`
	IPAddr      string `validate:"ip_addr"`
	Unix        string `validate:"unix_addr"`
}

type Files struct {
	File     string `validate:"file"`
	Dir      string `validate:"dir"`
	FilePath string `validate:"filepath"`
	DirPath  string `validate:"dirpath"`
	Image    []byte `validate:"image"`
}

type Structs struct {
	Keys  map[string]map[string]int `validate:"dive,dive,keys,required,endkeys,gt=0"`
	Array [2]string                 `validate:"dive,min=3"`
	Only  Inner                     `validate:"structonly"`
	NoLvl Inner                     `validate:"nostructlevel"`
	Def   string                    `validate:"isdefault"`
	OE    []string                  `validate:"omitempty,dive,min=2"`
}

type CrossStruct struct {
	Ref    Inner  `validate:"required"`
	Same   string `validate:"eqcsfield=Ref.Street"`
	Diff   string `validate:"necsfield=Ref.Street"`
	Num    int    `validate:"gtcsfield=Other.Num"`
	NumGte int    `validate:"gtecsfield=Other.Num"`
	NumLt  int    `validate:"ltcsfield=Other.Num"`
	NumLte int    `validate:"ltecsfield=Other.Num"`
	Other  OtherNested
}

type OtherNested struct {
	Num int
}

// Formats3 covers the format tags that earlier revisions of this corpus left
// untested.
type Formats3 struct {
	AlphaUni  string `validate:"alphaunicode"`
	AlnumUni  string `validate:"alphanumunicode"`
	CIDRv4    string `validate:"cidrv4"`
	CIDRv6    string `validate:"cidrv6"`
	Host1123  string `validate:"hostname_rfc1123"`
	HSLA      string `validate:"hsla"`
	HTML      string `validate:"html"`
	HTMLEnc   string `validate:"html_encoded"`
	URLEnc    string `validate:"url_encoded"`
	IP6Addr   string `validate:"ip6_addr"`
	TCP4Addr  string `validate:"tcp4_addr"`
	TCP6Addr  string `validate:"tcp6_addr"`
	TCPAddr   string `validate:"tcp_addr"`
	UDP4Addr  string `validate:"udp4_addr"`
	UDP6Addr  string `validate:"udp6_addr"`
	UDPAddr   string `validate:"udp_addr"`
	MD4       string `validate:"md4"`
	RIPEMD128 string `validate:"ripemd128"`
	RIPEMD160 string `validate:"ripemd160"`
	SHA384    string `validate:"sha384"`
	SHA512    string `validate:"sha512"`
	TIGER128  string `validate:"tiger128"`
	TIGER160  string `validate:"tiger160"`
	TIGER192  string `validate:"tiger192"`
	UUID3     string `validate:"uuid3"`
	UUID3Rfc  string `validate:"uuid3_rfc4122"`
	UUID4Rfc  string `validate:"uuid4_rfc4122"`
	UUID5     string `validate:"uuid5"`
	UUID5Rfc  string `validate:"uuid5_rfc4122"`
	UUIDRfc   string `validate:"uuid_rfc4122"`
	EU2       string `validate:"iso3166_1_alpha2_eu"`
	EU3       string `validate:"iso3166_1_alpha3_eu"`
}

// Same-package named types. validator reaches them through reflection and reads
// the underlying type, so the generated code has to convert explicitly — Go does
// not convert implicitly on the way into a helper call.
type (
	NamedStr   string
	NamedInt   int
	NamedUint  uint16
	NamedFloat float64
)

type Named struct {
	Email  NamedStr     `validate:"required,email"`
	Age    NamedInt     `validate:"gte=0,lte=130"`
	Code   NamedUint    `validate:"iso3166_1_alpha_numeric"`
	Ratio  NamedFloat   `validate:"gte=0,lte=100"`
	MinS   NamedStr     `validate:"min=3"`
	EqS    NamedStr     `validate:"eq=abc"`
	Pick   NamedStr     `validate:"oneof=red green"`
	Post   NamedStr     `validate:"postcode_iso3166_alpha2=GB"`
	When   NamedStr     `validate:"datetime=2006-01-02"`
	Sub    NamedStr     `validate:"eqfield=Other"`
	Other  NamedStr
	Emails []NamedStr `validate:"dive,email"`

	// Cross field comparisons where one side is a named type and the other the
	// builtin: the reference reads both through reflection, so the values are
	// compared as their underlying type.
	MixEq    NamedStr `validate:"eqfield=Plain"`
	MixNe    NamedStr `validate:"nefield=Plain"`
	MixGt    NamedStr `validate:"gtfield=Plain"`
	MixNum   NamedInt `validate:"eqfield=PlainNum"`
	MixNumGt NamedInt `validate:"gtfield=PlainNum"`
	// unique compares through interface{}: two different types are never equal,
	// whatever the values are.
	MixUniq NamedStr `validate:"unique=Plain"`
	// Different reflect kinds compare as a failure, whatever the values are:
	// int8 and int64 are not the same kind to reflect.
	MixWidth int8 `validate:"eqfield=Wide"`
	Wide     int64
	Plain    string
	PlainNum int
}

// Embedded structs: validator keys them by type name, so the namespace is
// Embed.Base.Common and an untagged embedded struct is still recursed into.
type Base struct {
	Common string `validate:"required"`
}

type PtrBase struct {
	Note string `validate:"min=3"`
}

type Embed struct {
	Base
	*PtrBase
	Named Base   `validate:"required"`
	Skip  string `validate:"-"`
}

// Ifaces covers interface fields. The dynamic value is unknown at generation
// time, so only the tags that depend on the interface being nil can be emitted.
type Ifaces struct {
	Req   interface{} `validate:"required"`
	Omit  interface{} `validate:"omitempty"`
	Def   interface{} `validate:"isdefault"`
	Plain interface{}
}

// PointerPaths covers cross field paths that walk through a pointer: the
// reference reports a path through a nil pointer as "not found", which the
// negative comparisons treat as a pass.
type PointerPaths struct {
	Inner  *Inner `validate:"omitempty"`
	Same   string `validate:"eqfield=Inner.Street"`
	Diff   string `validate:"nefield=Inner.Street"`
	Longer string `validate:"gtfield=Inner.Street"`
	NoTag  *Inner
	Other  string `validate:"nefield=NoTag.Street"`
	Miss   string `validate:"nefield=Nope.Nope"`
}

// NamedDoc is a []byte type declared in this package.
type NamedDoc []byte

// JSONs covers the one string tag the reference also accepts as a byte slice.
type JSONs struct {
	Str string   `validate:"json"`
	Doc []byte   `validate:"json"`
	Raw NamedDoc `validate:"json"`
}

// Numerics pins down the tags whose reference implementation switches on the
// field kind: they accept string, signed and unsigned integer fields, so every
// one of those kinds must generate code that compiles.
type Numerics struct {
	CodeStr   string `validate:"iso3166_1_alpha_numeric"`
	CodeInt   int    `validate:"iso3166_1_alpha_numeric"`
	CodeInt8  int8   `validate:"iso3166_1_alpha_numeric"`
	CodeUint  uint32 `validate:"iso3166_1_alpha_numeric"`
	CodeU64   uint64 `validate:"iso3166_1_alpha_numeric"`
	CodeEUInt int    `validate:"iso3166_1_alpha_numeric_eu"`
	CodeEUU64 uint64 `validate:"iso3166_1_alpha_numeric_eu"`
	CurrInt   int16  `validate:"iso4217_numeric"`
	CurrUint  uint16 `validate:"iso4217_numeric"`
	CurrU64   uint64 `validate:"iso4217_numeric"`
}
