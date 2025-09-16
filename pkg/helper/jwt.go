package helper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(
	secret string,
	minutesToExpired int,
	userId, companyId, roleId uint,
	apiKey string,
) (string, error) {
	willExpiredAt := time.Now().Add(time.Duration(minutesToExpired) * time.Minute)

	claims := jwt.MapClaims{}
	claims["user_id"] = userId
	claims["company_id"] = companyId
	claims["role_id"] = roleId
	claims["api_key"] = apiKey
	claims["exp"] = willExpiredAt.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func ExtractClaimsFromJWT(token, secret string) (*jwt.MapClaims, error) {
	claims := &jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(requestToken *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func extractUintClaim(claims *jwt.MapClaims, key string) (uint, error) {
	val, found := (*claims)[key]
	if !found {
		return 0, errors.New("missing key in claims: " + key)
	}

	strVal := fmt.Sprintf("%v", val)
	parsedVal, err := strconv.ParseUint(strVal, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s format: %v", key, err)
	}

	return uint(parsedVal), nil
}

func extractStringClaim(claims *jwt.MapClaims, key string) (string, error) {
	val, found := (*claims)[key]
	if !found {
		return "", errors.New("missing key in claims: " + key)
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("claim %s is not a string", key)
	}

	return strVal, nil
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

func ExtractApiKeyFromClaims(claims *jwt.MapClaims) (string, error) {
	return extractStringClaim(claims, "api_key")
}
