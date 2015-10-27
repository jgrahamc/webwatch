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
	client                          *http.Client
)

func main() {
	err := parseAndValidateConfiguration()
	if err != nil {
		log.Fatalln(err)
	}

	client = &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	body, err := fetchAndReturnPage()

	if strings.Contains(body, *warn) {
		err = sendReportWithMessage("%q FOUND in %s", *warn, *url)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Printf("%q NOT found in %s\n", *warn, *url)
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

	if *warn == "" {
		return errors.New("The -warn parameter is required")
	}
	if *to == "" {
		return errors.New("The -to parameter is required")
	}
	if *from == "" {
		return errors.New("The -from parameter is required")
	}
	if *url == "" {
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
	request, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		log.Println(err)
		return "", errors.New(fmt.Sprintf("Failed to get the URL %s: %s", *url, err))
	}
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 7.0; WOW64) AppleWebKit/537.35 (KHTML, like Gecko) Chrome/31.0.2049.16 Safari/537.35")

	page, err := client.Do(request)
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
	fullEmailContent := createAndReturnMessageHeader() + createAndReturnMessage(format, values...)
	log.Println(fullEmailContent)
	err := sendReportEmailThroughSMTP(fullEmailContent)
	if err != nil {
		return err
	}
	return nil
}

func createAndReturnMessageHeader() string {
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
