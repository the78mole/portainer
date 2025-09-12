# PGP Secrets Demo

This directory contains a sample Docker Compose stack that demonstrates the PGP secrets feature.

## Files

- `docker-compose.yml` - The main compose file that references environment variables
- `Dockerfile` - Custom Docker image with PHP/Apache for serving the demo
- `stack.env` - Regular (non-sensitive) environment variables for GitOps stacks
- `stack.secrets.env` - Example of what the decrypted secrets would look like
- `stack.secrets.env.pgp` - Encrypted secrets file (this is what gets committed to Git)
- `private.key` - Demo PGP private key (ASCII armored format)
- `html/index.php` - PHP web page that displays environment variables and verifies PGP decryption

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

### Option 1: GitOps Deployment (Recommended)

Deploy the stack through Portainer's GitOps functionality:

1. **Configure PGP Settings in Portainer:**
   - Navigate to **Settings** → **PGP Settings**
   - Paste the contents of `private.key` into the "PGP Private Key" field
   - Enter `demo123` as the passphrase
   - Save the settings

2. **Create GitOps Stack:**
   - Go to **Stacks** → **Add Stack**
   - Select **Git Repository**
   - Configure the repository:
     - **Repository URL**: `https://github.com/the78mole/portainer.git`
     - **Repository reference**: `copilot/fix-7182f72f-a66a-482f-b496-bf0986cdb51c`
     - **Compose path**: `examples/pgp-secrets-demo/docker-compose.yml`
   - Set environment files:
     - **Additional files**: `examples/pgp-secrets-demo/stack.env`
     - **PGP encrypted secrets**: `examples/pgp-secrets-demo/stack.secrets.env.pgp` (automatically detected)
   - Deploy the stack

3. **Verify PGP Decryption:**
   - Once deployed, visit: `http://localhost:8080`
   - The web interface will show:
     - ✅ All environment variables (public and decrypted secrets)
     - ✅ Color-coded status indicators
     - ✅ Verification that PGP decryption worked correctly

### Option 2: Local Docker Compose

For local testing without GitOps:

```bash
# Clone and navigate to the demo
git clone https://github.com/the78mole/portainer.git
cd portainer/examples/pgp-secrets-demo

# Build and start the stack
docker compose up --build
```

**Note:** This method won't decrypt the PGP secrets unless you manually set the environment variables.

## Verification

After successful deployment, the verification page at `http://localhost:8080` will display:

- **Public Variables** (from `stack.env`):
  - `PUBLIC_VAR=hello_world`
  - `APP_NAME=PGP Demo Stack`
  - `DEBUG=false`

- **Secret Variables** (decrypted from `stack.secrets.env.pgp`):
  - `DATABASE_PASSWORD=super_secret_db_password_123`
  - `API_TOKEN=sk_live_abc123xyz789_secret_token`
  - `WEBHOOK_SECRET=whsec_example_webhook_secret_456`

If PGP decryption is working correctly, all secret variables will show their actual values. If decryption fails, they will appear as empty or show default values.

## What Happens

When Portainer deploys this stack via GitOps:

1. **Repository Analysis**: Portainer clones the Git repository and scans for environment files
2. **Public Variables**: Reads `stack.env` for regular environment variables
3. **PGP Decryption**: Automatically detects `stack.secrets.env.pgp` and decrypts it using the configured PGP key and passphrase
4. **Variable Merging**: Merges public and decrypted secret variables
5. **Stack Deployment**: Deploys the Docker Compose stack with all variables available to containers
6. **Verification**: The web application displays all variables to verify successful decryption

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
