package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(
	secret string,
	minutesToExpired int,
	userId, companyId, roleId, quotaType uint,
	apiKey string,
) (string, error) {
	willExpiredAt := time.Now().Add(time.Duration(minutesToExpired) * time.Minute)

	claims := jwt.MapClaims{}
	claims["user_id"] = userId
	claims["company_id"] = companyId
	claims["role_id"] = roleId
	claims["api_key"] = apiKey
	claims["quota_type"] = quotaType
	claims["exp"] = willExpiredAt.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func ExtractClaimsFromJWT(tokenStr, secret string) (*jwt.MapClaims, error) {
	claims := &jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid JWT token")
	}

	return claims, nil
}

func extractUintClaim(claims *jwt.MapClaims, key string) (uint, error) {
	val, found := (*claims)[key]
	if !found {
		return 0, errors.New("missing key in claims: " + key)
	}

	switch v := val.(type) {
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case json.Number:
		num, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %v", key, err)
		}
		return uint(num), nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid %s format: %v", key, err)
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported claim type for %s: %T", key, v)
	}
}

func extractStringClaim(claims *jwt.MapClaims, key string) (string, error) {
	val, found := (*claims)[key]
	if !found {
		return "", errors.New("missing key in claims: " + key)
	}

	switch v := val.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(v), nil
	default:
		return "", fmt.Errorf("unsupported claim type for %s: %T", key, v)
	}
}

func ExtractUserIdFromClaims(claims *jwt.MapClaims) (uint, error) {
	return extractUintClaim(claims, "user_id")
}

func ExtractCompanyIdFromClaims(claims *jwt.MapClaims) (uint, error) {
	return extractUintClaim(claims, "company_id")
}

func ExtractRoleIdFromClaims(claims *jwt.MapClaims) (uint, error) {
	return extractUintClaim(claims, "role_id")
}

func ExtractQuotaTypeFromClaims(claims *jwt.MapClaims) (uint, error) {
	return extractUintClaim(claims, "quota_type")
}

func ExtractApiKeyFromClaims(claims *jwt.MapClaims) (string, error) {
	return extractStringClaim(claims, "api_key")
}
