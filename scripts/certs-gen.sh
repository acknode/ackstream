#!/bin/bash

set -e

CERTS_DIR=${CERTS_DIR:-./certs}
CERTS_SUBJ=${CERTS_SUBJ:-"/C=VN/ST=CanTho/L=CanTho/O=AckNode/OU=Computer/CN=*.acknode.com/emailAddress=help@acknode.com"}

# 1. Generate CA's private key and self-signed certificate
# Example:
# openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=FR/ST=Occitanie/L=Toulouse/O=Tech School/OU=Education/CN=*.techschool.guru/emailAddress=techschool.guru@gmail.com"
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout "$CERTS_DIR/ca-key.pem" -out "$CERTS_DIR/ca-cert.pem" -subj "$CERTS_SUBJ"

printf "\n========= CA's self-signed certificate ========="
openssl x509 -in ./certs/ca-cert.pem -noout -text
printf "========= CA's self-signed certificate =========\n"

# 2. Generate web server's private key and certificate signing request (CSR)
# Example:
# openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=PC Book/OU=Computer/CN=*.pcbook.com/emailAddress=pcbook@gmail.com"
openssl req -newkey rsa:4096 -nodes -keyout "$CERTS_DIR/server-key.pem" -out "$CERTS_DIR/server-req.pem" -subj "$CERTS_SUBJ"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
# Example:
# openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf
openssl x509 -req -in "$CERTS_DIR/server-req.pem" -days 60 -CA "$CERTS_DIR/ca-cert.pem" -CAkey "$CERTS_DIR/ca-key.pem" -CAcreateserial -out "$CERTS_DIR/server-cert.pem" -extfile "$CERTS_DIR/server-ext.cnf"

printf "\n========= Server's signed certificate ========="
openssl x509 -in "$CERTS_DIR/server-cert.pem" -noout -text
printf "========= Server's signed certificate =========\n"