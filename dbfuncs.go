package main

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

var userDB *leveldb.DB

type User struct{
	Name string
	Password string
	Email string
	Rating int
}

func initDBs(){
	var err error
	userDB, err = leveldb.OpenFile("userDB", nil)
	if err!=nil{
		log.Fatal(err)
		return
	}
}

func addUser(user User) error{
	marshaledUser, err := json.Marshal(user)
	if err!=nil{
		log.Panic(err)
		return err
	}
	err = userDB.Put([]byte(user.Name), marshaledUser, nil)
	if err!=nil{
		log.Panic(err)
		return err
	}
	return nil
}

func getUser(username string) (*User, error){
	marshaledUser, err := userDB.Get([]byte(username), nil)
	if err!=nil{
		log.Panic(err)
		return nil, err
	}
	var unMarshaledUser User
	err = json.Unmarshal(marshaledUser, &unMarshaledUser)
	if err!=nil{
		log.Panic(err)
		return nil, err
	}
	return &unMarshaledUser, nil
}

func checkUser(username string) bool{
	_, err := userDB.Get([]byte(username),nil)
	return err==nil
}

func checkEmail(email string) bool{
	iter := userDB.NewIterator(nil, nil)
	for iter.Next() {
		var u User
		err := json.Unmarshal(iter.Value(), &u)
		if err!=nil{
			log.Panic(err)
		}
		if u.Email==email {
			return true
		}
	}
	return false
}

func deleteUser(username string) error{
	return userDB.Delete([]byte(username),nil)
}


