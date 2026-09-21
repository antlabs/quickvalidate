// Concrete-typed, reflect-free ports of validators from
// github.com/go-playground/validator/v10 v10.25.0 (baked_in.go).
// Copyright (c) 2015 Dean Karn. MIT License.
//
// Porting rules applied:
//   - `fl.Field().String()` became a `string` parameter.
//   - `fl.Param()` became an additional `string` parameter (see IsSpiceDB).
//   - `fieldMatchesRegexByStringerValOrString(regex, fl)` collapsed to
//     `regex().MatchString(s)`, which is exactly the branch it takes for
//     reflect.String fields (the only kind these ports support).
//   - Private helpers the reference relies on (digitsHaveLuhnChecksum,
//     isISBN10/isISBN13 bodies, the inlined bech32 decoder) are ported too.
//   - No reflection anywhere.
package validators

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/crypto/sha3"
)

// ---------------------------------------------------------------------------
// UUID / ULID
// ---------------------------------------------------------------------------

// IsUUID5 reports whether str is a valid v5 UUID (baked_in.go:538, isUUID5).
func IsUUID5(str string) bool {
	return uUID5Regex().MatchString(str)
}

// IsUUID4 reports whether str is a valid v4 UUID (baked_in.go:543, isUUID4).
func IsUUID4(str string) bool {
	return uUID4Regex().MatchString(str)
}

// IsUUID3 reports whether str is a valid v3 UUID (baked_in.go:548, isUUID3).
func IsUUID3(str string) bool {
	return uUID3Regex().MatchString(str)
}

// IsUUID reports whether str is a valid UUID of any version (baked_in.go:553, isUUID).
func IsUUID(str string) bool {
	return uUIDRegex().MatchString(str)
}

// IsUUID5RFC4122 reports whether str is a valid RFC4122 v5 UUID (baked_in.go:558, isUUID5RFC4122).
func IsUUID5RFC4122(str string) bool {
	return uUID5RFC4122Regex().MatchString(str)
}

// IsUUID4RFC4122 reports whether str is a valid RFC4122 v4 UUID (baked_in.go:563, isUUID4RFC4122).
func IsUUID4RFC4122(str string) bool {
	return uUID4RFC4122Regex().MatchString(str)
}

// IsUUID3RFC4122 reports whether str is a valid RFC4122 v3 UUID (baked_in.go:568, isUUID3RFC4122).
func IsUUID3RFC4122(str string) bool {
	return uUID3RFC4122Regex().MatchString(str)
}

// IsUUIDRFC4122 reports whether str is a valid RFC4122 UUID of any version (baked_in.go:573, isUUIDRFC4122).
func IsUUIDRFC4122(str string) bool {
	return uUIDRFC4122Regex().MatchString(str)
}

// IsULID reports whether str is a valid ULID (baked_in.go:578, isULID).
func IsULID(str string) bool {
	return uLIDRegex().MatchString(str)
}

// ---------------------------------------------------------------------------
// Hashes
// ---------------------------------------------------------------------------

// IsMD4 reports whether str is a valid MD4 hash (baked_in.go:583, isMD4).
func IsMD4(str string) bool {
	return md4Regex().MatchString(str)
}

// IsMD5 reports whether str is a valid MD5 hash (baked_in.go:588, isMD5).
func IsMD5(str string) bool {
	return md5Regex().MatchString(str)
}

// IsSHA256 reports whether str is a valid SHA256 hash (baked_in.go:593, isSHA256).
func IsSHA256(str string) bool {
	return sha256Regex().MatchString(str)
}

// IsSHA384 reports whether str is a valid SHA384 hash (baked_in.go:598, isSHA384).
func IsSHA384(str string) bool {
	return sha384Regex().MatchString(str)
}

// IsSHA512 reports whether str is a valid SHA512 hash (baked_in.go:603, isSHA512).
func IsSHA512(str string) bool {
	return sha512Regex().MatchString(str)
}

// IsRIPEMD128 reports whether str is a valid RIPEMD128 hash (baked_in.go:608, isRIPEMD128).
func IsRIPEMD128(str string) bool {
	return ripemd128Regex().MatchString(str)
}

// IsRIPEMD160 reports whether str is a valid RIPEMD160 hash (baked_in.go:613, isRIPEMD160).
func IsRIPEMD160(str string) bool {
	return ripemd160Regex().MatchString(str)
}

// IsTIGER128 reports whether str is a valid TIGER128 hash (baked_in.go:618, isTIGER128).
func IsTIGER128(str string) bool {
	return tiger128Regex().MatchString(str)
}

// IsTIGER160 reports whether str is a valid TIGER160 hash (baked_in.go:623, isTIGER160).
func IsTIGER160(str string) bool {
	return tiger160Regex().MatchString(str)
}

// IsTIGER192 reports whether str is a valid TIGER192 hash (baked_in.go:628, isTIGER192).
func IsTIGER192(str string) bool {
	return tiger192Regex().MatchString(str)
}

// ---------------------------------------------------------------------------
// ISBN / ISSN
// ---------------------------------------------------------------------------

// IsISBN reports whether str is a valid v10 or v13 ISBN (baked_in.go:633, isISBN).
func IsISBN(str string) bool {
	return IsISBN10(str) || IsISBN13(str)
}

// IsISBN13 reports whether str is a valid v13 ISBN (baked_in.go:638, isISBN13).
func IsISBN13(str string) bool {
	s := strings.Replace(strings.Replace(str, "-", "", 4), " ", "", 4)

	if !iSBN13Regex().MatchString(s) {
		return false
	}

	var checksum int32
	var i int32

	factor := []int32{1, 3}

	for i = 0; i < 12; i++ {
		checksum += factor[i%2] * int32(s[i]-'0')
	}

	return (int32(s[12]-'0'))-((10-(checksum%10))%10) == 0
}

// IsISBN10 reports whether str is a valid v10 ISBN (baked_in.go:658, isISBN10).
func IsISBN10(str string) bool {
	s := strings.Replace(strings.Replace(str, "-", "", 3), " ", "", 3)

	if !iSBN10Regex().MatchString(s) {
		return false
	}

	var checksum int32
	var i int32

	for i = 0; i < 9; i++ {
		checksum += (i + 1) * int32(s[i]-'0')
	}

	if s[9] == 'X' {
		checksum += 10 * 10
	} else {
		checksum += 10 * int32(s[9]-'0')
	}

	return checksum%11 == 0
}

// IsISSN reports whether str is a valid ISSN (baked_in.go:682, isISSN).
func IsISSN(str string) bool {
	s := str

	if !iSSNRegex().MatchString(s) {
		return false
	}
	s = strings.ReplaceAll(s, "-", "")

	pos := 8
	checksum := 0

	for i := 0; i < 7; i++ {
		checksum += pos * int(s[i]-'0')
		pos--
	}

	if s[7] == 'X' {
		checksum += 10
	} else {
		checksum += int(s[7] - '0')
	}

	return checksum%11 == 0
}

// ---------------------------------------------------------------------------
// Ethereum / Bitcoin
// ---------------------------------------------------------------------------

// IsEthereumAddress reports whether address is a valid Ethereum address (baked_in.go:708, isEthereumAddress).
func IsEthereumAddress(address string) bool {
	return ethAddressRegex().MatchString(address)
}

// IsEthereumAddressChecksum reports whether address is a valid checksummed
// Ethereum address (baked_in.go:715, isEthereumAddressChecksum).
func IsEthereumAddressChecksum(address string) bool {
	if !ethAddressRegex().MatchString(address) {
		return false
	}
	// Checksum validation. Reference: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-55.md
	address = address[2:] // Skip "0x" prefix.
	h := sha3.NewLegacyKeccak256()
	// hash.Hash's io.Writer implementation says it never returns an error. https://golang.org/pkg/hash/#Hash
	_, _ = h.Write([]byte(strings.ToLower(address)))
	hash := hex.EncodeToString(h.Sum(nil))

	for i := 0; i < len(address); i++ {
		if address[i] <= '9' { // Skip 0-9 digits: they don't have upper/lower-case.
			continue
		}
		if hash[i] > '7' && address[i] >= 'a' || hash[i] <= '7' && address[i] <= 'F' {
			return false
		}
	}

	return true
}

// IsBitcoinAddress reports whether address is a valid base58 bitcoin address
// (baked_in.go:741, isBitcoinAddress).
func IsBitcoinAddress(address string) bool {
	if !btcAddressRegex().MatchString(address) {
		return false
	}

	alphabet := []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

	decode := [25]byte{}

	for _, n := range []byte(address) {
		d := bytes.IndexByte(alphabet, n)

		for i := 24; i >= 0; i-- {
			d += 58 * int(decode[i])
			decode[i] = byte(d % 256)
			d /= 256
		}
	}

	h := sha256.New()
	_, _ = h.Write(decode[:21])
	d := h.Sum([]byte{})
	h = sha256.New()
	_, _ = h.Write(d)

	validchecksum := [4]byte{}
	computedchecksum := [4]byte{}

	copy(computedchecksum[:], h.Sum(d[:0]))
	copy(validchecksum[:], decode[21:])

	return validchecksum == computedchecksum
}

// IsBitcoinBech32Address reports whether address is a valid bech32 bitcoin
// address (baked_in.go:778, isBitcoinBech32Address, including its inlined
// bech32 checksum decoder).
func IsBitcoinBech32Address(address string) bool {
	if !btcLowerAddressRegexBech32().MatchString(address) && !btcUpperAddressRegexBech32().MatchString(address) {
		return false
	}

	am := len(address) % 8

	if am == 0 || am == 3 || am == 5 {
		return false
	}

	address = strings.ToLower(address)

	alphabet := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	hr := []int{3, 3, 0, 2, 3} // the human readable part will always be bc
	addr := address[3:]
	dp := make([]int, 0, len(addr))

	for _, c := range addr {
		dp = append(dp, strings.IndexRune(alphabet, c))
	}

	ver := dp[0]

	if ver < 0 || ver > 16 {
		return false
	}

	if ver == 0 {
		if len(address) != 42 && len(address) != 62 {
			return false
		}
	}

	values := append(hr, dp...)

	GEN := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

	p := 1

	for _, v := range values {
		b := p >> 25
		p = (p&0x1ffffff)<<5 ^ v

		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				p ^= GEN[i]
			}
		}
	}

	if p != 1 {
		return false
	}

	b := uint(0)
	acc := 0
	mv := (1 << 5) - 1
	var sw []int

	for _, v := range dp[1 : len(dp)-6] {
		acc = (acc << 5) | v
		b += 5
		for b >= 8 {
			b -= 8
			sw = append(sw, (acc>>b)&mv)
		}
	}

	if len(sw) < 2 || len(sw) > 40 {
		return false
	}

	return true
}

// ---------------------------------------------------------------------------
// Misc formats
// ---------------------------------------------------------------------------

// IsBIC reports whether bicString is a valid Business Identifier Code (SWIFT
// code), defined in ISO 9362 (baked_in.go:2935, isIsoBicFormat).
func IsBIC(bicString string) bool {
	return bicRegex().MatchString(bicString)
}

// IsSemver reports whether semverString is a valid semver version, defined in
// Semantic Versioning 2.0.0 (baked_in.go:2942, isSemverFormat).
func IsSemver(semverString string) bool {
	return semverRegex().MatchString(semverString)
}

// IsCVE reports whether cveString is a valid CVE id, defined in CVE mitre org
// (baked_in.go:2949, isCveFormat).
func IsCVE(cveString string) bool {
	return cveRegex().MatchString(cveString)
}

// IsDnsRFC1035Label reports whether val is a valid dns RFC 1035 label, defined
// in RFC 1035 (baked_in.go:2958, isDnsRFC1035LabelFormat).
func IsDnsRFC1035Label(val string) bool {
	return dnsRegexRFC1035Label().MatchString(val)
}

// IsMongoDBObjectID reports whether val is a valid MongoDB ObjectID
// (baked_in.go:2987, isMongoDBObjectId).
func IsMongoDBObjectID(val string) bool {
	return mongodbIdRegex().MatchString(val)
}

// IsMongoDBConnectionString reports whether val is a valid MongoDB Connection
// String (baked_in.go:2993, isMongoDBConnectionString).
func IsMongoDBConnectionString(val string) bool {
	return mongodbConnectionRegex().MatchString(val)
}

// IsSpiceDB reports whether val is valid for use with Authzed SpiceDB in the
// indicated way (baked_in.go:2999, isSpiceDB).
//
// param is the tag parameter (`fl.Param()` in the reference): "permission",
// "type", "id" or "" (empty selects the "id" rule, just like the reference).
// An unrecognized parameter panics, exactly as the reference does.
func IsSpiceDB(val, param string) bool {
	switch param {
	case "permission":
		return spicedbPermissionRegex().MatchString(val)
	case "type":
		return spicedbTypeRegex().MatchString(val)
	case "id", "":
		return spicedbIDRegex().MatchString(val)
	}

	panic("Unrecognized parameter: " + param)
}

// IsCreditCard reports whether val is a valid credit card number
// (baked_in.go:3016, isCreditCard).
func IsCreditCard(val string) bool {
	var creditCard bytes.Buffer
	segments := strings.Split(val, " ")
	for _, segment := range segments {
		if len(segment) < 3 {
			return false
		}
		creditCard.WriteString(segment)
	}

	ccDigits := strings.Split(creditCard.String(), "")
	size := len(ccDigits)
	if size < 12 || size > 19 {
		return false
	}

	return digitsHaveLuhnChecksum(ccDigits)
}

// IsLuhnChecksum reports whether str has a valid Luhn checksum
// (baked_in.go:3037, hasLuhnChecksum, string branch).
//
// The reference also accepts signed/unsigned integer kinds by formatting them
// with strconv; those branches are not applicable to a concrete-typed port and
// are intentionally omitted.
func IsLuhnChecksum(str string) bool {
	size := len(str)
	if size < 2 { // there has to be at least one digit that carries a meaning + the checksum
		return false
	}
	digits := strings.Split(str, "")
	return digitsHaveLuhnChecksum(digits)
}

// IsCron reports whether cronString is a valid cron expression
// (baked_in.go:3059, isCron).
func IsCron(cronString string) bool {
	return cronRegex().MatchString(cronString)
}

// digitsHaveLuhnChecksum returns true if and only if the last element of the
// given digits slice is the Luhn checksum of the previous elements
// (baked_in.go:2964, digitsHaveLuhnChecksum). Private helper, ported as-is.
func digitsHaveLuhnChecksum(digits []string) bool {
	size := len(digits)
	sum := 0
	for i, digit := range digits {
		value, err := strconv.Atoi(digit)
		if err != nil {
			return false
		}
		if size%2 == 0 && i%2 == 0 || size%2 == 1 && i%2 == 1 {
			v := value * 2
			if v >= 10 {
				sum += 1 + (v % 10)
			} else {
				sum += v
			}
		} else {
			sum += value
		}
	}
	return (sum % 10) == 0
}

// ---------------------------------------------------------------------------
// Image / file system
// ---------------------------------------------------------------------------

// imageMimeTypes is the set of MIME types the reference `image` validator
// accepts (baked_in.go:1577).
var imageMimeTypes = map[string]bool{
	"image/bmp":                true,
	"image/cis-cod":            true,
	"image/gif":                true,
	"image/ief":                true,
	"image/jpeg":               true,
	"image/jp2":                true,
	"image/jpx":                true,
	"image/jpm":                true,
	"image/pipeg":              true,
	"image/png":                true,
	"image/svg+xml":            true,
	"image/tiff":               true,
	"image/webp":               true,
	"image/x-cmu-raster":       true,
	"image/x-cmx":              true,
	"image/x-icon":             true,
	"image/x-portable-anymap":  true,
	"image/x-portable-bitmap":  true,
	"image/x-portable-graymap": true,
	"image/x-portable-pixmap":  true,
	"image/x-rgb":              true,
	"image/x-xbitmap":          true,
	"image/x-xpixmap":          true,
	"image/x-xwindowdump":      true,
}

// IsImage reports whether data holds image content, sniffed from the bytes
// themselves with github.com/gabriel-vasile/mimetype.
//
// This is the "content" half of the reference `image` validator: the reference
// (baked_in.go:1576, isImage) only implements the reflect.String branch and
// sniffs the *content* of the file it points at (mimetype.DetectReader),
// independent of the file name/extension; feeding the very same bytes through
// mimetype.Detect yields the identical MIME type and therefore the identical
// verdict. See also IsImageFile for the path branch. Note that the reference's
// `image` tag always reports false for a []byte field (no Slice case in its
// kind switch), so IsImage is a deliberate superset for in-memory data.
func IsImage(data []byte) bool {
	mime := mimetype.Detect(data)
	return imageMimeTypes[mime.String()]
}

// IsImageFile reports whether path is the path to an existing image file.
//
// This is a verbatim port of the reference `image` validator's path branch
// (baked_in.go:1576, isImage, case reflect.String): os.Stat, reject
// directories, open, mimetype.DetectReader on the file's content and check the
// detected MIME type against the accepted set. Detection is content based, not
// extension based.
func IsImageFile(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return false
	}

	if fileInfo.IsDir() {
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return false
	}

	if _, ok := imageMimeTypes[mime.String()]; ok {
		return true
	}

	return false
}

// IsFile reports whether path is an existing file path
// (baked_in.go:1559, isFile, case reflect.String).
func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !fileInfo.IsDir()
}

// IsFilePath reports whether path is a valid file path
// (baked_in.go:1637, isFilePath, case reflect.String).
func IsFilePath(path string) bool {
	// Not valid if it is a directory.
	if IsDir(path) {
		return false
	}
	// If it exists, it obviously is valid.
	// This is done first to avoid code duplication and unnecessary additional logic.
	if IsFile(path) {
		return true
	}

	// It does not exist but may still be a valid filepath.
	// Every OS allows for whitespace, but none
	// let you use a file with no filename (to my knowledge).
	// Unless you're dealing with raw inodes, but I digress.
	if strings.TrimSpace(path) == "" {
		return false
	}
	// We make sure it isn't a directory.
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		switch t := err.(type) {
		case *fs.PathError:
			if t.Err == syscall.EINVAL {
				// It's definitely an invalid character in the filepath.
				return false
			}
			// It could be a permission error, a does-not-exist error, etc.
			// Out-of-scope for this validation, though.
			return true
		default:
			// Something went *seriously* wrong.
			panic(err)
		}
	}

	// Unreachable for a string path: os.Stat just succeeded, so the path is
	// either a directory (IsDir above) or not (IsFile above). Kept to mirror the
	// reference, which panics on this spot for any other field type.
	panic(fmt.Sprintf("Bad field type %T", path))
}

// IsDir reports whether path is an existing directory
// (baked_in.go:2630, isDir, case reflect.String).
func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

// IsDirPath reports whether path is a valid directory path
// (baked_in.go:2646, isDirPath, case reflect.String).
func IsDirPath(path string) bool {
	// If it exists, it obviously is valid.
	// This is done first to avoid code duplication and unnecessary additional logic.
	if IsDir(path) {
		return true
	}

	// It does not exist but may still be a valid path.
	// Every OS allows for whitespace, but none
	// let you use a dir with no name (to my knowledge).
	// Unless you're dealing with raw inodes, but I digress.
	if strings.TrimSpace(path) == "" {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		switch t := err.(type) {
		case *fs.PathError:
			if t.Err == syscall.EINVAL {
				// It's definitely an invalid character in the path.
				return false
			}
			// It could be a permission error, a does-not-exist error, etc.
			// Out-of-scope for this validation, though.
			// Lastly, we make sure it is a directory.
			if strings.HasSuffix(path, string(os.PathSeparator)) {
				return true
			} else {
				return false
			}
		default:
			// Something went *seriously* wrong.
			panic(err)
		}
	}
	// We repeat the check here to make sure it is an explicit directory in case the above os.Stat didn't trigger an error.
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return true
	} else {
		return false
	}
}
