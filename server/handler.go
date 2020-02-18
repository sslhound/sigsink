package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/yukimochi/httpsig"
)

type handler struct {
	keyCache *keyCache
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.Path == "/" {
		fmt.Fprint(w, "OK")
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
		http.Error(w, fmt.Sprintf("http signature algorithm not supported: %s", algo), http.StatusInternalServerError)
		return
	}

	verifier, err := httpsig.NewVerifier(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pubKeyId := verifier.KeyId()

	knownKey := h.keyCache.contains(pubKeyId)
	if !knownKey {
		err = h.keyCache.fetchRemote(pubKeyId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	publicKey, err := h.keyCache.get(pubKeyId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = verifier.Verify(publicKey, algo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("signed request verified with", pubKeyId)

	w.WriteHeader(http.StatusNoContent)
}
