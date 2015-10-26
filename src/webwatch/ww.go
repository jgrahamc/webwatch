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
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// report adds a message (printf style) to message to be emailed
func report(msg *string, format string, values ...interface{}) {
	add := fmt.Sprintf(format+"\n", values...)
	log.Printf(add)
	*msg += add
}

// sendReport sends any report of whois differences via email
func sendReport(server, from, warn, url, msg string, to []string) {
	if msg == "" {
		return
	}

	t := strings.Join(to, ", ")
	header := fmt.Sprintf(`From: %s
To: %s
Date: %s
Subject: WARNING! String %s found in URL %s

`, from, t, time.Now().Format(time.RFC822Z), warn, url)

	msg = header + msg
	err := smtp.SendMail(server, nil, from, to, []byte(msg))
	if err != nil {
		log.Printf("Error sending message from %s to %s via %s: %s",
			from, t, server, err)
	}
}

func main() {
	url := flag.String("url", "", "URL to check")
	warn := flag.String("warn", "",
		"Send email if this string is found in the web page")
	from := flag.String("from", "", "Email addresses to send from")
	to := flag.String("to", "",
		"Comma-separated list of email addresses to send to")
	smtpServer := flag.String("smtp", "gmail-smtp-in.l.google.com:25",
		"Address of SMTP server to use (host:port)")
	flag.Parse()

	if *warn == "" {
		log.Fatalf("The -warn parameter is required")
	}
	if *to == "" {
		log.Fatalf("The -to parameter is required")
	}
	if *from == "" {
		log.Fatalf("The -from parameter is required")
	}
	if *url == "" {
		log.Fatalf("The -url parameter is required")
	}

	_, _, err := net.SplitHostPort(*smtpServer)
	if err != nil {
		log.Fatalf("The -smtp parameter must have format host:port: %s",
			err)
	}

	recipients := strings.Split(*to, ",")

	page, err := http.Get(*url)
	if err != nil {
		log.Fatalf("Failed to get the URL %s: %s", *url, err)
	}
	defer page.Body.Close()

	body, err := ioutil.ReadAll(page.Body)
	if err != nil {
		log.Fatalf("Failed to read body of the URL %s: %s", *url, err)
	}

	if strings.Contains(string(body), *warn) {
		var msg string
		report(&msg, "Found %s in %s", *warn, *url)
		sendReport(*smtpServer, *from, *warn, *url, msg, recipients)
	}
}
