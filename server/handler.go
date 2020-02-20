package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yukimochi/httpsig"
)

type handler struct {
	keyCache *keyCache
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)
	if r.Method == "GET" {
		if r.URL.Path == "/" {
			fmt.Fprint(w, "OK")
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var algo = httpsig.RSA_SHA256
	if xalgo := r.Header.Get("X-ALGO"); len(xalgo) > 0 {
		algo = httpsig.Algorithm(xalgo)
	}
	if !strings.HasPrefix(string(algo), "rsa") {
		log.Println("unsupported algorithm:", algo)
		http.Error(w, fmt.Sprintf("http signature algorithm not supported: %s", algo), http.StatusInternalServerError)
		return
	}

	verifier, err := httpsig.NewVerifier(r)
	if err != nil {
		log.Println("unable to create verifier:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pubKeyId := verifier.KeyId()

	knownKey := h.keyCache.contains(pubKeyId)
	if !knownKey {
		err = h.keyCache.fetchRemote(pubKeyId)
		if err != nil {
			log.Println("unable to fetch remote key:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	publicKey, err := h.keyCache.get(pubKeyId)
	if err != nil {
		log.Println("no key for key id", pubKeyId, ":", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = verifier.Verify(publicKey, algo); err != nil {
		log.Println("unable to verify signature:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("signed request verified with", pubKeyId)

	w.WriteHeader(http.StatusNoContent)
}
