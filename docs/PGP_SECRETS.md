# PGP Secrets Support for GitOps Stacks

This feature adds support for PGP-encrypted secrets in GitOps stacks, allowing you to securely store sensitive environment variables in your Git repositories.

## Overview

When deploying stacks from Git repositories, Portainer now supports automatic decryption of PGP-encrypted secret files. This allows you to:

1. Store sensitive data (API keys, passwords, etc.) encrypted in your Git repository
2. Keep secrets separate from regular environment variables
3. Maintain security while using GitOps workflows

## How It Works

1. **Setup**: Configure a PGP private key in Portainer
2. **Encrypt**: Create a `stack.secrets.env.pgp` file with your encrypted secrets
3. **Deploy**: Deploy your stack - Portainer automatically decrypts and loads the secrets

## Configuration

### PGP Private Key Setup

Set the PGP private key using the environment variable:

```bash
export PORTAINER_PGP_PRIVATE_KEY="-----BEGIN PGP PRIVATE KEY BLOCK-----
...your private key content...
-----END PGP PRIVATE KEY BLOCK-----"
```

### File Structure

Your Git repository should contain:

```
my-stack/
├── docker-compose.yml        # Your compose file
├── stack.env                # Regular environment variables (optional)
└── stack.secrets.env.pgp    # PGP-encrypted secrets
```

**Note**: For GitOps repository-based stacks, Portainer uses `stack.env` as the default environment file name, not `.env`.

## Usage Example

### 1. Generate PGP Key Pair

```bash
# Generate a new key pair
gpg --gen-key

# Export public key for encryption
gpg --armor --export your-email@example.com > public.key

# Export private key for Portainer
gpg --armor --export-secret-keys your-email@example.com > private.key
```

### 2. Create Secrets File

Create a plain text file with your secrets:

```bash
# secrets.env
DATABASE_PASSWORD=super_secret_password
API_TOKEN=abc123xyz789
WEBHOOK_SECRET=my_webhook_secret
```

### 3. Encrypt the Secrets

```bash
# Encrypt the secrets file
gpg --armor --encrypt --recipient your-email@example.com secrets.env
mv secrets.env.asc stack.secrets.env.pgp
```

### 4. Configure GPG Agent (Optional but Recommended)

To avoid entering your passphrase repeatedly, add your key to the GPG agent:

```bash
# Add your private key to the GPG agent
gpg --import private.key

# Start gpg-agent and add to your shell profile
echo 'eval $(gpg-agent --daemon)' >> ~/.bashrc
source ~/.bashrc

# Pre-load the key in the agent (you'll be prompted for passphrase once)
echo "test" | gpg --clearsign --default-key your-email@example.com > /dev/null

# Optional: Configure agent to cache passphrase longer (default is 10 minutes)
echo "default-cache-ttl 28800" >> ~/.gnupg/gpg-agent.conf  # 8 hours
echo "max-cache-ttl 86400" >> ~/.gnupg/gpg-agent.conf      # 24 hours
gpg-connect-agent reloadagent /bye
```

### 5. Configure Portainer

Set the private key in Portainer's environment:

```bash
export PORTAINER_PGP_PRIVATE_KEY="$(cat private.key)"
```

### 6. Deploy Stack

When you deploy a stack from Git, Portainer will:

1. Clone the repository
2. Detect the `stack.secrets.env.pgp` file
3. Decrypt it using the configured private key
4. Merge the secrets with other environment variables
5. Deploy the stack with all variables available

## File Processing Order

Environment variables are processed in this order (later values override earlier ones):

1. Default `stack.env` file from the repository
2. Environment variables configured in Portainer UI
3. Decrypted secrets from `stack.secrets.env.pgp`

## Security Considerations

- **Private Key Storage**: Store the private key securely and restrict access
- **Key Rotation**: Regularly rotate your PGP keys
- **Repository Access**: Even encrypted, limit access to your Git repositories
- **Logging**: Decrypted secrets are not logged by Portainer

## Error Handling

If PGP decryption fails:

- A warning is logged but stack deployment continues
- Only the encrypted file is skipped - other environment variables are still loaded
- Check Portainer logs for specific error messages

## Troubleshooting

### Common Issues

1. **"PGP private key not found"**
   - Ensure `PORTAINER_PGP_PRIVATE_KEY` environment variable is set
   - Verify the private key format is correct

2. **"Failed to decrypt PGP secrets file"**
   - Verify the file was encrypted with the correct public key
   - Check that the private key matches the public key used for encryption

3. **"Invalid PGP private key"**
   - Ensure the private key is in ASCII armored format
   - Verify the key is not corrupted or incomplete

### Debug Tips

- Check Portainer logs for detailed error messages
- Verify PGP file can be decrypted manually: `gpg --decrypt stack.secrets.env.pgp`
- Test with a simple secrets file first

## Backward Compatibility

This feature is fully backward compatible:

- Existing stacks without PGP files work unchanged
- Regular environment variables continue to work as before
- No configuration is required if you don't use PGP secrets