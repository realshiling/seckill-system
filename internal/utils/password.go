package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	return string(bytes), err
}

func ComparePassword(hashed, input string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input)) == nil
}
