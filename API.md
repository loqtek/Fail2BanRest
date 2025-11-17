# Fail2Rest V2 API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication

All endpoints except `/auth/login` require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Authentication

#### POST /auth/login
Get a JWT token for API access.

**Request Body:**
```json
{
  "token": "any-token-value"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### Status

#### GET /status
Get the overall status of fail2ban service.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": {
      "Number of jail": "3",
      "Jail list": "sshd, nginx-http-auth, apache-auth"
    },
    "timestamp": 1704110400
  }
}
```

---

### Jails

#### GET /jails
List all configured jails.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "sshd"
    },
    {
      "name": "nginx-http-auth"
    }
  ]
}
```

#### GET /jails/:name
Get detailed information about a specific jail.

**Response:**
```json
{
  "success": true,
  "data": {
    "name": "sshd",
    "status": {
      "Status": "Active",
      "Filter": "sshd",
      "Currently banned": "5",
      "Total banned": "42"
    }
  }
}
```

#### GET /jails/:name/status
Get the status of a specific jail.

**Response:**
```json
{
  "success": true,
  "data": {
    "Status": "Active",
    "Filter": "sshd",
    "Currently banned": "5",
    "Total banned": "42"
  }
}
```

#### POST /jails/:name/start
Start a jail.

**Response:**
```json
{
  "success": true,
  "message": "Jail started successfully"
}
```

#### POST /jails/:name/stop
Stop a jail.

**Response:**
```json
{
  "success": true,
  "message": "Jail stopped successfully"
}
```

#### POST /jails/:name/restart
Restart a jail.

**Response:**
```json
{
  "success": true,
  "message": "Jail restarted successfully"
}
```

#### POST /jails/:name/reload
Reload a jail configuration.

**Response:**
```json
{
  "success": true,
  "message": "Jail reloaded successfully"
}
```

---

### IP Management

#### GET /jails/:name/banned
Get list of banned IPs for a jail.

**Response:**
```json
{
  "success": true,
  "data": {
    "jail": "sshd",
    "banned_ips": [
      "192.168.1.100",
      "10.0.0.50"
    ]
  }
}
```

#### POST /jails/:name/ban
Ban an IP address in a jail.

**Request Body:**
```json
{
  "ip": "192.168.1.100"
}
```

**Response:**
```json
{
  "success": true,
  "message": "IP banned successfully",
  "data": {
    "jail": "sshd",
    "ip": "192.168.1.100"
  }
}
```

#### POST /jails/:name/unban
Unban an IP address in a jail.

**Request Body:**
```json
{
  "ip": "192.168.1.100"
}
```

**Response:**
```json
{
  "success": true,
  "message": "IP unbanned successfully",
  "data": {
    "jail": "sshd",
    "ip": "192.168.1.100"
  }
}
```

---

### Statistics

#### GET /stats
Get overall statistics for all jails.

**Response:**
```json
{
  "success": true,
  "data": {
    "stats": {
      "jail_count": 3,
      "jails": ["sshd", "nginx-http-auth", "apache-auth"],
      "total_banned_ips": 15,
      "jail_details": {
        "sshd": {
          "currently_banned": "5",
          "total_banned": "42"
        }
      },
      "timestamp": 1704110400
    },
    "timestamp": 1704110400
  }
}
```

#### GET /jails/:name/stats
Get statistics for a specific jail.

**Response:**
```json
{
  "success": true,
  "data": {
    "stats": {
      "filter": "sshd",
      "currently_banned": "5",
      "total_banned": "42",
      "banned_ips": ["192.168.1.100", "10.0.0.50"]
    },
    "timestamp": 1704110400
  }
}
```

---

## Error Responses

All endpoints return errors in the following format:

```json
{
  "success": false,
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid token)
- `404` - Not Found (jail not found)
- `500` - Internal Server Error

---

## Example Usage

### Using curl

1. **Get a token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"token": "my-secret-token"}'
```

2. **List jails:**
```bash
curl -X GET http://localhost:8080/api/v1/jails \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

3. **Ban an IP:**
```bash
curl -X POST http://localhost:8080/api/v1/jails/sshd/ban \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ip": "192.168.1.100"}'
```

4. **Get statistics:**
```bash
curl -X GET http://localhost:8080/api/v1/stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

