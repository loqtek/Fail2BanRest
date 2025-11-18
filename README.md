# Fail2Rest V2

A simple, light, secure REST API for managing and monitoring Fail2ban servers, written in Go.

## Features

- **Jail Management**: List, view status, and manage Fail2ban jails
- **IP Management**: View banned IPs, ban/unban IP addresses
- **Statistics**: Get detailed statistics about Fail2ban operations
- **Status Monitoring**: Check Fail2ban service status
- **Secure Authentication**: JWT-based authentication with configurable tokens
- **HTTPS Support**: Secure communication with TLS

## Installation

### Option 1: Automated Install Script (Recommended)

Install with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/loqtek/Fail2BanRest/main/install.sh | bash
```

This will:
- Install all dependencies
- Build the application
- Set up systemd service
- Create configuration file
- Start the service

See [INSTALL.md](INSTALL.md) for detailed installation instructions.

**Uninstall:**
```bash
curl -fsSL https://raw.githubusercontent.com/loqtek/Fail2BanRest/main/install.sh | bash -s uninstall
```

### Option 2: Docker

See [DOCKER.md](DOCKER.md) for detailed Docker setup instructions.

Quick start:
```bash
# Create config
cp config.example.yaml config.yaml
# Edit config.yaml

# Start with Docker Compose
docker-compose up -d
```

### Option 3: Manual Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build the application:
   ```bash
   go build -o fail2restV2 ./cmd/server
   go build -o hash-password ./cmd/hash-password
   ```

## Configuration

Create a `config.yaml` file (see `config.example.yaml` for template):

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

auth:
  jwt_secret: "your-secret-key-change-this"
  token_expiry: 24h
  
  # API Keys for authentication (generate with: openssl rand -hex 32)
  api_keys:
    - "your-secure-api-key-here"
  
  # User accounts (passwords must be bcrypt hashed)
  users:
    - username: "admin"
      password: "$2a$10$..."  # Generate with: ./hash-password -password yourpassword

fail2ban:
  client_path: "/usr/bin/fail2ban-client"
```

### Setting Up Authentication

**Option 1: API Keys** (Recommended for automation/server-to-server)
1. Generate a secure API key:
   ```bash
   openssl rand -hex 32
   ```
2. Add it to `api_keys` in your config file

**Option 2: Username/Password** (For interactive use)
1. Hash your password:
   ```bash
   go build -o hash-password ./cmd/hash-password
   ./hash-password -password yourpassword
   ```
2. Copy the hashed output and add it to `users` in your config file

**Note:** You must configure at least one authentication method (API keys or users) for the server to start.

### Fail2ban Permissions

Fail2ban requires root privileges to access its socket. You have three options:

**Option 1: Run as root** (Simplest, but less secure)
```bash
sudo ./fail2restV2
```

**Option 2: Use sudo** (Recommended for production)
1. Configure passwordless sudo for fail2ban-client:
   ```bash
   sudo visudo
   ```
2. Add this line (replace `youruser` with your actual username):
   ```
   youruser ALL=(ALL) NOPASSWD: /usr/bin/fail2ban-client
   ```
3. Set `use_sudo: true` in your config.yaml:
   ```yaml
   fail2ban:
     client_path: "/usr/bin/fail2ban-client"
     use_sudo: true
   ```

**Option 3: Create a dedicated system user** (Most secure)
1. Create a system user:
   ```bash
   sudo useradd -r -s /bin/false fail2rest
   ```
2. Configure sudo for this user:
   ```bash
   sudo visudo
   ```
   Add:
   ```
   fail2rest ALL=(ALL) NOPASSWD: /usr/bin/fail2ban-client
   ```
3. Run the service as this user (via systemd, supervisor, etc.)

## Usage

Run the server:
```bash
./fail2restV2
```

Or with custom config:
```bash
./fail2restV2 -config /path/to/config.yaml
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - Get JWT token (requires API key or username/password)

### Status
- `GET /api/v1/status` - Get Fail2ban service status

### Jails
- `GET /api/v1/jails` - List all jails
- `GET /api/v1/jails/:name` - Get jail details
- `GET /api/v1/jails/:name/status` - Get jail status

### Banned IPs
- `GET /api/v1/jails/:name/banned` - List banned IPs for a jail
- `POST /api/v1/jails/:name/ban` - Ban an IP address
- `POST /api/v1/jails/:name/unban` - Unban an IP address

### Statistics
- `GET /api/v1/stats` - Get overall statistics
- `GET /api/v1/jails/:name/stats` - Get statistics for a specific jail

## Troubleshooting

### Permission Denied Error

If you see an error like:
```
Permission denied to socket: /var/run/fail2ban/fail2ban.sock, (you must be root)
```

**Quick Fix:** Enable sudo in your config:
```yaml
fail2ban:
  use_sudo: true
```

Then configure passwordless sudo (see "Fail2ban Permissions" section above).

**Alternative:** Run the server as root (not recommended for production):
```bash
sudo ./fail2restV2
```

## Security

- All endpoints (except `/auth/login`) require JWT authentication
- Use HTTPS in production
- Keep your JWT secret secure
- Run with appropriate system permissions to execute fail2ban-client

## License

MIT

