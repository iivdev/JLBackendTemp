package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	ServerIp string
	MailDetails SmtpDetails
}

func ReadConfig() error{
	data, err := ioutil.ReadFile("config.json")
	if err!=nil{
		CreateBlankConfig()
		return fmt.Errorf("no config found, blank created")
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err!=nil{
		CreateBlankConfig()
		return fmt.Errorf("incorrect config")
	}
	details = config.MailDetails
	serverIP = config.ServerIp
	return nil
}

func CreateBlankConfig(){
	var config Config
	data, err := json.Marshal(config)
	if err!=nil{
		return
	}
	err = ioutil.WriteFile("config.json", data, 0644)
	if err!=nil{
		return
	}
}
