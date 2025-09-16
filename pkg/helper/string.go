package helper

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"time"
	"unicode"
)

func GenerateAPIKey() string {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		return err.Error()
	}

	apiKey := hex.EncodeToString(seed)

	return apiKey
}

func GeneratePassword() (string, error) {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=[]{}|<>/?~"

	b := make([]byte, 10)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}

	return string(b), nil
}

func ValidatePasswordStrength(password string) bool {
	var (
		upp, low, num, sym bool
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			low = true
		case unicode.IsUpper(char):
			upp = true
		case unicode.IsNumber(char):
			num = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
		default:
			return false
		}
	}

	if !upp || !low || !num || !sym {
		return false
	}

	return true
}

func ParseDate(layout, date string) error {
	_, err := time.Parse(layout, date)
	if err != nil {
		return err
	}

	return nil
}

func FormatStartTimeForSQL(date string) string {
	return date + " 00:00:00"
}

func FormatEndTimeForSQL(date string) string {
	return date + " 24:00:00"
}

func FormatWIB(currentTime time.Time) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	currentTime = currentTime.In(loc)

	return currentTime.Format("2006-01-02 15:04:05 MST")
}

func IsValidTemplateHeader(x []string, str string) bool {
	// iterate using the for loop
	for i := 0; i < len(x); i++ {
		// check
		if x[i] == str {
			return true
		}
	}
	return false
}

func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}
