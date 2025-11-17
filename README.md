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

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Build the application:
   ```bash
   go build -o fail2restV2
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

fail2ban:
  client_path: "/usr/bin/fail2ban-client"
```

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
- `POST /api/v1/auth/login` - Get JWT token

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

## Security

- All endpoints (except `/auth/login`) require JWT authentication
- Use HTTPS in production
- Keep your JWT secret secure
- Run with appropriate system permissions to execute fail2ban-client

## License

MIT

