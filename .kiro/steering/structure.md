# Project Structure

## Repository Organization

```
bookmark-sync-service/
├── backend/                    # Go backend services
│   ├── cmd/                   # Application entry points
│   │   ├── api/              # API server
│   │   ├── sync/             # Sync service
│   │   ├── worker/           # Background workers
│   │   └── migrate/          # Database migrations
│   ├── internal/             # Private application code
│   │   ├── auth/             # Authentication logic
│   │   ├── bookmark/         # Bookmark business logic
│   │   ├── sync/             # Synchronization logic
│   │   ├── community/        # Social features
│   │   ├── search/           # Search integration
│   │   └── storage/          # File storage logic
│   ├── pkg/                  # Public packages
│   │   ├── database/         # Database models and connections
│   │   ├── redis/            # Redis client operations
│   │   ├── websocket/        # WebSocket management
│   │   └── utils/            # Shared utilities
│   ├── api/                  # API definitions
│   │   └── v1/               # API v1 routes
│   ├── config/               # Configuration files
│   ├── migrations/           # Database schema migrations
│   └── docker/               # Docker configurations
├── extensions/               # Browser extensions
│   ├── chrome/              # Chrome extension
│   ├── firefox/             # Firefox extension
│   ├── safari/              # Safari extension
│   └── shared/              # Shared extension code
│       ├── api-client.js    # API communication
│       ├── sync-manager.js  # Sync logic
│       └── ui-components.js # Shared UI components
├── web/                     # Web application
│   ├── src/                 # Source code
│   ├── public/              # Static assets
│   └── dist/                # Built files
├── docs/                    # Documentation
│   ├── api/                 # API documentation
│   ├── deployment/          # Deployment guides
│   └── user/                # User guides
├── scripts/                 # Utility scripts
│   ├── setup.sh            # Development setup
│   ├── deploy.sh           # Deployment script
│   └── backup.sh           # Backup utilities
├── docker-compose.yml       # Development environment
├── docker-compose.prod.yml  # Production environment
└── .kiro/                   # Kiro configuration
    ├── specs/               # Feature specifications
    └── steering/            # AI assistant guidance
```

## Code Organization Principles

### Backend Structure
- **cmd/**: Each service has its own main.go entry point
- **internal/**: Business logic organized by domain (auth, bookmark, sync)
- **pkg/**: Reusable packages that could be imported by other projects
- **api/**: HTTP route definitions and handlers

### Extension Structure
- **shared/**: Common code used across all browser extensions
- **browser-specific/**: Platform-specific implementations and manifests
- **background/**: Service workers and background scripts
- **popup/**: Extension popup UI and logic

### Configuration Management
- Environment-specific configs in `config/` directory
- Docker Compose files for different deployment scenarios
- Secrets managed through environment variables
- Database migrations versioned and tracked

### Testing Organization
- Unit tests alongside source code (`*_test.go`)
- Integration tests in `tests/integration/`
- End-to-end tests in `tests/e2e/`
- Test fixtures and utilities in `tests/fixtures/`

## Naming Conventions

- **Go packages**: lowercase, single word when possible
- **Go files**: snake_case for multi-word files
- **Database tables**: snake_case with plural nouns
- **API endpoints**: kebab-case in URLs
- **Environment variables**: UPPER_SNAKE_CASE
- **Docker services**: kebab-case in compose files