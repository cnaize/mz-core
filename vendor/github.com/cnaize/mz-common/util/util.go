package util

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"math/rand"
	"strings"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits

	maxInStrLen = 128
)

func RandString(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func GenerateXToken() string {
	return RandString(32)
}

func ParseInStr(str string) string {
	if len(str) > maxInStrLen {
		str = str[:maxInStrLen]
	}

	return str
}

func DecodeInStr(str string) string {
	data, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return fmt.Sprintf("Common: in text decode failed: %+v", err)
	}

	return string(data)
}

func EncodeOutStr(str string) string {
	return base64.URLEncoding.EncodeToString([]byte(str))
}

// FROM HERE: https://rosettacode.org/wiki/Strip_control_codes_and_extended_characters_from_a_string#Go
//
// Advanced Unicode normalization and filtering,
// see http://blog.golang.org/normalization and
// http://godoc.org/golang.org/x/text/unicode/norm for more
// details.
func NormalizeText(text string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	// The isOk filter is such that there is no need to chain to norm.NFC
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	// This Transformer could also trivially be applied as an io.Reader
	// or io.Writer filter to automatically do such filtering when reading
	// or writing data anywhere.
	text, _, _ = transform.String(t, text)
	text = strings.Trim(text, " ")
	return strings.ToLower(text)
}
