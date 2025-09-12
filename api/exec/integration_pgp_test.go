package exec

import (
	"os"
	"path"
	"strings"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPGPSecretsIntegrationWorkflow tests the complete workflow of PGP secrets
func TestPGPSecretsIntegrationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for our "Git repository"
	gitRepoDir := t.TempDir()

	// Generate test key pair
	privateKey, publicKey := generateTestKeyPair(t)

	// Set up environment variable for private key
	originalKey := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Setenv("PORTAINER_PGP_PRIVATE_KEY", privateKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalKey)
		} else {
			os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
		}
	}()

	// Create sample docker-compose.yml
	composeContent := `version: '3.8'
services:
  web:
    image: nginx:alpine
    environment:
      - PUBLIC_VAR=${PUBLIC_VAR}
      - SECRET_DB_PASSWORD=${SECRET_DB_PASSWORD}
      - API_KEY=${API_KEY}
      - WEBHOOK_URL=${WEBHOOK_URL}
    ports:
      - "8080:80"
`
	err := os.WriteFile(path.Join(gitRepoDir, "docker-compose.yml"), []byte(composeContent), 0644)
	require.NoError(t, err)

	// Create regular .env file
	envContent := `PUBLIC_VAR=hello_world
LOG_LEVEL=info
ENVIRONMENT=test`
	err = os.WriteFile(path.Join(gitRepoDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	// Create secrets content
	secretsContent := `SECRET_DB_PASSWORD=ultra_secret_password_123
API_KEY=sk_live_super_secret_api_key_xyz
WEBHOOK_URL=https://example.com/webhook?secret=abc123`

	// Encrypt the secrets
	encryptedSecrets := encryptTestData(t, secretsContent, publicKey)
	err = os.WriteFile(path.Join(gitRepoDir, "stack.secrets.env.pgp"), encryptedSecrets, 0600)
	require.NoError(t, err)

	// Create a stack object similar to what would be created from GitOps
	stack := &portainer.Stack{
		ID:          1,
		Name:        "test-pgp-stack",
		ProjectPath: gitRepoDir,
		EntryPoint:  "docker-compose.yml",
		Env: []portainer.Pair{
			{Name: "PORTAINER_VAR", Value: "from_portainer_config"},
		},
	}

	// Test the createEnvFile function (this simulates what happens during deployment)
	envFilePath, err := testCreateEnvFile(stack)
	require.NoError(t, err)
	require.NotEmpty(t, envFilePath)

	// Verify the env file was created
	assert.FileExists(t, envFilePath)

	// Read and verify the content
	envFileContent, err := os.ReadFile(envFilePath)
	require.NoError(t, err)

	envContentStr := string(envFileContent)

	// Verify all sources of environment variables are present
	t.Log("Generated env file content:")
	t.Log(envContentStr)

	// From .env file
	assert.Contains(t, envContentStr, "PUBLIC_VAR=hello_world")
	assert.Contains(t, envContentStr, "LOG_LEVEL=info")
	assert.Contains(t, envContentStr, "ENVIRONMENT=test")

	// From Portainer configuration
	assert.Contains(t, envContentStr, "PORTAINER_VAR=from_portainer_config")

	// From decrypted PGP secrets
	assert.Contains(t, envContentStr, "SECRET_DB_PASSWORD=ultra_secret_password_123")
	assert.Contains(t, envContentStr, "API_KEY=sk_live_super_secret_api_key_xyz")
	assert.Contains(t, envContentStr, "WEBHOOK_URL=https://example.com/webhook?secret=abc123")

	// Verify the PGP section header is present
	assert.Contains(t, envContentStr, "# PGP-encrypted secrets")

	// Verify the file ends with a newline
	assert.True(t, strings.HasSuffix(envContentStr, "\n"))
}

// TestPGPSecretsWithSubdirectory tests PGP secrets when compose file is in a subdirectory
func TestPGPSecretsWithSubdirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory structure
	gitRepoDir := t.TempDir()
	appDir := path.Join(gitRepoDir, "app")
	err := os.MkdirAll(appDir, 0755)
	require.NoError(t, err)

	// Generate test key pair
	privateKey, publicKey := generateTestKeyPair(t)

	// Set up environment variable
	originalKey := os.Getenv("PORTAINER_PGP_PRIVATE_KEY")
	os.Setenv("PORTAINER_PGP_PRIVATE_KEY", privateKey)
	defer func() {
		if originalKey != "" {
			os.Setenv("PORTAINER_PGP_PRIVATE_KEY", originalKey)
		} else {
			os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
		}
	}()

	// Create docker-compose.yml in subdirectory
	composeContent := `version: '3.8'
services:
  app:
    image: alpine:latest
    environment:
      - APP_SECRET=${APP_SECRET}
`
	err = os.WriteFile(path.Join(appDir, "docker-compose.yml"), []byte(composeContent), 0644)
	require.NoError(t, err)

	// Create PGP secrets in the same subdirectory as compose file
	secretsContent := `APP_SECRET=subdirectory_secret_123`
	encryptedSecrets := encryptTestData(t, secretsContent, publicKey)
	err = os.WriteFile(path.Join(appDir, "stack.secrets.env.pgp"), encryptedSecrets, 0600)
	require.NoError(t, err)

	// Create stack with subdirectory entry point
	stack := &portainer.Stack{
		ID:          2,
		Name:        "test-subdir-stack",
		ProjectPath: gitRepoDir,
		EntryPoint:  "app/docker-compose.yml",
		Env:         []portainer.Pair{},
	}

	// Test createEnvFile
	envFilePath, err := testCreateEnvFile(stack)
	require.NoError(t, err)
	require.NotEmpty(t, envFilePath)

	// Verify content
	envFileContent, err := os.ReadFile(envFilePath)
	require.NoError(t, err)

	envContentStr := string(envFileContent)
	assert.Contains(t, envContentStr, "APP_SECRET=subdirectory_secret_123")
	assert.Contains(t, envContentStr, "# PGP-encrypted secrets")
}

// TestPGPSecretsErrorHandling tests various error scenarios
func TestPGPSecretsErrorHandling(t *testing.T) {
	testCases := []struct {
		name            string
		setupFunc       func(t *testing.T, dir string) *portainer.Stack
		expectEnvFile   bool
		expectLogEntry  string
	}{
		{
			name: "missing private key",
			setupFunc: func(t *testing.T, dir string) *portainer.Stack {
				// Unset the private key
				os.Unsetenv("PORTAINER_PGP_PRIVATE_KEY")
				
				// Create a PGP file
				err := os.WriteFile(path.Join(dir, "stack.secrets.env.pgp"), []byte("fake encrypted content"), 0600)
				require.NoError(t, err)

				return &portainer.Stack{
					ID:          3,
					Name:        "test-no-key",
					ProjectPath: dir,
					EntryPoint:  "docker-compose.yml",
					Env: []portainer.Pair{
						{Name: "REGULAR_VAR", Value: "regular_value"},
					},
				}
			},
			expectEnvFile:  true,
			expectLogEntry: "PGP private key not found",
		},
		{
			name: "corrupted PGP file",
			setupFunc: func(t *testing.T, dir string) *portainer.Stack {
				// Set a valid private key
				privateKey, _ := generateTestKeyPair(t)
				os.Setenv("PORTAINER_PGP_PRIVATE_KEY", privateKey)
				
				// Create a corrupted PGP file
				err := os.WriteFile(path.Join(dir, "stack.secrets.env.pgp"), []byte("invalid pgp content"), 0600)
				require.NoError(t, err)

				return &portainer.Stack{
					ID:          4,
					Name:        "test-corrupted",
					ProjectPath: dir,
					EntryPoint:  "docker-compose.yml",
					Env: []portainer.Pair{
						{Name: "REGULAR_VAR", Value: "regular_value"},
					},
				}
			},
			expectEnvFile:  true,
			expectLogEntry: "Failed to decrypt PGP secrets file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			stack := tc.setupFunc(t, dir)

			// Test createEnvFile
			envFilePath, err := testCreateEnvFile(stack)
			
			if tc.expectEnvFile {
				require.NoError(t, err)
				require.NotEmpty(t, envFilePath)
				
				// Verify regular env vars are still present
				envFileContent, err := os.ReadFile(envFilePath)
				require.NoError(t, err)
				assert.Contains(t, string(envFileContent), "REGULAR_VAR=regular_value")
			} else {
				// If we don't expect an env file, then there should be no env vars and no file
				assert.Empty(t, envFilePath)
			}
		})
	}
}