# Hyacinth

Hyacinth is a cloud-native, enterprise-level newsletter service built in Go, integrating PostgreSQL as a database and Redis for caching and session support. The service is designed to be secure, scalable, and highly customizable.

## Table of Contents

- [Pre-requisites](#pre-requisites)
    - [Linux](#linux)
    - [macOS](#macos)
    - [Windows](#windows)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Launch](#launch)
  - [Admin Interface](#admin-interface)
- [Contributing](#contributing)
- [License](#license)

## Pre-requisites

- [Go](https://go.dev/doc/install)
```bash
# Verify your Go installation by running the following command
go version
```

- [PostgreSQL](https://www.postgresql.org/download/)
```bash
# Verify your psql installation by running the following command
psql --version
```

### Linux
- PostgreSQL CLI
```bash
# Ubuntu 
sudo apt-get update
sudo apt-get install postgresql-client
# Arch 
sudo pacman -S postgresql
```

- Go Migrate
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### macOS
- PostgreSQL CLI
```bash
brew doctor
brew update
brew install libpq
brew link --force libpq
```

- Go Migrate
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Windows
- Go Migrate
```
# Go Migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Configuration

The Service can be customized to suit your specific requirements by cloning the `dev.yaml` file located within `api/configs` into a separate `production.yaml` file within the same directory. Below are the key configuration options and their explanations:

### Application Configuration

- `port`: The port on which the service will listen for incoming requests (e.g., `8000`).
- `host`: The host address to bind the service to (e.g., `0.0.0.0` to listen on all available network interfaces).

### Database Configuration

- `host`: The hostname or IP address of your PostgreSQL database server (e.g., `"localhost"`).
- `port`: The port on which PostgreSQL is running (e.g., `"5432"`).
- `username`: The username to connect to the PostgreSQL database (e.g., `"postgres"`).
- `password`: The password for the PostgreSQL user (e.g., `"password"`).
- `database_name`: The name of the PostgreSQL database (e.g., `"newsletter"`).

### Email Client Configuration

- `server`: The base server for your email service (e.g., `"postmark"`).
- `port`: The port on which the server is running (e.g., `597`).
- `username`: The username required for server access (e.g., `"username"`).
- `password`: The password required for server access (e.g., `"password"`).
- `sender`: The sender email registered to the server (e.g., `"name@example.com"`

### Redis Configuration

- `host`: The hostname or IP address of your Redis database server (e.g., `"localhost"`).
- `port`: The port on which Redis is running (e.g., `"6379"`).
- `conn`: The connection type (e.g., `"tcp"`).

To customize your service, open the `production.yaml` file you created within `api/configs` and update the desired values according to your environment and requirements. After making changes, be sure to rebuild and restart the service with the `-cfg production` flag (as described below) for the new configuration to take effect.

Please ensure that sensitive information such as passwords, authentication tokens, and cryptographic secrets are kept secure and are not exposed in your version control system.

## Usage
### Launch

1. Initialize the PostgreSQL and Redis containers:

```bash
./scripts/init_db.sh
./scripts/init_redis.sh
```

3. Build and run via 'go':
```bash
go build ./api
go run ./api
```

4. (OPTIONAL) Build and run via 'go' with a production configuration:
```bash
go build ./api
go run ./api -cfg production
```

### Admin interface
Access the admin interface at http://127.0.0.1:8000/login

    Default account:
        Username: admin
        Password: gloriainvigilata

## Contributing

Contributions are welcome! If you'd like to contribute to this project, please follow these steps:

1. Fork the repository on GitHub.
2. Clone your forked repository to your local machine.
3. Create a new branch for your feature or bug fix.
4. Make your changes and commit them.
5. Push your changes to your fork on GitHub.
6. Open a pull request to the main repository.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
