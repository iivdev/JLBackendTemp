package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

type UnregisteredUser struct{
	Name string
	Password string
	Email string
	Confirmation string
}

var currentlyAwaiting []UnregisteredUser

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

func registerHandler(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err!=nil{
		_, err = w.Write([]byte("error"))
		if err!=nil{
			log.Panic(err)
		}
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	email := r.Form.Get("email")
	if username==""||password==""||email==""{
		_, err = w.Write([]byte("error"))
		if err!=nil{
			log.Panic(err)
		}
		return
	}
	if !isMailCorrect(email){
		_, err = w.Write([]byte("error"))
		if err!=nil{
			log.Panic(err)
		}
		return
	}
	if checkUser(username){
		_, err = w.Write([]byte("exists"))
		if err!=nil{
			log.Panic(err)
		}
		return
	}
	if checkEmail(email){
		_, err = w.Write([]byte("exists"))
		if err!=nil{
			log.Panic(err)
		}
		return
	}
	confirmation := randomString(16)
	currentlyAwaiting = append(currentlyAwaiting, UnregisteredUser{
		Name:     username,
		Password: password,
		Email:    email,
		Confirmation: confirmation,
	})
	err = sendMail(email,"http://"+serverIP+"/auth/confirm?salt="+confirmation,"JL Registration",details)
	if err!=nil{
		_, err = w.Write([]byte("error"))
		if err!=nil{
			log.Panic(err)
		}
		log.Panic(err)
		return
	}
	_, err = w.Write([]byte("ok"))
	if err!=nil{
		log.Panic(err)
	}
	return
}

func confirmHandler (w http.ResponseWriter, r *http.Request){
	parameters := r.URL.Query()["salt"]
	if len(parameters)==0{
		_, err := w.Write([]byte("error"))
		if err!=nil{
			log.Panic(err)
		}
		return
	}
	salt := parameters[0]
	for i, notUser := range currentlyAwaiting{
		if notUser.Confirmation==salt{
			err := addUser(User{
				Name:     notUser.Name,
				Password: notUser.Password,
				Email:    notUser.Email,
				Rating:   0,
			})
			if err!=nil{
				_, err = w.Write([]byte("error"))
				if err!=nil{
					log.Panic(err)
				}
				return
			}
			_, err = w.Write([]byte("Аккаунт успешно создан"))
			if err!=nil{
				log.Panic(err)
			}
			currentlyAwaiting = append(currentlyAwaiting[:i], currentlyAwaiting[i+1:]...)
		}
	}
}

func pingHandler (w http.ResponseWriter, _ *http.Request){
	_, err := w.Write([]byte("ok"))
	if err!=nil{
		log.Panic(err)
	}
}

func checkAuthHandler (w http.ResponseWriter, r *http.Request){
	u := r.Form.Get("username")
	p := r.Form.Get("password")
	ok := true
	if ok{
		if checkUser(u){
			user, err := getUser(u)
			if err!=nil{
				log.Panic(err)
			}
			if user.Password==p{
				_, err := w.Write([]byte("ok"))
				if err!=nil{
					log.Panic(err)
				}
			}else{
				_, err := w.Write([]byte("error"))
				if err!=nil{
					log.Panic(err)
				}
			}
		}else{
			_, err := w.Write([]byte("error"))
			if err!=nil{
				log.Panic(err)
			}
		}
	}else{
		_, err := w.Write([]byte("error"))
		if err!=nil{
			log.Panic(err)
		}
	}
}