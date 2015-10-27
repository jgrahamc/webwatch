# webwatch

Small program to download a web page, see if a string appears in it
and send email if it does.

# Usage

Grab https://www.cloudflare.com/ and see if it contains "Breaking
News" and send email if it does.

```webwatch -url=https://www.cloudflare.com/ -warn="Breaking News" -from=alice@example.com -to=bert@example.com```

`-url`: URL to get

`-warn`: string to search in the page

`-from`: From address for the email

`-to`: List of comma-separated addresses to send mail to

`-smtp`: The SMTP server to use