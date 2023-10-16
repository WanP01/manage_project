package jwts

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JwtToken struct {
	AccessToken  string
	RefreshToken string
	AccessExp    int64
	RefreshExp   int64
}

func CreateToken(val string, exp time.Duration, secret string, refreshExp time.Duration, refreshSecret string) (*JwtToken, error) {
	aExp := time.Now().Add(exp).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   aExp,
	})
	aToken, err := token.SignedString([]byte(secret))

	rExp := time.Now().Add(refreshExp).Unix()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token": val,
		"exp":   rExp,
	})
	rToken, err := refreshToken.SignedString([]byte(refreshSecret))

	return &JwtToken{
		AccessToken:  aToken,
		AccessExp:    aExp,
		RefreshExp:   rExp,
		RefreshToken: rToken,
	}, err

}
