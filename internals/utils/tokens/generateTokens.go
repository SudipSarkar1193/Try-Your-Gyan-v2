package tokens

import (
	"os"
	"time"

	"github.com/SudipSarkar1193/Try-Your-Gyan-v2.git/internals/types"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateVerifyToken(user *types.User) (string, error) {

	tokenClaims := jwt.MapClaims{
		"sub":  user.Id,
		"name": user.Username,
		"exp":  time.Now().Add(20 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateTokens(user *types.User) (string, string, error) {

	accessTokenClaims := jwt.MapClaims{
		"sub":  user.Id,
		"name": user.Username,
		"exp":  time.Now().Add(18 * time.Minute).Unix(),
	}
	refreshTokenClaims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	accessTokenStr, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return "", "", err
	}
	refreshTokenStr, err := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return "", "", err
	}

	return accessTokenStr, refreshTokenStr, nil
}
