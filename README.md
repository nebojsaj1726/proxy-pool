# Proxy Pool API

A Go-based proxy pool and authentication service with an optional web dashboard.  
Manages HTTP proxies, performs health checks, stores their state in SQLite, and provides authenticated endpoints for allocation and monitoring.

---

## Features

- Authenticated proxy allocation and monitoring API
- Automatic health checks for proxies
- Maintains proxy stats: alive/dead status, score, usage, success/fail counts, latency
- SQLite database storage
- Optional web dashboard for visualization and proxy management
- CLI-friendly: interact via `curl` or any HTTP client

---

## Installation

1. **Clone the repository**

```bash
git clone https://github.com/yourusername/proxy-pool.git
cd proxy-pool
```

2. **Install Go dependencies**

```bash
go mod tidy
```

3. **Create your configuration file**

Create `config.yaml` in the project root. You can refer to `config.example.yaml` for the required structure — make sure you enter your own proxy URLs.

4. **Create your env file**

Create `.env` file in the project root. You can copy `.env.example` for the required variables.

---

## Usage

### Run the proxy API server

Start the API server with:

```bash
make run
```

This starts the API server on `http://localhost:8080`.

### Interact with API

#### Login

```bash
curl -X POST http://localhost:8080/auth/login \
	-H "Content-Type: application/json" \
	-d '{"username":"admin","password":"password"}'
```

#### Register

```bash
curl -X POST http://localhost:8080/auth/register \
	-H "Content-Type: application/json" \
	-d '{"username":"admin","password":"password"}'
```

#### List proxies

```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/proxies
```

#### Allocate a proxy

```bash
curl -X POST -H "Authorization: Bearer <TOKEN>" http://localhost:8080/allocate
```

#### Get proxy statistics

```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/proxies/stats
```

### Optional: Run the web dashboard

```bash
make web
```

The web dashboard allows you to:

- View proxies and their status
- Allocate a proxy with one click
- View historical proxy stats
