# Matching Engine

A high-performance order matching engine implemented in Go, designed for handling cryptocurrency trading pairs with a focus on reliability and efficiency.

## Features

- Fast order matching algorithm
- Thread-safe order book management
- Support for limit orders (buy/sell)
- Real-time order book updates
- Clean architecture design
- Comprehensive test coverage

## Prerequisites

- Go 1.22 or higher
- Make (optional, for using Makefile commands)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/matchengine.git

# Navigate to project directory
cd matchengine

# Install dependencies
go mod download
```

## Project Structure

```
.
├── cmd/
│   ├── api/          # API entry point
│   └── test/         # Test utilities
├── internal/
│   ├── domain/       # Domain models and business logic
│   │   ├── order/    # Order related entities
│   │   └── orderbook/# Order book implementation
│   ├── handler/      # HTTP handlers
│   ├── middleware/   # HTTP middleware
│   └── service/      # Business services
├── pkg/              # Shared packages
└── scripts/          # Build and deployment scripts
```

## Usage

### Running the Service

```bash
# Run the API server
go run cmd/api/main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## API Documentation

### Order Management

```
POST /api/v1/orders
GET /api/v1/orders/{id}
DELETE /api/v1/orders/{id}
```

### Order Book

```
GET /api/v1/orderbook/{symbol}
GET /api/v1/orderbook/{symbol}/best
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors who have helped shape this project
- Inspired by real-world cryptocurrency exchanges 