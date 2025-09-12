# PGP Secrets Demo

This directory contains a sample Docker Compose stack that demonstrates the PGP secrets feature.

## Files

- `docker-compose.yml` - The main compose file that references environment variables
- `stack.env` - Regular (non-sensitive) environment variables for GitOps stacks
- `secrets.env` - Example of what the decrypted secrets would look like
- `html/index.html` - Simple web page served by nginx

## Usage

1. **Generate a PGP key pair** (or use an existing one):
   ```bash
   gpg --gen-key
   ```

2. **Encrypt the secrets file**:
   ```bash
   gpg --armor --encrypt --recipient your-email@example.com secrets.env
   mv secrets.env.asc stack.secrets.env.pgp
   ```

3. **Set up Portainer with the private key**:
   ```bash
   export PORTAINER_PGP_PRIVATE_KEY="$(gpg --armor --export-secret-keys your-email@example.com)"
   ```

4. **Deploy the stack** through Portainer's GitOps functionality

## What Happens

When Portainer deploys this stack:

1. It reads `stack.env` for regular environment variables
2. It detects `stack.secrets.env.pgp` and decrypts it
3. The decrypted secrets are merged with other environment variables
4. The stack is deployed with all variables available

The final environment will contain:
- `PUBLIC_VAR=hello_world` (from stack.env)
- `APP_NAME=PGP Demo Stack` (from stack.env)
- `DEBUG=false` (from stack.env)
- `DATABASE_PASSWORD=super_secret_db_password_123` (from decrypted secrets)
- `API_TOKEN=sk_live_abc123xyz789_secret_token` (from decrypted secrets)
- `WEBHOOK_SECRET=whsec_example_webhook_secret_456` (from decrypted secrets)

## Security Note

In a real scenario:
- The `secrets.env` file should be deleted after encryption
- Only the `stack.secrets.env.pgp` file should be committed to Git
- The private key should be stored securely and not in the repository