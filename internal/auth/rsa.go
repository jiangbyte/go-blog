package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type RSACipher struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
}

// NewRSACipher 从 path 加载 PKCS#1 私钥；文件不存在则生成并落盘。path 为空则仅内存生成。
func NewRSACipher(path string) (*RSACipher, error) {
	path = strings.TrimSpace(path)
	if path != "" {
		if key, err := loadRSAPrivateKey(path); err == nil {
			return &RSACipher{privateKey: key}, nil
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	if path != "" {
		if err := saveRSAPrivateKey(path, key); err != nil {
			return nil, fmt.Errorf("persist rsa key: %w", err)
		}
	}
	return &RSACipher{privateKey: key}, nil
}

func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("invalid rsa pem")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func saveRSAPrivateKey(path string, key *rsa.PrivateKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return os.WriteFile(path, pemBytes, 0o600)
}

func (c *RSACipher) PublicKeyPEM() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pub := x509.MarshalPKCS1PublicKey(&c.privateKey.PublicKey)
	// Prefer PKIX SubjectPublicKeyInfo for browser forge compatibility
	der, err := x509.MarshalPKIXPublicKey(&c.privateKey.PublicKey)
	if err != nil {
		b := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub})
		return string(b)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func (c *RSACipher) DecryptBase64(cipherB64 string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	raw, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("invalid password cipher")
	}
	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.privateKey, raw, nil)
	if err != nil {
		return "", fmt.Errorf("password decrypt failed")
	}
	return string(plain), nil
}
