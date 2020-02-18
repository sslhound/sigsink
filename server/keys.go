package server

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type keyCache struct {
	keys              map[string]*rsa.PublicKey
	lock              sync.Mutex
	enableFetchRemote bool
	logger            *zap.Logger
}

func (c *keyCache) addDirectory(directory string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".pem") {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			block, _ := pem.Decode(data)
			if block == nil {
				return fmt.Errorf("invalid RSA PEM")
			}

			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return err
			}
			rsaPublicKey, ok := pub.(*rsa.PublicKey)
			if !ok {
				return fmt.Errorf("invalid RSA PEM")
			}

			keyID, ok := block.Headers["key_id"]
			if ok {
				c.keys[keyID] = rsaPublicKey
				c.logger.Info("added key", zap.String("key_id", keyID), zap.String("source", directory), zap.String("path", path))
			}

		}
		return nil
	})
}

func (c *keyCache) fetchRemote(location string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.enableFetchRemote {
		return fmt.Errorf("fetching remote keys is disabled")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("invalid RSA PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	rsaPublicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("invalid RSA PEM")
	}

	c.keys[location] = rsaPublicKey
	c.logger.Info("added key", zap.String("key_id", location))

	return nil
}

func (c *keyCache) contains(keyID string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok := c.keys[keyID]
	return ok
}

func (c *keyCache) get(keyID string) (*rsa.PublicKey, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key, ok := c.keys[keyID]
	if !ok {
		return nil, fmt.Errorf("unknown key")
	}
	return key, nil
}
