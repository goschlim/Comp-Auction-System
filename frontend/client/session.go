package client

import (
	"errors"
	"goLive/frontend/common"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const cookieName = "token"

// set session time here
const sessionTime = time.Minute * 30

// claims struct to be encoded to JWT
type claims struct {
	Username string
	jwt.StandardClaims
}

// setToken sets cookie with JWT JSON Web Token.
func setToken(username string) *http.Cookie {
	expireCookie := time.Now().Add(sessionTime)
	expireToken := expireCookie.Unix()

	// set new token claims
	claims := claims{
		username,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    addr,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString(GetEnvJWT())

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    signedToken,
		Expires:  expireCookie,
		HttpOnly: true,
		Path:     "/",
		Domain:   "127.0.0.1",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie
}

// generateToken generates JWT JSON Web Token.
func generateToken(getUser user, res http.ResponseWriter) (string, *claims, error) {
	// JWT JSON Web Token
	// set token expiry time
	expiryTime := time.Now().Add(sessionTime)
	// set JWT claims that contains username and expiry time
	claims := &claims{
		Username: getUser.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiryTime.Unix(), // unix milliseconds
		},
	}

	// create token with alogorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// create signed token
	signedToken, err := token.SignedString(GetEnvJWT())
	if err != nil { // error in creating token
		return "", claims, errors.New("token not created")
	}

	return signedToken, claims, nil
}

// GetEnvJWT gets jwt from .env file
func GetEnvJWT() []byte {
	return []byte(common.GetEnv("JWT_SECRET_KEY"))
}
