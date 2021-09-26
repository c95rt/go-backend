package helpers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func RemoveAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		return s
	}
	return output
}

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

func AddFileToS3(ctx *config.AppContext, file *bytes.Buffer, fileName string) (string, error) {
	_, err := s3.New(ctx.AwsS3).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(ctx.Config.AwsS3.S3Bucket),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(file.Bytes()),
		ContentLength:        aws.Int64(int64(file.Len())),
		ContentType:          aws.String(http.DetectContentType(file.Bytes())),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", ctx.Config.AwsS3.S3Url, ctx.Config.AwsS3.S3Bucket, fileName), nil
}
