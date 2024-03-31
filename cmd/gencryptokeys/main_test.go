package main

//nolint:all
import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeys_Success(t *testing.T) {
	path := "/tmp/test_keys"
	defer os.RemoveAll(path)

	err := generateKeys(path)
	assert.NoError(t, err)

	_, err = os.Stat(path + "/cert.pem")
	assert.NoError(t, err)

	_, err = os.Stat(path + "/privkey.pem")
	assert.NoError(t, err)
}

func TestGenerateKeys_FailOnInvalidPath(t *testing.T) {
	path := "" // Invalid path

	err := generateKeys(path)
	assert.Error(t, err)
}

func TestGenerateKeys_FailOnExistingDirectory(t *testing.T) {
	path := "existing_dir"
	os.Mkdir(path, 0700)
	defer os.RemoveAll(path)

	err := generateKeys(path)
	assert.Error(t, err)
}

func TestGenerateKeys_CertificateContent(t *testing.T) {
	path := "test_keys"
	defer os.RemoveAll(path)

	err := generateKeys(path)
	assert.NoError(t, err)

	certPEM, err := ioutil.ReadFile(path + "/cert.pem")
	assert.NoError(t, err)
	assert.Contains(t, string(certPEM), "CERTIFICATE")
}

func TestGenerateKeys_PrivateKeyContent(t *testing.T) {
	path := "test_keys"
	defer os.RemoveAll(path)

	err := generateKeys(path)
	assert.NoError(t, err)

	privKeyPEM, err := ioutil.ReadFile(path + "/privkey.pem")
	assert.NoError(t, err)
	assert.Contains(t, string(privKeyPEM), "PRIVATE KEY")
}
