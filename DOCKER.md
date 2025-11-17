# Docker Deployment Guide

This guide explains how to run Fail2Rest V2 in a Docker container while maintaining access to fail2ban.

## Prerequisites

- Docker and Docker Compose installed
- Fail2ban installed and running on the host system
- Access to `/var/run/fail2ban/fail2ban.sock` on the host

## Quick Start

### Option 1: Using Docker Compose (Recommended)

1. **Create your config file:**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

2. **Set up fail2ban socket permissions:**
   ```bash
   # Make socket readable by group (if using non-root user)
   sudo chmod g+r /var/run/fail2ban/fail2ban.sock
   sudo chgrp fail2ban /var/run/fail2ban/fail2ban.sock
   ```

3. **Start the container:**
   ```bash
   docker-compose up -d
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f
   ```

### Option 2: Using Docker directly

1. **Build the image:**
   ```bash
   docker build -t fail2rest:latest .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name fail2rest \
     --restart unless-stopped \
     -p 8080:8080 \
     -v /var/run/fail2ban/fail2ban.sock:/var/run/fail2ban/fail2ban.sock:ro \
     -v $(pwd)/config.yaml:/etc/fail2rest/config.yaml:ro \
     fail2rest:latest
   ```

## Configuration Options

### Running as Root (Simplest)

If you run the container as root, it can directly access the fail2ban socket:

```yaml
# In docker-compose.yml
user: "0:0"
```

**Pros:**
- Simplest setup
- No permission issues

**Cons:**
- Less secure
- Container runs as root

### Running as Non-Root (Recommended)

For better security, run as a non-root user:

1. **On the host, set up socket permissions:**
   ```bash
   # Create fail2ban group if it doesn't exist
   sudo groupadd -f fail2ban
   
   # Make socket group-readable
   sudo chmod g+r /var/run/fail2ban/fail2ban.sock
   sudo chgrp fail2ban /var/run/fail2ban/fail2ban.sock
   ```

2. **In docker-compose.yml, use the fail2ban group:**
   ```yaml
   user: "1000:$(id -g fail2ban)"
   ```

   Or manually specify:
   ```yaml
   user: "1000:999"  # Adjust GID based on your system
   ```

3. **Update Dockerfile to add user to fail2ban group:**
   ```dockerfile
   RUN addgroup -g 999 fail2ban && \
       adduser -D -u 1000 -G fail2ban fail2rest
   ```

### Using Sudo (Alternative)

If you prefer using sudo instead of direct socket access:

1. **In your config.yaml:**
   ```yaml
   fail2ban:
     client_path: "/usr/bin/fail2ban-client"
     use_sudo: true
   ```

2. **Configure passwordless sudo in the container:**
   Add to Dockerfile:
   ```dockerfile
   RUN echo "fail2rest ALL=(ALL) NOPASSWD: /usr/bin/fail2ban-client" >> /etc/sudoers
   ```

## Network Configuration

### Bridge Network (Default)
The container runs in a bridge network, accessible via port mapping:
```yaml
ports:
  - "8080:8080"
```

### Host Network
For direct host network access:
```yaml
network_mode: "host"
```
Then bind to `127.0.0.1` in config to only listen locally.

## Volume Mounts

### Required
- **Config file:** `./config.yaml:/etc/fail2rest/config.yaml:ro`
- **Fail2ban socket:** `/var/run/fail2ban/fail2ban.sock:/var/run/fail2ban/fail2ban.sock:ro`

### Optional
- **Logs directory:** `./logs:/app/logs`
- **Custom config location:** `/etc/fail2rest/config.yaml`

## Security Considerations

1. **Read-only socket mount:** The socket is mounted read-only (`:ro`) for safety
2. **Non-root user:** Run as non-root user when possible
3. **Config file:** Mount config as read-only
4. **Network:** Use firewall rules to restrict API access
5. **TLS:** Enable TLS in production:
   ```yaml
   server:
     tls:
       enabled: true
       cert_file: "/path/to/cert.pem"
       key_file: "/path/to/key.pem"
   ```
   Mount certificates as volumes.

## Troubleshooting

### Permission Denied Errors

If you see permission errors:

1. **Check socket permissions:**
   ```bash
   ls -l /var/run/fail2ban/fail2ban.sock
   ```

2. **Check container user:**
   ```bash
   docker exec fail2rest id
   ```

3. **Verify group membership:**
   ```bash
   docker exec fail2rest groups
   ```

### Socket Not Found

If the socket doesn't exist:

1. **Check if fail2ban is running:**
   ```bash
   sudo systemctl status fail2ban
   ```

2. **Start fail2ban:**
   ```bash
   sudo systemctl start fail2ban
   ```

3. **Verify socket exists:**
   ```bash
   ls -l /var/run/fail2ban/fail2ban.sock
   ```

### Container Can't Access Host Socket

If using non-root user, ensure:
- Socket has group read permissions
- Container user is in the fail2ban group
- Group IDs match between host and container

## Production Deployment

### Using Docker Compose

```yaml
version: '3.8'

services:
  fail2rest:
    build: .
    restart: always
    ports:
      - "127.0.0.1:8080:8080"  # Only bind to localhost
    volumes:
      - /var/run/fail2ban/fail2ban.sock:/var/run/fail2ban/fail2ban.sock:ro
      - /etc/fail2rest/config.yaml:/etc/fail2rest/config.yaml:ro
      - /etc/ssl/certs/fail2rest:/etc/ssl/certs/fail2rest:ro
    environment:
      - TZ=UTC
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Using Systemd with Docker

Create `/etc/systemd/system/fail2rest.service`:

```ini
[Unit]
Description=Fail2Rest V2 Docker Container
Requires=docker.service
After=docker.service fail2ban.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/fail2rest
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

## Building Custom Images

### With Custom Config

```dockerfile
FROM fail2rest:latest
COPY config.yaml /etc/fail2rest/config.yaml
```

### Multi-stage Build for Smaller Images

The Dockerfile already uses multi-stage builds to minimize image size.

## Health Checks

The container includes a health check that pings `/health` endpoint:

```bash
# Check container health
docker ps
# Look for "healthy" status

# Manual health check
docker exec fail2rest wget -q -O- http://localhost:8080/health
```

## Logs

View container logs:
```bash
# Docker Compose
docker-compose logs -f

# Docker
docker logs -f fail2rest
```

## Updating

1. **Pull/build new image:**
   ```bash
   docker-compose build
   # or
   docker build -t fail2rest:latest .
   ```

2. **Restart container:**
   ```bash
   docker-compose up -d
   # or
   docker restart fail2rest
   ```

