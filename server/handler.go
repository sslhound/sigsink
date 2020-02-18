package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yukimochi/httpsig"
)

type handler struct {
	keyCache *keyCache
}

func (h handler) verifySignatureHandler(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}

	var algo = httpsig.RSA_SHA256
	if xalgo := c.GetHeader("X-ALGO"); len(xalgo) > 0 {
		algo = httpsig.Algorithm(xalgo)
	}
	if !strings.HasPrefix(string(algo), "rsa") {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("http signature algorithm not supported: %s", algo),
		})
		return
	}

	verifier, err := httpsig.NewVerifier(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	pubKeyId := verifier.KeyId()

	knownKey := h.keyCache.contains(pubKeyId)
	if !knownKey {
		err = h.keyCache.fetchRemote(pubKeyId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}
	}
	publicKey, err := h.keyCache.get(pubKeyId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err = verifier.Verify(publicKey, algo); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
