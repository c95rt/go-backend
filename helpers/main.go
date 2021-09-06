package helpers

import (
	"time"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func ParserTokenUnverified(tokenStr string) (jwt.MapClaims, bool) {
	var p jwt.Parser
	token, _, ok := p.ParseUnverified(tokenStr, jwt.MapClaims{})
	if ok != nil {
		return nil, false
	}
	tokendata, _ := token.Claims.(jwt.MapClaims)
	return tokendata, true
}

func Contains(a []int, x int) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func AuthenticateHashedPassword(hashed string, inputPassword string) bool {
	configPasswordHashBytes := []byte(hashed)
	inputPasswordBytes := []byte(inputPassword)
	err := bcrypt.CompareHashAndPassword(configPasswordHashBytes, inputPasswordBytes)
	if err != nil {
		return false
	}
	return true
}

func GenerateToken(user *models.User, jwtSecret string) (string, error) {
	var r []int
	for _, role := range user.Roles {
		r = append(r, role.ID)
	}
	claims := struct {
		User map[string]interface{} `json:"u"`
		jwt.StandardClaims
	}{
		map[string]interface{}{
			"r":         r,
			"i":         user.ID,
			"email":     user.Email,
			"lastName":  user.Lastname,
			"firstName": user.Firstname,
		},
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return token, nil
}
