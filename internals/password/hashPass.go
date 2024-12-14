package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password interface{}) (string, error) {
	// Type assertion to ensure `password` is of type `string`
	strPassword, ok := password.(string)
	if !ok {
		return "", fmt.Errorf("invalid type for password: expected string")
	}

	// Generate a hashed password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(strPassword), 14)
	if err != nil {
		return "", err
	}

	// Return the hashed password as a string
	return string(hashedPassword), nil
}

func CheckPassword(password string, hash interface{}) (bool, error) {
	// Type assertion to ensure `hash` is of type `string`
	strHash, ok := hash.(string)
	if !ok {
		return false, fmt.Errorf("invalid type for hash: expected string")
	}

	// Compare the hash and password
	err := bcrypt.CompareHashAndPassword([]byte(strHash), []byte(password))
	return err == nil, err
}
