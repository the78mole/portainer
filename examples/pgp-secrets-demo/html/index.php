<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PGP Secrets Demo - Status</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #007cbb;
            padding-bottom: 10px;
        }
        .status {
            margin: 20px 0;
            padding: 15px;
            border-radius: 5px;
            font-weight: bold;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .warning {
            background-color: #fff3cd;
            color: #856404;
            border: 1px solid #ffeaa7;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .env-section {
            margin: 20px 0;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 5px;
            border-left: 4px solid #007cbb;
        }
        .env-item {
            margin: 10px 0;
            font-family: monospace;
            background-color: white;
            padding: 8px;
            border-radius: 3px;
            border: 1px solid #dee2e6;
        }
        .env-name {
            font-weight: bold;
            color: #007cbb;
        }
        .env-value {
            color: #333;
            margin-left: 10px;
        }
        .masked {
            color: #6c757d;
            font-style: italic;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #dee2e6;
            color: #6c757d;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 PGP Secrets Demo - Status</h1>
        
        <?php
        // Check if secrets are loaded
        $publicVar = getenv('PUBLIC_VAR');
        $appName = getenv('APP_NAME');
        $debug = getenv('DEBUG');
        
        // Secret variables (from encrypted file)
        $dbPassword = getenv('DATABASE_PASSWORD');
        $apiToken = getenv('API_TOKEN');
        $webhookSecret = getenv('WEBHOOK_SECRET');
        
        $secretsLoaded = !empty($dbPassword) && !empty($apiToken) && !empty($webhookSecret);
        $publicVarsLoaded = !empty($publicVar) && !empty($appName);
        ?>
        
        <div class="status <?php echo $secretsLoaded && $publicVarsLoaded ? 'success' : ($publicVarsLoaded ? 'warning' : 'error'); ?>">
            <?php if ($secretsLoaded && $publicVarsLoaded): ?>
                ✅ SUCCESS: PGP secrets were successfully decrypted and loaded!
            <?php elseif ($publicVarsLoaded): ?>
                ⚠️ PARTIAL: Public variables loaded, but PGP secrets are missing!
            <?php else: ?>
                ❌ ERROR: No environment variables loaded. Check your configuration.
            <?php endif; ?>
        </div>
        
        <div class="env-section">
            <h3>📂 Public Environment Variables (from stack.env)</h3>
            <div class="env-item">
                <span class="env-name">PUBLIC_VAR:</span>
                <span class="env-value"><?php echo htmlspecialchars($publicVar ?: 'NOT SET'); ?></span>
            </div>
            <div class="env-item">
                <span class="env-name">APP_NAME:</span>
                <span class="env-value"><?php echo htmlspecialchars($appName ?: 'NOT SET'); ?></span>
            </div>
            <div class="env-item">
                <span class="env-name">DEBUG:</span>
                <span class="env-value"><?php echo htmlspecialchars($debug ?: 'NOT SET'); ?></span>
            </div>
        </div>
        
        <div class="env-section">
            <h3>🔒 Secret Environment Variables (from stack.secrets.env.pgp)</h3>
            <div class="env-item">
                <span class="env-name">DATABASE_PASSWORD:</span>
                <span class="env-value <?php echo empty($dbPassword) ? 'masked' : ''; ?>">
                    <?php echo empty($dbPassword) ? 'NOT SET' : '***' . substr($dbPassword, -3) . ' (loaded successfully)'; ?>
                </span>
            </div>
            <div class="env-item">
                <span class="env-name">API_TOKEN:</span>
                <span class="env-value <?php echo empty($apiToken) ? 'masked' : ''; ?>">
                    <?php echo empty($apiToken) ? 'NOT SET' : substr($apiToken, 0, 8) . '***' . substr($apiToken, -4) . ' (loaded successfully)'; ?>
                </span>
            </div>
            <div class="env-item">
                <span class="env-name">WEBHOOK_SECRET:</span>
                <span class="env-value <?php echo empty($webhookSecret) ? 'masked' : ''; ?>">
                    <?php echo empty($webhookSecret) ? 'NOT SET' : substr($webhookSecret, 0, 6) . '***' . substr($webhookSecret, -3) . ' (loaded successfully)'; ?>
                </span>
            </div>
        </div>
        
        <div class="env-section">
            <h3>🔍 Verification Details</h3>
            <div class="env-item">
                <span class="env-name">Timestamp:</span>
                <span class="env-value"><?php echo date('Y-m-d H:i:s T'); ?></span>
            </div>
            <div class="env-item">
                <span class="env-name">Container Hostname:</span>
                <span class="env-value"><?php echo htmlspecialchars(gethostname()); ?></span>
            </div>
            <div class="env-item">
                <span class="env-name">PGP Decryption Status:</span>
                <span class="env-value <?php echo $secretsLoaded ? 'success' : 'error'; ?>">
                    <?php echo $secretsLoaded ? '✅ Secrets successfully decrypted' : '❌ Secrets not loaded'; ?>
                </span>
            </div>
        </div>
        
        <?php if ($secretsLoaded): ?>
        <div class="env-section">
            <h3>🧪 Demo Verification</h3>
            <p>You can verify that the secrets are working by checking:</p>
            <ul>
                <li>Database connection would use: <code>***<?php echo substr($dbPassword, -3); ?></code></li>
                <li>API calls would use token: <code><?php echo substr($apiToken, 0, 8); ?>***</code></li>
                <li>Webhooks would use secret: <code><?php echo substr($webhookSecret, 0, 6); ?>***</code></li>
            </ul>
        </div>
        <?php endif; ?>
        
        <div class="footer">
            <strong>Note:</strong> In production, never display actual secret values on web pages. 
            This demo shows masked values to verify that secrets are loaded without exposing them.
            <br><br>
            <strong>Portainer PGP Secrets Demo</strong> - Secrets are automatically decrypted from 
            <code>stack.secrets.env.pgp</code> using the configured PGP private key.
        </div>
    </div>
</body>
</html>