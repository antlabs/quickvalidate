// Ported from github.com/go-playground/validator/v10 (baked_in.go, v10.25.0).
// Copyright (c) 2015 Dean Karn. MIT License — see LICENSE.
//
// 本文件是 go-playground/validator v10.25.0 中下列内置校验函数的「逐字等价」移植：
// 语义（边界、长度限制、空串行为、正则、DNS/网络解析）与参考实现完全一致，仅把
// reflect 驱动的 FieldLevel 参数换成具体类型（string / uint64）。
//
// 参考实现中依赖的私有辅助函数（isFileURL、isIP4Addr、isIP6Addr）以及它们依赖的
// 包级 regex 变量（hostnameRegexRFC952/hostnameRegexRFC1123/fqdnRegexRFC1123，
// 由本包的 regexes.go 提供）一并按原样保留。
//
// 与参考实现的唯一差别（有意为之，均不影响语义）：
//   - 参考实现中 field.Kind() 不是 reflect.String 时会 panic("Bad field type %T")；
//     这里参数已经是 string，不存在该分支。
//   - isPort 在参考实现里读的是 fl.Field().Uint()，对任何无符号整型生效；这里统一
//     收窄为 uint64（对 int/负数是参考实现本身就会 panic 的非法输入）。
package validators

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	urn "github.com/leodido/go-urn"
)

// IsURI is the validation function for validating if the current field's value is a valid URI.
func IsURI(s string) bool {
	// checks needed as of Go 1.6 because of change https://github.com/golang/go/commit/617c93ce740c3c3cc28cdd1a0d712be183d0b328#diff-6c2d018290e298803c0c9419d8739885L195
	// emulate browser and strip the '#' suffix prior to validation. see issue-#237
	if i := strings.Index(s, "#"); i > -1 {
		s = s[:i]
	}

	if len(s) == 0 {
		return false
	}

	_, err := url.ParseRequestURI(s)

	return err == nil
}

// IsFileURL is the helper function for validating if the `path` valid file URL as per RFC8089
func IsFileURL(path string) bool {
	if !strings.HasPrefix(path, "file:/") {
		return false
	}
	_, err := url.ParseRequestURI(path)
	return err == nil
}

// IsURL is the validation function for validating if the current field's value is a valid URL.
func IsURL(s string) bool {
	s = strings.ToLower(s)

	if len(s) == 0 {
		return false
	}

	if IsFileURL(s) {
		return true
	}

	url, err := url.Parse(s)
	if err != nil || url.Scheme == "" {
		return false
	}

	if url.Host == "" && url.Fragment == "" && url.Opaque == "" {
		return false
	}

	return true
}

// IsHTTPURL is the validation function for validating if the current field's value is a valid HTTP(s) URL.
func IsHTTPURL(s string) bool {
	if !IsURL(s) {
		return false
	}

	s = strings.ToLower(s)

	url, err := url.Parse(s)
	if err != nil || url.Host == "" {
		return false
	}

	return url.Scheme == "http" || url.Scheme == "https"
}

// IsUrnRFC2141 is the validation function for validating if the current field's value is a valid URN as per RFC 2141.
func IsUrnRFC2141(s string) bool {
	_, match := urn.Parse([]byte(s))

	return match
}

// IsMAC is the validation function for validating if the field's value is a valid MAC address.
func IsMAC(s string) bool {
	_, err := net.ParseMAC(s)

	return err == nil
}

// IsCIDRv4 is the validation function for validating if the field's value is a valid v4 CIDR address.
func IsCIDRv4(s string) bool {
	ip, n, err := net.ParseCIDR(s)

	return err == nil && ip.To4() != nil && n.IP.Equal(ip)
}

// IsCIDRv6 is the validation function for validating if the field's value is a valid v6 CIDR address.
func IsCIDRv6(s string) bool {
	ip, _, err := net.ParseCIDR(s)

	return err == nil && ip.To4() == nil
}

// IsCIDR is the validation function for validating if the field's value is a valid v4 or v6 CIDR address.
func IsCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)

	return err == nil
}

// IsIPv4 is the validation function for validating if a value is a valid v4 IP address.
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)

	return ip != nil && ip.To4() != nil
}

// IsIPv6 is the validation function for validating if the field's value is a valid v6 IP address.
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)

	return ip != nil && ip.To4() == nil
}

// IsIP is the validation function for validating if the field's value is a valid v4 or v6 IP address.
func IsIP(s string) bool {
	ip := net.ParseIP(s)

	return ip != nil
}

// IsHostnameRFC952 is the validation function for validating if the field's value is a valid RFC 952 hostname.
func IsHostnameRFC952(s string) bool {
	return hostnameRegexRFC952().MatchString(s)
}

// IsHostnameRFC1123 is the validation function for validating if the field's value is a valid RFC 1123 hostname.
func IsHostnameRFC1123(s string) bool {
	return hostnameRegexRFC1123().MatchString(s)
}

// IsFQDN is the validation function for validating if the field's value is a valid FQDN.
func IsFQDN(s string) bool {
	if s == "" {
		return false
	}

	return fqdnRegexRFC1123().MatchString(s)
}

// IsHostnamePort validates a <dns>:<port> combination for fields typically used for socket address.
func IsHostnamePort(s string) bool {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return false
	}
	// Port must be a iny <= 65535.
	if portNum, err := strconv.ParseInt(
		port, 10, 32,
	); err != nil || portNum > 65535 || portNum < 1 {
		return false
	}

	// If host is specified, it should match a DNS name
	if host != "" {
		return hostnameRegexRFC1123().MatchString(host)
	}
	return true
}

// IsPort validates if the current field's value represents a valid port
func IsPort(port uint64) bool {
	return port >= 1 && port <= 65535
}

// IsTCP4AddrResolvable is the validation function for validating if the field's value is a resolvable tcp4 address.
func IsTCP4AddrResolvable(s string) bool {
	if !IsIP4Addr(s) {
		return false
	}

	_, err := net.ResolveTCPAddr("tcp4", s)
	return err == nil
}

// IsTCP6AddrResolvable is the validation function for validating if the field's value is a resolvable tcp6 address.
func IsTCP6AddrResolvable(s string) bool {
	if !IsIP6Addr(s) {
		return false
	}

	_, err := net.ResolveTCPAddr("tcp6", s)

	return err == nil
}

// IsTCPAddrResolvable is the validation function for validating if the field's value is a resolvable tcp address.
func IsTCPAddrResolvable(s string) bool {
	if !IsIP4Addr(s) && !IsIP6Addr(s) {
		return false
	}

	_, err := net.ResolveTCPAddr("tcp", s)

	return err == nil
}

// IsUDP4AddrResolvable is the validation function for validating if the field's value is a resolvable udp4 address.
func IsUDP4AddrResolvable(s string) bool {
	if !IsIP4Addr(s) {
		return false
	}

	_, err := net.ResolveUDPAddr("udp4", s)

	return err == nil
}

// IsUDP6AddrResolvable is the validation function for validating if the field's value is a resolvable udp6 address.
func IsUDP6AddrResolvable(s string) bool {
	if !IsIP6Addr(s) {
		return false
	}

	_, err := net.ResolveUDPAddr("udp6", s)

	return err == nil
}

// IsUDPAddrResolvable is the validation function for validating if the field's value is a resolvable udp address.
func IsUDPAddrResolvable(s string) bool {
	if !IsIP4Addr(s) && !IsIP6Addr(s) {
		return false
	}

	_, err := net.ResolveUDPAddr("udp", s)

	return err == nil
}

// IsIP4AddrResolvable is the validation function for validating if the field's value is a resolvable ip4 address.
func IsIP4AddrResolvable(s string) bool {
	if !IsIPv4(s) {
		return false
	}

	_, err := net.ResolveIPAddr("ip4", s)

	return err == nil
}

// IsIP6AddrResolvable is the validation function for validating if the field's value is a resolvable ip6 address.
func IsIP6AddrResolvable(s string) bool {
	if !IsIPv6(s) {
		return false
	}

	_, err := net.ResolveIPAddr("ip6", s)

	return err == nil
}

// IsIPAddrResolvable is the validation function for validating if the field's value is a resolvable ip address.
func IsIPAddrResolvable(s string) bool {
	if !IsIP(s) {
		return false
	}

	_, err := net.ResolveIPAddr("ip", s)

	return err == nil
}

// IsUnixAddrResolvable is the validation function for validating if the field's value is a resolvable unix address.
func IsUnixAddrResolvable(s string) bool {
	_, err := net.ResolveUnixAddr("unix", s)

	return err == nil
}

// IsIP4Addr is the validation function for validating if the field's value is a valid v4 address with a port.
func IsIP4Addr(s string) bool {
	val := s

	if idx := strings.LastIndex(val, ":"); idx != -1 {
		val = val[0:idx]
	}

	ip := net.ParseIP(val)

	return ip != nil && ip.To4() != nil
}

// IsIP6Addr is the validation function for validating if the field's value is a valid v6 address with a port.
func IsIP6Addr(s string) bool {
	val := s

	if idx := strings.LastIndex(val, ":"); idx != -1 {
		if idx != 0 && val[idx-1:idx] == "]" {
			val = val[1 : idx-1]
		}
	}

	ip := net.ParseIP(val)

	return ip != nil && ip.To4() == nil
}
