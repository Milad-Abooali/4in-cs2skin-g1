package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"
)

// InArray - Global Helper
func InArray[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

// ToLowerArray - Global Helper
func ToLowerArray(arr []string) []string {
	lowerArr := make([]string, len(arr))
	for i, v := range arr {
		lowerArr[i] = strings.ToLower(v)
	}
	return lowerArr
}

// MD5UserID - Global Helper
func MD5UserID(userID int) string {
	data := []byte(fmt.Sprintf("%d", userID))
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// RoundToTwoDigits - Global Helper
func RoundToTwoDigits(val float64) float64 {
	return math.Round(val*100) / 100
}

func ValidateAdminKey(data map[string]interface{}) (string, error) {
	val, exists := data["adminKey"]
	if !exists {
		return "", fmt.Errorf("ADMIN_KEY_EXPECTED:2001")
	}
	keyStr, ok := val.(string)
	if !ok || keyStr == "" {
		return "", fmt.Errorf("ADMIN_KEY_EMPTY:2001")
	}
	if keyStr != os.Getenv("ADMIN_KEY") {
		return "", fmt.Errorf("ADMIN_KEY_INVALID:2001")
	}
	return keyStr, nil
}
