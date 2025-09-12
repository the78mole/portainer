package exec

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_testCreateEnvFile(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name         string
		stack        *portainer.Stack
		expected     string
		expectedFile bool
	}{
		{
			name: "should not add env file option if stack doesn't have env variables",
			stack: &portainer.Stack{
				ProjectPath: dir,
				Env:         nil,
			},
			expected: "",
		},
		{
			name: "should not add env file option if stack's env variables are empty",
			stack: &portainer.Stack{
				ProjectPath: dir,
				Env:         []portainer.Pair{},
			},
			expected: "",
		},
		{
			name: "should add env file option if stack has env variables",
			stack: &portainer.Stack{
				ProjectPath: dir,
				Env: []portainer.Pair{
					{Name: "var1", Value: "value1"},
					{Name: "var2", Value: "value2"},
				},
			},
			expected: "var1=value1\nvar2=value2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := testCreateEnvFile(tt.stack)

			if tt.expected != "" {
				assert.Equal(t, filepath.Join(tt.stack.ProjectPath, "stack.env"), result)

				f, _ := os.Open(path.Join(dir, "stack.env"))
				content, _ := io.ReadAll(f)

				assert.Equal(t, tt.expected, string(content))
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func Test_createEnvFile_mergesDefultAndInplaceEnvVars(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(path.Join(dir, ".env"), []byte("VAR1=VAL1\nVAR2=VAL2\n"), 0600)
	stack := &portainer.Stack{
		ProjectPath: dir,
		Env: []portainer.Pair{
			{Name: "VAR1", Value: "NEW_VAL1"},
			{Name: "VAR3", Value: "VAL3"},
		},
	}
	result, err := testCreateEnvFile(stack)
	assert.Equal(t, filepath.Join(stack.ProjectPath, "stack.env"), result)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dir, "stack.env"))
	f, _ := os.Open(path.Join(dir, "stack.env"))
	content, _ := io.ReadAll(f)

	assert.Equal(t, []byte("VAR1=VAL1\nVAR2=VAL2\n\nVAR1=NEW_VAL1\nVAR3=VAL3\n"), content)
}

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

func Test_createEnvFile_withPGPSecrets(t *testing.T) {
	dir := t.TempDir()
	
	// Generate test key pair
	privateKey, publicKey := generateTestKeyPair(t)
	
	// Set the PGP private key in environment
	originalKey := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Setenv("PORTAINER_PGP_PRIVATE_KEY", privateKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalKey)
		} else {
			os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
		}
	}()

	// Create encrypted secrets file
	secretsContent := "SECRET_DB_PASSWORD=supersecret123\nAPI_KEY=abc123xyz\n"
	encryptedData := encryptTestData(t, secretsContent, publicKey)
	
	err := os.WriteFile(path.Join(dir, "stack.secrets.env.pgp"), encryptedData, 0600)
	require.NoError(t, err)

	stack := &portainer.Stack{
		ProjectPath: dir,
		EntryPoint:  "docker-compose.yml",
		Env: []portainer.Pair{
			{Name: "REGULAR_VAR", Value: "regular_value"},
		},
	}

	result, err := testCreateEnvFile(stack)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(stack.ProjectPath, "stack.env"), result)

	// Read the created env file
	content, err := os.ReadFile(result)
	require.NoError(t, err)

	contentStr := string(content)
	
	// Verify that regular env vars are present
	assert.Contains(t, contentStr, "REGULAR_VAR=regular_value")
	
	// Verify that PGP secrets are present
	assert.Contains(t, contentStr, "SECRET_DB_PASSWORD=supersecret123")
	assert.Contains(t, contentStr, "API_KEY=abc123xyz")
	assert.Contains(t, contentStr, "# PGP-encrypted secrets")
}

func Test_createEnvFile_withPGPSecretsNoPGPKey(t *testing.T) {
	dir := t.TempDir()
	
	// Ensure no PGP key is set
	originalKey := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalKey)
		}
	}()

	// Create fake encrypted secrets file
	err := os.WriteFile(path.Join(dir, "stack.secrets.env.pgp"), []byte("encrypted content"), 0600)
	require.NoError(t, err)

	stack := &portainer.Stack{
		ProjectPath: dir,
		EntryPoint:  "docker-compose.yml",
		Env: []portainer.Pair{
			{Name: "REGULAR_VAR", Value: "regular_value"},
		},
	}

	result, err := testCreateEnvFile(stack)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(stack.ProjectPath, "stack.env"), result)

	// Read the created env file
	content, err := os.ReadFile(result)
	require.NoError(t, err)

	contentStr := string(content)
	
	// Verify that regular env vars are present
	assert.Contains(t, contentStr, "REGULAR_VAR=regular_value")
	
	// Verify that PGP secrets are NOT present (should have failed silently)
	assert.NotContains(t, contentStr, "# PGP-encrypted secrets")
}

func Test_createEnvFile_onlyPGPSecrets(t *testing.T) {
	dir := t.TempDir()
	
	// Generate test key pair
	privateKey, publicKey := generateTestKeyPair(t)
	
	// Set the PGP private key in environment
	originalKey := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Setenv("PORTAINER_PGP_PRIVATE_KEY", privateKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalKey)
		} else {
			os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
		}
	}()

	// Create encrypted secrets file
	secretsContent := "SECRET_VAR=secret_value\n"
	encryptedData := encryptTestData(t, secretsContent, publicKey)
	
	err := os.WriteFile(path.Join(dir, "stack.secrets.env.pgp"), encryptedData, 0600)
	require.NoError(t, err)

	stack := &portainer.Stack{
		ProjectPath: dir,
		EntryPoint:  "docker-compose.yml",
		Env:         []portainer.Pair{}, // No regular env vars
	}

	result, err := testCreateEnvFile(stack)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(stack.ProjectPath, "stack.env"), result)

	// Read the created env file
	content, err := os.ReadFile(result)
	require.NoError(t, err)

	contentStr := string(content)
	
	// Verify that PGP secrets are present
	assert.Contains(t, contentStr, "SECRET_VAR=secret_value")
	assert.Contains(t, contentStr, "# PGP-encrypted secrets")
}

func Test_testCopyPGPSecretsFile(t *testing.T) {
	// Generate test key pair
	privateKey, publicKey := generateTestKeyPair(t)
	
	// Test data
	secretsContent := "TEST_SECRET=test_value\nANOTHER_SECRET=another_value\n"
	encryptedData := encryptTestData(t, secretsContent, publicKey)
	
	// Create temporary encrypted file
	tempFile, err := os.CreateTemp("", "test-secrets-*.pgp")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	
	_, err = tempFile.Write(encryptedData)
	require.NoError(t, err)
	err = tempFile.Close()
	require.NoError(t, err)

	// Set the PGP private key in environment
	originalKey := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Setenv("PORTAINER_PGP_PRIVATE_KEY", privateKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalKey)
		} else {
			os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
		}
	}()

	// Test copying PGP secrets
	var buffer strings.Builder
	err = testCopyPGPSecretsFile(&buffer, tempFile.Name())
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "# PGP-encrypted secrets")
	assert.Contains(t, output, "TEST_SECRET=test_value")
	assert.Contains(t, output, "ANOTHER_SECRET=another_value")
}
