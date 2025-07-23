# Technology Stack

## Backend Architecture

- **Language**: Go with Gin web framework
- **Database**: Self-hosted Supabase PostgreSQL with GORM ORM
- **Cache**: Redis with Pub/Sub for real-time sync
- **Search**: Typesense with Chinese language support
- **Storage**: MinIO (primary storage for all files)
- **Authentication**: Self-hosted Supabase Auth with JWT
- **Real-time**: Self-hosted Supabase Realtime + WebSocket with Gorilla WebSocket library

## Frontend

- **Browser Extensions**: Chrome, Firefox, Safari (WebExtensions API)
- **Web Interface**: Responsive web app with grid-based UI
- **Mobile**: Progressive Web App (PWA)

## Infrastructure

- **Containerization**: Docker + Docker Compose
- **Load Balancer**: Nginx
- **Monitoring**: Prometheus + Grafana
- **Deployment**: Self-hosted with horizontal scaling support

## Key Libraries

```go
// Core dependencies
"github.com/gin-gonic/gin"           // Web framework
"gorm.io/gorm"                       // ORM
"github.com/go-redis/redis/v8"       // Redis client
"github.com/gorilla/websocket"       // WebSocket
"github.com/typesense/typesense-go"  // Search client
"github.com/golang-jwt/jwt/v4"       // JWT handling

// Supabase Integration
"github.com/supabase-community/supabase-go"    // Supabase client
"github.com/supabase-community/postgrest-go"   // PostgREST client
```

## Common Commands

### Development
```bash
# Start all services
docker-compose up -d

# Run backend with hot reload
go run cmd/api/main.go

# Run database migrations
go run cmd/migrate/main.go up

# Run tests
go test ./...
```

### Production
```bash
# Deploy production stack
docker-compose -f docker-compose.prod.yml up -d

# Scale API services
docker-compose up --scale api=3

# Backup Supabase database
docker-compose exec supabase-db pg_dump -U postgres postgres > backup.sql
```

### Search Index Management
```bash
# Rebuild search index
curl -X POST "http://localhost:8108/collections/bookmarks/documents/import"

# Check search health
curl "http://localhost:8108/health"
```

### Supabase Management
```bash
# Check Supabase services health
curl "http://localhost:9999/health"  # Auth service
curl "http://localhost:3000/"        # REST API
curl "http://localhost:4000/api/health" # Realtime service

# Access Supabase database directly
docker-compose exec supabase-db psql -U postgres -d postgres

# View Supabase logs
docker-compose logs supabase-auth
docker-compose logs supabase-rest
docker-compose logs supabase-realtime
```