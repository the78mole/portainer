package crypto

import (
	"os"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPrivateKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lQdGBGYxmD0BEADKvW8kOvW+YT1UfH4xKf7L2GK3ZfYXb7b7b7b7b7b7b7b7b7b7
b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7
b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7
... (truncated for brevity)
-----END PGP PRIVATE KEY BLOCK-----`

const testPublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBGYxmD0BEADKvW8kOvW+YT1UfH4xKf7L2GK3ZfYXb7b7b7b7b7b7b7b7b7b7
b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7
b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7b7
... (truncated for brevity)
-----END PGP PUBLIC KEY BLOCK-----`

// generateTestKeyPair generates a test key pair for testing
func generateTestKeyPair(t *testing.T) (privateKey, publicKey string) {
	entity, err := openpgp.NewEntity("Test User", "Test for PGP decryption", "test@example.com", nil)
	require.NoError(t, err)

	// Serialize private key
	var privKeyBuf strings.Builder
	privKeyWriter, err := armor.Encode(&privKeyBuf, openpgp.PrivateKeyType, nil)
	require.NoError(t, err)
	err = entity.SerializePrivate(privKeyWriter, nil)
	require.NoError(t, err)
	err = privKeyWriter.Close()
	require.NoError(t, err)

	// Serialize public key
	var pubKeyBuf strings.Builder
	pubKeyWriter, err := armor.Encode(&pubKeyBuf, openpgp.PublicKeyType, nil)
	require.NoError(t, err)
	err = entity.Serialize(pubKeyWriter)
	require.NoError(t, err)
	err = pubKeyWriter.Close()
	require.NoError(t, err)

	return privKeyBuf.String(), pubKeyBuf.String()
}

// encryptTestData encrypts test data using the public key
func encryptTestData(t *testing.T, data, publicKey string) []byte {
	// Parse public key
	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(publicKey))
	require.NoError(t, err)

	// Encrypt data
	var encryptedBuf strings.Builder
	encryptedWriter, err := armor.Encode(&encryptedBuf, "PGP MESSAGE", nil)
	require.NoError(t, err)

	plaintextWriter, err := openpgp.Encrypt(encryptedWriter, keyring, nil, nil, nil)
	require.NoError(t, err)

	_, err = plaintextWriter.Write([]byte(data))
	require.NoError(t, err)

	err = plaintextWriter.Close()
	require.NoError(t, err)

	err = encryptedWriter.Close()
	require.NoError(t, err)

	return []byte(encryptedBuf.String())
}

func TestNewPGPDecryptor(t *testing.T) {
	privateKey := "test-private-key"
	decryptor := NewPGPDecryptor(privateKey)
	
	assert.NotNil(t, decryptor)
	assert.Equal(t, privateKey, decryptor.privateKey)
}

func TestValidatePGPPrivateKey(t *testing.T) {
	tests := []struct {
		name        string
		privateKey  string
		expectError bool
	}{
		{
			name:        "empty private key",
			privateKey:  "",
			expectError: true,
		},
		{
			name:        "invalid private key",
			privateKey:  "invalid-key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePGPPrivateKey(tt.privateKey)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPGPDecryptor_DecryptData(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	
	testData := "SECRET_VAR=secret_value\nANOTHER_SECRET=another_value\n"
	encryptedData := encryptTestData(t, testData, publicKey)

	decryptor := NewPGPDecryptor(privateKey)
	decryptedData, err := decryptor.DecryptData(encryptedData)
	
	require.NoError(t, err)
	assert.Equal(t, testData, string(decryptedData))
}

func TestPGPDecryptor_DecryptFile(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	
	testData := "SECRET_VAR=secret_value\nANOTHER_SECRET=another_value\n"
	encryptedData := encryptTestData(t, testData, publicKey)

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test-pgp-*.pgp")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(encryptedData)
	require.NoError(t, err)
	err = tempFile.Close()
	require.NoError(t, err)

	decryptor := NewPGPDecryptor(privateKey)
	decryptedData, err := decryptor.DecryptFile(tempFile.Name())
	
	require.NoError(t, err)
	assert.Equal(t, testData, string(decryptedData))
}

func TestGetPGPPrivateKeyFromEnv(t *testing.T) {
	// Test with no environment variable set
	originalValue := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
	
	key := GetPGPPrivateKeyFromEnv()
	assert.Empty(t, key)

	// Test with environment variable set
	testKey := "test-private-key"
	os.Setenv("PORTAINER_PGP_PRIVATE_KEY", testKey)
	
	key = GetPGPPrivateKeyFromEnv()
	assert.Equal(t, testKey, key)

	// Restore original value
	if originalValue != "" {
		os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalValue)
	} else {
		os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
	}
}

func TestPGPDecryptor_InvalidPrivateKey(t *testing.T) {
	decryptor := NewPGPDecryptor("invalid-private-key")
	
	encryptedData := []byte("invalid encrypted data")
	_, err := decryptor.DecryptData(encryptedData)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestPGPDecryptor_NonExistentFile(t *testing.T) {
	decryptor := NewPGPDecryptor("any-key")
	
	_, err := decryptor.DecryptFile("/non/existent/file.pgp")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read encrypted file")
}