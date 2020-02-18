package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/yukimochi/httpsig"
)

func main() {
	var endpoint string
	var keyID string
	var keyPath string
	var payloadPath string
	var sendAlgo bool
	flag.StringVar(&endpoint, "endpoint", "http://localhost:7000/test", "The endpoint to publish the signed payload to.")
	flag.StringVar(&keyID, "key-id", "httpsig", "The key id of the signature.")
	flag.StringVar(&keyPath, "key", "./httpsig.pem", "The path to the key to sign the request with.")
	flag.StringVar(&payloadPath, "payload", "./body.json", "The path to the payload of the request.")
	flag.BoolVar(&sendAlgo, "algo", false, "Send the algorithm hint.")

	flag.Parse()

	keyPem, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	block, _ := pem.Decode(keyPem)
	if block == nil {
		log.Fatal("invalid RSA PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal(err.Error())
	}

	payload, err := ioutil.ReadFile(payloadPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	sigConfig := []httpsig.Algorithm{httpsig.RSA_SHA256}
	headersToSign := []string{httpsig.RequestTarget, "date"}
	signer, algo, err := httpsig.NewSigner(sigConfig, headersToSign, httpsig.Signature)
	if err != nil {
		log.Fatal(err.Error())
	}

	dateHeader := time.Now().UTC().Format(http.TimeFormat)

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader(payload))
	if err != nil {
		log.Fatal(err.Error())
	}
	request.Header.Add("content-type", "application/json")
	request.Header.Add("date", dateHeader)
	if sendAlgo {
		request.Header.Add("X-ALG", string(algo))
	}

	if err = signer.SignRequest(key, keyID, request); err != nil {
		log.Fatal(err.Error())
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Status code:", resp.StatusCode)
	if len(data) > 0 {
		fmt.Println(string(data))
	}
}
