package helper

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	randMath "math/rand"
)

const (
	maskChar = "*"
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

func GenerateTrx(productInitial string) string {
	now := time.Now()
	micro := now.Nanosecond() / 1000
	random := randMath.Intn(100)

	return fmt.Sprintf(
		"%s-%s-%06d-%02d",
		productInitial,
		now.Format("20060102-150405"),
		micro,
		random,
	)
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

func ValidateDateYYYYMMDD(value string) error {
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		return errors.New("timestamp must be in YYYY-MM-DD format")
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

func MaskingHead(value string, num int) string {
	if len(value) <= num {
		return value
	}

	head := value[:len(value)-num]
	tail := value[len(value)-num:]

	mask := strings.Repeat(maskChar, len(head))

	return mask + tail
}

func MaskingMiddle(value string) string {
	totalLen := len(value)

	if totalLen <= 4 {
		return value
	}

	maskLen := 3
	if totalLen <= 6 {
		maskLen = 1
	} else if totalLen <= 8 {
		maskLen = 2
	}

	remain := totalLen - maskLen
	headLen := remain / 2
	tailLen := remain - headLen

	head := value[:headLen]
	tail := value[totalLen-tailLen:]
	mask := strings.Repeat(maskChar, maskLen)

	return head + mask + tail
}
