package ssl

import (
	"bytes"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestParsePrivateKey(t *testing.T) {
	err := os.WriteFile("/tmp/key.pem", []byte(`-----BEGIN PRIVATE KEY-----
MIGkAgEBBDADD6tfCoDUIbeNmLtdnJOOLybRqMwjYn4A9LXYipEylNK2mLIJmMdn
+PPjG/kubxegBwYFK4EEACKhZANiAARDTxBBIgWlblL6zEObX/j+EbhQms93dJji
b8w3Q8p2uxxmdyvNLpyaInQkSvE6V7TptBJFED8KAhXqo8Mq1Li2EkQUSovMiyv7
dIm8OZfjEa610lCZyPcBBKVLZHfwIGw=
-----END PRIVATE KEY-----
`), 0644)
	assert.NoError(t, err)
	_, err = ParsePrivateKey("/tmp/key.pem")
	assert.NoError(t, err)
	os.Remove("/tmp/key.pem")
}

//Func TestTerminateMiddlewareWithValidBody(t *testing.T) {
//	log := zerolog.Nop()
//	privkey, _ := ecies.GenerateKey(rand.Reader, elliptic.P384(), nil)
//
//	handler := Terminate(log, privkey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
//
//	plaintext := []byte("Hello, World!")
//	cyphertext, _ := ecies.Encrypt(rand.Reader, &privkey.PublicKey, plaintext, nil, nil)
//
//	req := httptest.NewRequest("POST", "/", bytes.NewReader(cyphertext))
//	w := httptest.NewRecorder()
//
//	handler.ServeHTTP(w, req)
//
//	assert.Equal(t, http.StatusOK, w.Code)
//	assert.Equal(t, plaintext, w.Body.Bytes())
//}.

func TestTerminateMiddlewareWithInvalidBody(t *testing.T) {
	log := zerolog.Nop()
	privkey, _ := ecies.GenerateKey(rand.Reader, ecies.DefaultCurve, nil)

	handler := Terminate(log, privkey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("invalid body")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTerminateMiddlewareWithEmptyBody(t *testing.T) {
	log := zerolog.Nop()
	privkey, _ := ecies.GenerateKey(rand.Reader, ecies.DefaultCurve, nil)

	handler := Terminate(log, privkey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("POST", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
