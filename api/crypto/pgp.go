package crypto

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

// PGPDecryptor handles PGP decryption operations
type PGPDecryptor struct {
	privateKey string
	passphrase string
}

// NewPGPDecryptor creates a new PGP decryptor with the provided private key
func NewPGPDecryptor(privateKey string) *PGPDecryptor {
	return &PGPDecryptor{
		privateKey: privateKey,
		passphrase: "",
	}
}

// NewPGPDecryptorWithPassphrase creates a new PGP decryptor with the provided private key and passphrase
func NewPGPDecryptorWithPassphrase(privateKey, passphrase string) *PGPDecryptor {
	return &PGPDecryptor{
		privateKey: privateKey,
		passphrase: passphrase,
	}
}

// DecryptFile decrypts a PGP-encrypted file and returns the decrypted content
func (p *PGPDecryptor) DecryptFile(encryptedFilePath string) ([]byte, error) {
	// Read the encrypted file
	encryptedData, err := os.ReadFile(encryptedFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	return p.DecryptData(encryptedData)
}

// DecryptData decrypts PGP-encrypted data and returns the decrypted content
func (p *PGPDecryptor) DecryptData(encryptedData []byte) ([]byte, error) {
	// Parse the private key
	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(p.privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Decrypt the private key if it has a passphrase
	if p.passphrase != "" {
		for _, entity := range keyring {
			if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
				err = entity.PrivateKey.Decrypt([]byte(p.passphrase))
				if err != nil {
					return nil, fmt.Errorf("failed to decrypt private key with passphrase: %w", err)
				}
			}
			for _, subkey := range entity.Subkeys {
				if subkey.PrivateKey != nil && subkey.PrivateKey.Encrypted {
					err = subkey.PrivateKey.Decrypt([]byte(p.passphrase))
					if err != nil {
						return nil, fmt.Errorf("failed to decrypt subkey with passphrase: %w", err)
					}
				}
			}
		}
	}

	// Try to decrypt the data
	var decryptedData []byte

	// First, try to decode as armored data
	armorBlock, err := armor.Decode(bytes.NewReader(encryptedData))
	if err != nil {
		// If not armored, try to decrypt as binary
		return p.decryptBinary(encryptedData, keyring)
	}

	// Decrypt the armored data
	md, err := openpgp.ReadMessage(armorBlock.Body, keyring, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read PGP message: %w", err)
	}

	decryptedData, err = io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decryptedData, nil
}

// decryptBinary attempts to decrypt binary (non-armored) PGP data
func (p *PGPDecryptor) decryptBinary(encryptedData []byte, keyring openpgp.EntityList) ([]byte, error) {
	md, err := openpgp.ReadMessage(bytes.NewReader(encryptedData), keyring, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read PGP message: %w", err)
	}

	decryptedData, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decryptedData, nil
}

// GetPGPPrivateKeyFromEnv retrieves the PGP private key from environment variable
func GetPGPPrivateKeyFromEnv() string {
	return os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
}

// ValidatePGPPrivateKey validates that the provided string is a valid PGP private key
func ValidatePGPPrivateKey(privateKey string) error {
	if privateKey == "" {
		return fmt.Errorf("private key is empty")
	}

	_, err := openpgp.ReadArmoredKeyRing(strings.NewReader(privateKey))
	if err != nil {
		return fmt.Errorf("invalid PGP private key: %w", err)
	}

	return nil
}