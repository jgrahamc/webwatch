// webwatch watches a URL and if a string appears in it, it sends an
// email
//
// Copyright (c) 2015 John Graham-Cumming

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

var (
	message, fullEmailContent       string
	url, warn, from, to, smtpServer *string
	recipients                      []string
)

// sendReport sends any report of whois differences via email

//server, from, warn, url, fullEmailContent string, to []string
func sendReport() error {
	if fullEmailContent == "" {
		return errors.New("")
	}

	toHeaderValue := strings.Join(recipients, ", ")

	emailHeaderFormat := `From: %s
To: %s
Date: %s
Subject: WARNING! String %q found in URL %s

`
	header := fmt.Sprintf(emailHeaderFormat, *from, toHeaderValue, time.Now().Format(time.RFC822Z), *warn, *url)

	fullEmailContent = header + fullEmailContent

	//<DEBUG>
	fmt.Println(fullEmailContent)
	//</DEBUG>

	err := smtp.SendMail(*smtpServer, nil, *from, recipients, []byte(fullEmailContent))
	if err != nil {
		log.Printf("Error sending message from %s to %s via %s: %s",
			*from, toHeaderValue, *smtpServer, err)
	}

	return nil
}

// addMessageToReport adds a message (printf style) to message to be emailed
func addMessageToReport(format string, values ...interface{}) {
	add := fmt.Sprintf(format+"\n", values...)
	log.Printf(add)
	fullEmailContent += add
}

func main() {
	err := parseConfiguration()
	if err != nil {
		log.Fatalln(err)
	}

	body, err := fetchPage()

	if strings.Contains(body, *warn) {
		addMessageToReport("%q FOUND in %s", *warn, *url)

		//*smtpServer, *from, *warn, *url, fullEmailContent, recipients
		err = sendReport()
		if err != nil {
			log.Fatalln(err)
		}

	} else {
		fmt.Printf("%q NOT found in %s", *warn, *url)
	}
}

func fetchPage() (string, error) {
	page, err := http.Get(*url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get the URL %s: %s", *url, err))
	}
	defer page.Body.Close()

	body, err := ioutil.ReadAll(page.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to read body of the URL %s: %s", *url, err))
	}

	return string(body), nil
}

func parseConfiguration() error {
	url = flag.String("url", "",
		"URL to check")

	warn = flag.String("warn", "",
		"Send email if this string is found in the web page")

	from = flag.String("from", "",
		"Email addresses to send from")

	to = flag.String("to", "",
		"Comma-separated list of email addresses to send to")

	smtpServer = flag.String("smtp", "gmail-smtp-in.l.google.com:25",
		"Address of SMTP server to use (host:port)")

	flag.Parse()

	if len(*warn) < 1 {
		return errors.New("The -warn parameter is required")
	}
	if len(*to) < 1 {
		return errors.New("The -to parameter is required")
	}
	if len(*from) < 1 {
		return errors.New("The -from parameter is required")
	}
	if len(*url) < 1 {
		return errors.New("The -url parameter is required")
	}

	err := checkIfSMTPURLIsValid()
	if err != nil {
		return err
	}

	parseRecipients()

	return nil
}

func parseRecipients() {
	recipients = strings.Split(*to, ",")
}

func checkIfSMTPURLIsValid() error {
	_, _, err := net.SplitHostPort(*smtpServer)
	if err != nil {
		return errors.New(fmt.Sprintf("The -smtp parameter must have format host:port: %s", err))
	}
	return nil
}
