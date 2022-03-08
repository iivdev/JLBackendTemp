package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type UnregisteredUser struct {
	Name         string
	Password     string
	Email        string
	Confirmation string
}

type AuthResponse struct {
	Token   string
	Expires int64
	Refresh string
}

type ErrorMessage struct {
	Message string
}

var currentlyAwaiting []UnregisteredUser

var cryptSecret string

const tokenExpirationDuration = time.Minute * 5

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var serverIP string

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return stringWithCharset(length, charset)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	email := r.Form.Get("email")
	if username == "" || password == "" || email == "" {
		http.Error(w, "Bad request", 400)
		return
	}
	if !isMailCorrect(email) {
		http.Error(w, "Bad request", 400)
		return
	}
	if checkUser(username) {
		http.Error(w, "Username already in use", 409)
		return
	}
	if checkEmail(email) {
		http.Error(w, "Email already in use", 409)
		return
	}
	confirmation := randomString(16)
	currentlyAwaiting = append(currentlyAwaiting, UnregisteredUser{
		Name:         username,
		Password:     fmt.Sprintf("%x", sha256.Sum256([]byte(password))),
		Email:        email,
		Confirmation: confirmation,
	})
	err = sendMail(email, "http://"+serverIP+"/auth/confirm?salt="+confirmation, "JL Registration", details)
	if err != nil {
		http.Error(w, "Bad request", 400)
		log.Panic(err)
		return
	}
	_, err = w.Write([]byte("ok"))
	if err != nil {
		log.Panic(err)
	}
	return
}

func confirmHandler(w http.ResponseWriter, r *http.Request) {
	parameters := r.URL.Query()["salt"]
	if len(parameters) == 0 {
		http.Error(w, "Bad request", 400)
		return
	}
	salt := parameters[0]
	for i, notUser := range currentlyAwaiting {
		if notUser.Confirmation == salt {
			err := putUser(User{
				Name:     notUser.Name,
				Password: notUser.Password,
				Email:    notUser.Email,
				Rating:   0,
			})
			if err != nil {
				http.Error(w, "Bad request", 400)
				return
			}
			_, err = w.Write([]byte("Аккаунт успешно создан"))
			if err != nil {
				log.Panic(err)
			}
			currentlyAwaiting = append(currentlyAwaiting[:i], currentlyAwaiting[i+1:]...)
		}
	}
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("ok"))
	if err != nil {
		log.Panic(err)
	}
}

func checkAuthHandler(w http.ResponseWriter, r *http.Request) {
	err := authenticate(r)
	if err != nil {
		if err.Error() == "expired" {
			http.Error(w, "Token expired", 401)
		} else {
			http.Error(w, "Invalid token", 400)
		}
		return
	}
	_, err = w.Write([]byte("ok"))
	if err != nil {
		log.Panic(err)
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	u := r.Form.Get("username")
	p := r.Form.Get("password")

	if checkUser(u) {
		user, err := getUser(u)
		if err != nil {
			http.Error(w, "User not found", 404)
			return
		}
		if user.Password == fmt.Sprintf("%x", sha256.Sum256([]byte(p))) {
			err = tokenForUser(user, w)
			if err != nil {
				return
			}
		} else {
			http.Error(w, "Bad request", 400)
		}
	} else {
		http.Error(w, "Bad request", 400)
	}
}

func refreshAuthHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	tokenString := r.Form.Get("token")
	refresh := r.Form.Get("refresh")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cryptSecret), nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		http.Error(w, "Internal error", 500)
		return
	}
	user, err := getUser(fmt.Sprint(claims["user"]))
	if err != nil {
		http.Error(w, "Internal error", 500)
		return
	}
	if user.RefreshToken == refresh {
		err = tokenForUser(user, w)
		if err != nil {
			http.Error(w, "Bad request", 400)
			return
		}
	} else {
		http.Error(w, "Invalid refresh", 403)
		return
	}
}

func tokenForUser(user *User, w http.ResponseWriter) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.Name,
		"nbf":  time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cryptSecret))
	if err != nil {
		http.Error(w, "Internal error", 500)
		return err
	}

	user.RefreshToken = randomString(32)
	user.TokenExpires = time.Now().Add(tokenExpirationDuration).Unix()

	err = putUser(*user)
	if err != nil {
		http.Error(w, "Internal error", 500)
		return fmt.Errorf("bad data")
	}

	res, err := json.Marshal(AuthResponse{
		Token:   tokenString,
		Expires: user.TokenExpires,
		Refresh: user.RefreshToken,
	})
	if err != nil {
		http.Error(w, "Internal error", 500)
		log.Panic(err)
	}
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, "Internal error", 500)
		log.Panic(err)
	}
	return nil
}

func authenticate(r *http.Request) error {
	authHeader := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if authHeader == "" {
		return fmt.Errorf("empty")
	}
	token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cryptSecret), nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("bad data")
	}
	user, err := getUser(fmt.Sprint(claims["user"]))
	if err != nil {
		return fmt.Errorf("bad data")
	}
	if user.TokenExpires < time.Now().Unix() {
		return fmt.Errorf("expired")
	}
	return nil
}
