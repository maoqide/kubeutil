package utils

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"unicode/utf8"
)

// ReadFile read bytes from file
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	return b, err
}

// Int32Ptr convert int32 value to a pointer.
func Int32Ptr(i int32) *int32 {
	return &i
}

// Int64Ptr convert int64 value to a pointer.
func Int64Ptr(i int64) *int64 {
	return &i
}

// StringPtr convert string value to a pointer.
func StringPtr(s string) *string {
	return &s
}

func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func StringToBool(s string) (bool, error) {
	if s == "false" {
		return false, nil
	} else if s == "true" {
		return true, nil
	}
	return false, errors.New("params error: should be true or false")
}

// ToValidUTF8 treats s as UTF-8-encoded bytes and returns a copy with each run of bytes
// representing invalid UTF-8 replaced with the bytes in replacement, which may be empty.
func ToValidUTF8(s, replacement []byte) []byte {
	b := make([]byte, 0, len(s)+len(replacement))
	invalid := false // previous byte was from an invalid UTF-8 sequence
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			i++
			invalid = false
			b = append(b, byte(c))
			continue
		}
		_, wid := utf8.DecodeRune(s[i:])
		if wid == 1 {
			i++
			if !invalid {
				invalid = true
				b = append(b, replacement...)
			}
			continue
		}
		invalid = false
		b = append(b, s[i:i+wid]...)
		i += wid
	}
	return b
}

// Env get enviroment variables with default value
func Env(key, defaultVal string) string {
	val, exists := os.LookupEnv(key)
	if exists {
		return val
	}
	return defaultVal
}
