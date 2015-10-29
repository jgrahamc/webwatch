// webwatch watches a URL and if a string appears in it, it sends an
// email
//
// Copyright (c) 2015 John Graham-Cumming

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// Construct the message to send from the parameters. The message headers are
// written one-by-one followed by an empty line and the message body.
func buildMessage(url, warn, from, to string) string {
	var m string
	for k, v := range map[string]string{
		"Date":         time.Now().Format(time.RFC822Z),
		"From":         from,
		"To":           to,
		"Subject":      fmt.Sprintf("WARNING! %s found in %s", warn, url),
		"Content-Type": "text/plain; charset=\"utf-8\"",
	} {
		m += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	m += "\r\n"
	m += fmt.Sprintf("Found %s in %s", warn, url)
	return m
}

// Send the message to the specified recipients.
func sendMessage(smtpServer, username, password, from, to, m string) error {
	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, smtpServer)
	}
	return smtp.SendMail(
		smtpServer,
		auth,
		from,
		strings.Split(to, ","),
		[]byte(m),
	)
}

// Retrieve the contents of the specified page.
func fetchPage(url string) ([]byte, error) {
	p, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer p.Body.Close()
	b, err := ioutil.ReadAll(p.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func main() {
	var (
		url        = flag.String("url", "", "`URL` to check")
		warn       = flag.String("warn", "", "Send email if this `string` is found in the web page")
		from       = flag.String("from", "", "Email `address` to send from")
		to         = flag.String("to", "", "Comma-separated list of email `addresses` to send to")
		smtpServer = flag.String("smtp", "gmail-smtp-in.l.google.com:25", "Address of SMTP `server` to use (host:port)")
		username   = flag.String("username", "", "`Username` for SMTP server")
		password   = flag.String("password", "", "`Password` for SMTP server")
	)
	flag.Parse()

	// Ensure the minimum arguments were provided.
	if *url == "" || *warn == "" || *from == "" || *to == "" {
		log.Fatalln("-url, -warn, -from, and -to must be provided")
	}

	// Fetch the contents of the page
	b, err := fetchPage(*url)
	if err != nil {
		log.Fatalf("Failed to fetch %s: %s\n", *url, err)
	}

	// Check for the search term
	if strings.Contains(string(b), *warn) {

		// Build the message
		m := buildMessage(*url, *warn, *from, *to)

		// Send the message
		err = sendMessage(*smtpServer, *username, *password, *from, *to, m)
		if err != nil {
			log.Fatalf("Failed to send message: %s\n", err)
		}
	}
}
