# PGP Secrets Demo

This directory contains a sample Docker Compose stack that demonstrates the PGP secrets feature.

## Files

- `docker-compose.yml` - The main compose file that references environment variables
- `stack.env` - Regular (non-sensitive) environment variables for GitOps stacks
- `stack.secrets.env` - Example of what the decrypted secrets would look like
- `stack.secrets.env.pgp` - Encrypted secrets file (this is what gets committed to Git)
- `private.key` - Demo PGP private key (ASCII armored format)
- `html/index.html` - Simple web page served by nginx

## Demo Credentials

- **PGP Key**: demo@portainer.io
- **Passphrase**: demo123
- **Private Key File**: private.key (included in this demo)

## Quick Setup

Use the included demo key:

1. Copy the private key:
   ```bash
   export PORTAINER_PGP_PRIVATE_KEY="$(cat private.key)"
   ```

2. Or configure in Portainer UI:
   - Go to Settings → PGP Settings
   - Paste the contents of private.key into the "PGP Private Key" field
   - Enter demo123 as the passphrase
   - Save settings

## Usage

Deploy the stack through Portainer's GitOps functionality

## What Happens

When Portainer deploys this stack:

1. It reads stack.env for regular environment variables
2. It detects stack.secrets.env.pgp and decrypts it using the configured PGP key
3. The decrypted secrets are merged with other environment variables
4. The stack is deployed with all variables available

The final environment will contain:
- PUBLIC_VAR=hello_world (from stack.env)
- APP_NAME=PGP Demo Stack (from stack.env)
- DEBUG=false (from stack.env)
- DATABASE_PASSWORD=super_secret_db_password_123 (from decrypted secrets)
- API_TOKEN=sk_live_abc123xyz789_secret_token (from decrypted secrets)
- WEBHOOK_SECRET=whsec_example_webhook_secret_456 (from decrypted secrets)

## Security Note

In a real scenario:
- The stack.secrets.env file should be deleted after encryption
- Only the stack.secrets.env.pgp file should be committed to Git
- The private key should be stored securely and not in the repository
- Use strong passphrases in production (not demo123!)

## Testing the Demo

You can test the PGP decryption functionality:

```bash
# Test decryption with the demo key
gpg --decrypt stack.secrets.env.pgp

# Test with environment variable
export PORTAINER_PGP_PRIVATE_KEY="$(cat private.key)"
```
