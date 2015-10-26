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
	url, warn, from, to, smtpServer *string
	recipients                      []string
)

func main() {
	err := parseAndValidateConfiguration()
	if err != nil {
		log.Fatalln(err)
	}

	body, err := fetchAndReturnPage()

	if strings.Contains(body, *warn) {
		err = sendReportWithMessage("%q FOUND in %s", *warn, *url)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Printf("%q NOT found in %s", *warn, *url)
	}
}

func parseAndValidateConfiguration() error {
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
	refactorRecipients()

	return nil
}

func checkIfSMTPURLIsValid() error {
	_, _, err := net.SplitHostPort(*smtpServer)
	if err != nil {
		return errors.New(fmt.Sprintf("The -smtp parameter must have format host:port: %s", err))
	}
	return nil
}

func parseRecipients() {
	recipients = strings.Split(*to, ",")
}

func refactorRecipients() {
	toHeaderValue := strings.Join(recipients, ", ")
	*to = toHeaderValue
}

func fetchAndReturnPage() (string, error) {
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

// sendReportWithMessage sends any report of whois differences via email
func sendReportWithMessage(format string, values ...interface{}) error {
	fullEmailContent := createAndReturnHeader() + createAndReturnMessage(format, values...)
	log.Println(fullEmailContent)
	err := sendReportEmailThroughSMTP(fullEmailContent)
	if err != nil {
		return err
	}
	return nil
}

func createAndReturnHeader() string {
	emailHeaderFormat := `From: %s
To: %s
Date: %s
Subject: WARNING! String %q found in URL %s

`
	header := fmt.Sprintf(emailHeaderFormat, *from, *to, time.Now().Format(time.RFC822Z), *warn, *url)
	return header
}

func createAndReturnMessage(format string, values ...interface{}) string {
	message := fmt.Sprintf(format+"\n", values...)
	return message
}

func sendReportEmailThroughSMTP(fullEmailContent string) error {
	err := smtp.SendMail(*smtpServer, nil, *from, recipients, []byte(fullEmailContent))
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending message from %s to %s via %s: %s",
			*from, *to, *smtpServer, err))
	}
	return nil
}
