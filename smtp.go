package main

import (
	"crypto/tls"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type SmtpDetails struct{
	Mail   string
	Password string
	Host     string
}

var details SmtpDetails

func sendMail(recipient string, content string, subject string, details SmtpDetails) error{
	from := mail.Address{Address: details.Mail}
	to   := mail.Address{Address: recipient}
	host, _, _ := net.SplitHostPort(details.Host)
	auth := smtp.PlainAuth("",details.Mail, details.Password, host)
	dial, err := tls.Dial("tcp", details.Host,&tls.Config{InsecureSkipVerify: true, ServerName: host})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(dial, host)
	if err != nil {
		return err
	}
	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(from.Address); err != nil {
		return err
	}
	if err = c.Rcpt(to.Address); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	message := "From: "+from.Address+"\r\n"
	message += "To: "+to.Address+"\r\n"
	message += "Subject: "+subject+"\r\n"
	message += "\r\n"+content
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	_ = c.Quit()
	return nil
}

func isMailCorrect(address string) bool{
	return strings.Contains(address,"@")&&strings.Contains(address,".")
}