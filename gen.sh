#!/bin/sh

openssl genrsa -out server.key 2048
openssl req -new -x509 -days 365 -key server.key -out server.crt -subj "/CN=proxy CA"
openssl x509 -noout -text -in server.crt