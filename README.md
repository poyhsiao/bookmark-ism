# Bookmark Sync Service

A self-hosted multi-user bookmark synchronization service that provides cross-browser bookmark management with a visual interface similar to Toby. The system enables users to sync bookmarks across Chrome, Firefox, and Safari browsers while offering social features, intelligent organization, and community discovery.

## Features

- **Cross-browser sync**: Real-time bookmark synchronization across all major browsers
- **Visual interface**: Grid-based bookmark management with preview thumbnails
- **Social features**: Public collections, community discovery, and collaborative bookmarking
- **Intelligent organization**: AI-powered tagging, categorization, and duplicate detection
- **Self-hosted**: Complete data control with containerized deployment
- **Multi-language support**: Optimized for Chinese (Traditional/Simplified) and English

## Technology Stack

- **Backend**: Go with Gin web framework
- **Database**: Self-hosted Supabase PostgreSQL with GORM ORM
- **Cache**: Redis with Pub/Sub for real-time sync
- **Search**: Typesense with Chinese language support
- **Storage**: MinIO (primary storage for all files)
- **Authentication**: Self-hosted Supabase Auth with JWT
- **Real-time**: Self-hosted Supabase Realtime + WebSocket with Gorilla WebSocket library

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for convenience commands)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd bookmark-sync-service
   ```

2. **Initial setup (recommended)**
   ```bash
   make setup
   # This will:
   # - Check prerequisites
   # - Create .env file from template
   # - Download Go dependencies
   # - Build the application
   # - Start infrastructure services
   # - Create initial database migration
   ```

3. **Manual setup (alternative)**
   ```bash
   # Copy environment configuration
   cp .env.example .env
   # Edit .env with your configuration

   # Install dependencies
   make deps

   # Start all services
   make docker-up

   # Initialize storage buckets
   make init-buckets
   ```

4. **Run the application**
   ```bash
   make run
   # or for development with hot reload
   make dev
   ```

5. **Check service health**
   ```bash
   make health
   # or
   ./scripts/health-check.sh
   ```

The API server will start on `http://localhost:8080`

### Available Commands

```bash
make help          # Show all available commands
make setup         # Initial setup of development environment
make build         # Build the application
make run           # Run the application
make dev           # Start development environment with hot reload
make test          # Run tests
make docker-up     # Start all services with Docker Compose
make docker-down   # Stop all services
make docker-logs   # Show logs from all services
make init-buckets  # Initialize MinIO storage buckets
make health        # Check service health
make prod-up       # Start production environment
make prod-down     # Stop production environment
```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Authentication (Planned)
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh
- `POST /api/v1/auth/logout` - User logout

### Bookmarks (Planned)
- `GET /api/v1/bookmarks` - List user bookmarks
- `POST /api/v1/bookmarks` - Create bookmark
- `GET /api/v1/bookmarks/:id` - Get bookmark details
- `PUT /api/v1/bookmarks/:id` - Update bookmark
- `DELETE /api/v1/bookmarks/:id` - Delete bookmark

### Collections (Planned)
- `GET /api/v1/collections` - List collections
- `POST /api/v1/collections` - Create collection
- `GET /api/v1/collections/:id` - Get collection
- `PUT /api/v1/collections/:id` - Update collection
- `DELETE /api/v1/collections/:id` - Delete collection

### Synchronization (Planned)
- `GET /api/v1/sync/changes` - Get recent changes
- `POST /api/v1/sync/push` - Push local changes
- `GET /api/v1/sync/status` - Get sync status
- `WebSocket /api/v1/sync/ws` - Real-time sync

## Configuration

The application can be configured using environment variables or a YAML configuration file. See `.env.example` and `config/config.yaml` for available options.

### Key Configuration Options

- **Server**: Port, host, timeouts, environment
- **Database**: Supabase PostgreSQL connection settings
- **Redis**: Cache and pub/sub configuration
- **Supabase**: Auth, Realtime, and REST API URLs
- **Storage**: MinIO S3-compatible storage settings
- **Search**: Typesense search engine configuration
- **JWT**: Token secret and expiration settings
- **Logger**: Log level, format, and output configuration

## Development Status

This project is currently in active development. The foundation has been implemented with:

**✅ Phase 1: MVP Foundation**
- ✅ Complete Docker containerization with self-hosted Supabase stack
- ✅ Project structure and configuration management
- ✅ Database connection with GORM and Supabase PostgreSQL
- ✅ Redis client with Pub/Sub support
- ✅ Structured logging with Zap
- ✅ HTTP server with Gin framework
- ✅ Nginx load balancer with SSL support
- ✅ MinIO for file storage
- ✅ Typesense for search functionality
- ✅ Health monitoring and service discovery
- ✅ Development and production environments
- ✅ Automated setup and deployment scripts

**🚧 Next Steps (Phase 2)**
- 🚧 Database schema and migrations
- 🚧 Supabase authentication integration
- 🚧 User profile management
- 🚧 Core bookmark CRUD operations
- 🚧 Real-time synchronization
- 🚧 Browser extensions

## Contributing

This project follows a specification-driven development approach. Please refer to the `.kiro/specs/bookmark-sync-service/` directory for detailed requirements, design, and implementation plans.

## License

[License information to be added]