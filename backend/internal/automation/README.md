# Automation Service

The Automation Service provides advanced automation features for the bookmark synchronization system, including webhooks, RSS feed generation, bulk operations, backup management, API integrations, and automation rules.

## Features

### 🔗 Webhook System
- **External Service Integration**: Send real-time notifications to external services when events occur
- **Event Types**: Support for bookmark, collection, and user events
- **Delivery Management**: Automatic retry with exponential backoff
- **Signature Validation**: HMAC-SHA256 signature verification for security
- **Custom Headers**: Support for custom HTTP headers in webhook requests

### 📡 RSS Feed Generation
- **Public Collections**: Generate RSS feeds for public bookmark collections
- **Customizable Feeds**: Configure title, description, language, and other metadata
- **Automatic Updates**: Real-time feed updates when bookmarks are added/modified
- **Multi-language Support**: Support for different languages and character sets
- **Secure Access**: Public key-based access control

### 📦 Bulk Operations
- **Import/Export**: Bulk import and export of bookmarks and collections
- **Progress Tracking**: Real-time progress monitoring with percentage completion
- **Background Processing**: Asynchronous processing to avoid blocking operations
- **Error Handling**: Detailed error reporting and recovery mechanisms
- **Cancellation Support**: Ability to cancel running operations

### 💾 Backup Management
- **Automated Backups**: Scheduled full and incremental backups
- **Compression**: Support for gzip and zip compression
- **Encryption**: Optional backup encryption for sensitive data
- **Retention Policies**: Configurable backup retention periods
- **Integrity Verification**: Checksum validation for backup files

### 🔌 API Integrations
- **External Services**: Integration with services like Pocket, Instapaper, Raindrop
- **Sync Management**: Bidirectional synchronization with external services
- **Rate Limiting**: Built-in rate limiting to respect API quotas
- **Authentication**: Support for various authentication methods (API keys, OAuth)
- **Testing Tools**: Built-in tools to test integration connectivity

### 🤖 Automation Rules
- **Event-Driven**: Trigger actions based on bookmark and collection events
- **Flexible Conditions**: Support for complex conditional logic
- **Multiple Actions**: Execute multiple actions per rule
- **Priority System**: Rule execution based on priority levels
- **Execution Tracking**: Monitor rule execution history and performance

## API Endpoints

### Webhook Endpoints
```
POST   /api/v1/automation/webhooks           # Create webhook endpoint
GET    /api/v1/automation/webhooks           # List webhook endpoints
PUT    /api/v1/automation/webhooks/:id       # Update webhook endpoint
DELETE /api/v1/automation/webhooks/:id       # Delete webhook endpoint
GET    /api/v1/automation/webhooks/:id/deliveries # Get delivery history
```

### RSS Feed Endpoints
```
POST   /api/v1/automation/rss                # Create RSS feed
GET    /api/v1/automation/rss                # List RSS feeds
PUT    /api/v1/automation/rss/:id            # Update RSS feed
DELETE /api/v1/automation/rss/:id            # Delete RSS feed
GET    /api/v1/rss/:publicKey                # Access public RSS feed
```

### Bulk Operation Endpoints
```
POST   /api/v1/automation/bulk               # Create bulk operation
GET    /api/v1/automation/bulk               # List bulk operations
GET    /api/v1/automation/bulk/:id           # Get bulk operation status
DELETE /api/v1/automation/bulk/:id           # Cancel bulk operation
```

### Backup Endpoints
```
POST   /api/v1/automation/backup             # Create backup job
GET    /api/v1/automation/backup             # List backup jobs
GET    /api/v1/automation/backup/:id         # Get backup job status
GET    /api/v1/automation/backup/:id/download # Download backup file
```

### API Integration Endpoints
```
POST   /api/v1/automation/integrations       # Create API integration
GET    /api/v1/automation/integrations       # List API integrations
PUT    /api/v1/automation/integrations/:id   # Update API integration
DELETE /api/v1/automation/integrations/:id   # Delete API integration
POST   /api/v1/automation/integrations/:id/sync # Trigger manual sync
POST   /api/v1/automation/integrations/:id/test # Test integration
```

### Automation Rule Endpoints
```
POST   /api/v1/automation/rules              # Create automation rule
GET    /api/v1/automation/rules              # List automation rules
PUT    /api/v1/automation/rules/:id          # Update automation rule
DELETE /api/v1/automation/rules/:id          # Delete automation rule
POST   /api/v1/automation/rules/:id/execute  # Execute automation rule
```

## Data Models

### WebhookEndpoint
```go
type WebhookEndpoint struct {
    ID          uint              `json:"id"`
    UserID      string            `json:"user_id"`
    Name        string            `json:"name"`
    URL         string            `json:"url"`
    Secret      string            `json:"-"` // Hidden from JSON
    Events      []string          `json:"events"`
    Active      bool              `json:"active"`
    RetryCount  int               `json:"retry_count"`
    Timeout     int               `json:"timeout"`
    Headers     map[string]string `json:"headers"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}
```

### RSSFeed
```go
type RSSFeed struct {
    ID          uint      `json:"id"`
    UserID      string    `json:"user_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Link        string    `json:"link"`
    Language    string    `json:"language"`
    PublicKey   string    `json:"public_key"`
    Collections []uint    `json:"collections"`
    Tags        []string  `json:"tags"`
    Active      bool      `json:"active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### BulkOperation
```go
type BulkOperation struct {
    ID             uint                   `json:"id"`
    UserID         string                 `json:"user_id"`
    Type           string                 `json:"type"`
    Status         string                 `json:"status"`
    Progress       int                    `json:"progress"`
    TotalItems     int                    `json:"total_items"`
    ProcessedItems int                    `json:"processed_items"`
    FailedItems    int                    `json:"failed_items"`
    Parameters     map[string]interface{} `json:"parameters"`
    Result         map[string]interface{} `json:"result"`
    CreatedAt      time.Time              `json:"created_at"`
    UpdatedAt      time.Time              `json:"updated_at"`
}
```

## Usage Examples

### Creating a Webhook Endpoint

```bash
curl -X POST http://localhost:8080/api/v1/automation/webhooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Slack Notifications",
    "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
    "events": ["bookmark.created", "bookmark.updated"],
    "active": true,
    "retry_count": 3,
    "timeout": 30,
    "headers": {
      "Content-Type": "application/json"
    }
  }'
```

### Creating an RSS Feed

```bash
curl -X POST http://localhost:8080/api/v1/automation/rss \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "My Tech Bookmarks",
    "description": "Latest bookmarks from my tech collection",
    "link": "https://mybookmarks.example.com",
    "language": "en",
    "collections": [1, 2, 3],
    "tags": ["tech", "programming"]
  }'
```

### Starting a Bulk Import

```bash
curl -X POST http://localhost:8080/api/v1/automation/bulk \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "type": "import",
    "parameters": {
      "source": "chrome",
      "file_path": "/uploads/bookmarks.html",
      "preserve_folders": true
    }
  }'
```

### Creating a Backup Job

```bash
curl -X POST http://localhost:8080/api/v1/automation/backup \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "type": "full",
    "compression": "gzip",
    "encrypted": true
  }'
```

### Setting up an API Integration

```bash
curl -X POST http://localhost:8080/api/v1/automation/integrations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Pocket Integration",
    "type": "pocket",
    "base_url": "https://getpocket.com/v3",
    "api_key": "YOUR_POCKET_API_KEY",
    "sync_enabled": true,
    "sync_interval": 3600
  }'
```

### Creating an Automation Rule

```bash
curl -X POST http://localhost:8080/api/v1/automation/rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Auto-tag GitHub Bookmarks",
    "description": "Automatically tag bookmarks from GitHub",
    "trigger": "bookmark_added",
    "conditions": {
      "url_contains": ["github.com"]
    },
    "actions": {
      "add_tags": ["github", "code", "development"]
    },
    "priority": 10
  }'
```

## Event Types

### Webhook Events
- `bookmark.created` - New bookmark added
- `bookmark.updated` - Bookmark modified
- `bookmark.deleted` - Bookmark removed
- `collection.created` - New collection created
- `collection.updated` - Collection modified
- `collection.deleted` - Collection removed
- `user.registered` - New user registered
- `user.updated` - User profile updated

### Automation Rule Triggers
- `bookmark_added` - When a bookmark is added
- `bookmark_updated` - When a bookmark is modified
- `bookmark_deleted` - When a bookmark is removed
- `collection_created` - When a collection is created
- `tag_added` - When a tag is added to a bookmark
- `scheduled` - Time-based triggers

## Security Features

### Webhook Security
- **HMAC Signature**: All webhook payloads are signed with HMAC-SHA256
- **Secret Management**: Webhook secrets are securely stored and never exposed
- **Retry Logic**: Failed deliveries are retried with exponential backoff
- **Timeout Protection**: Configurable timeouts prevent hanging requests

### API Integration Security
- **Credential Encryption**: API keys and secrets are encrypted at rest
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **Token Refresh**: Automatic token refresh for OAuth integrations
- **Secure Storage**: Sensitive data is never logged or exposed

### Access Control
- **User Isolation**: All resources are isolated by user ID
- **Authentication Required**: All endpoints require valid authentication
- **Authorization Checks**: Users can only access their own resources
- **Input Validation**: All inputs are validated and sanitized

## Performance Considerations

### Asynchronous Processing
- Webhook deliveries are processed asynchronously
- Bulk operations run in background goroutines
- Backup jobs are queued and processed separately
- RSS feed generation is cached for performance

### Resource Management
- Connection pooling for HTTP clients
- Database connection optimization
- Memory-efficient processing for large datasets
- Configurable timeouts and limits

### Monitoring and Observability
- Comprehensive logging for all operations
- Metrics collection for performance monitoring
- Error tracking and alerting
- Health check endpoints

## Testing

The automation service includes comprehensive tests:

### Unit Tests
```bash
go test ./internal/automation -v
```

### Integration Tests
```bash
go test ./internal/automation -tags=integration -v
```

### Test Coverage
```bash
go test -coverprofile=coverage.out ./internal/automation
go tool cover -html=coverage.out
```

### Load Testing
```bash
./scripts/test-automation.sh
```

## Configuration

### Environment Variables
```bash
# Webhook settings
WEBHOOK_TIMEOUT=30
WEBHOOK_RETRY_COUNT=3
WEBHOOK_MAX_CONCURRENT=10

# RSS settings
RSS_CACHE_TTL=300
RSS_MAX_ITEMS=100

# Backup settings
BACKUP_RETENTION_DAYS=30
BACKUP_COMPRESSION=gzip

# API integration settings
API_RATE_LIMIT=100
API_TIMEOUT=30
```

### Database Configuration
The service requires the following database tables:
- `webhook_endpoints`
- `webhook_deliveries`
- `rss_feeds`
- `bulk_operations`
- `backup_jobs`
- `api_integrations`
- `automation_rules`

## Deployment

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o automation-service ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/automation-service .
CMD ["./automation-service"]
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: automation-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: automation-service
  template:
    metadata:
      labels:
        app: automation-service
    spec:
      containers:
      - name: automation-service
        image: automation-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
```

## Monitoring and Alerting

### Health Checks
```bash
curl http://localhost:8080/health
```

### Metrics Endpoints
```bash
curl http://localhost:8080/metrics
```

### Log Aggregation
Logs are structured in JSON format for easy parsing:
```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "service": "automation",
  "component": "webhook",
  "message": "Webhook delivered successfully",
  "webhook_id": 123,
  "delivery_id": 456,
  "status_code": 200,
  "duration_ms": 150
}
```

## Troubleshooting

### Common Issues

#### Webhook Delivery Failures
- Check webhook URL accessibility
- Verify webhook signature validation
- Review timeout settings
- Check retry configuration

#### RSS Feed Generation Issues
- Verify collection permissions
- Check feed configuration
- Review bookmark data integrity
- Validate XML generation

#### Bulk Operation Failures
- Check available disk space
- Verify file permissions
- Review operation parameters
- Monitor memory usage

#### API Integration Problems
- Validate API credentials
- Check rate limiting
- Verify network connectivity
- Review authentication tokens

### Debug Mode
Enable debug logging:
```bash
export LOG_LEVEL=debug
export GIN_MODE=debug
```

### Performance Profiling
```bash
go tool pprof http://localhost:8080/debug/pprof/profile
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Code Style
- Follow Go conventions
- Use meaningful variable names
- Add comprehensive comments
- Include error handling
- Write unit tests

### Testing Requirements
- Unit test coverage > 90%
- Integration tests for all endpoints
- Performance tests for critical paths
- Security tests for authentication

## License

This automation service is part of the bookmark synchronization system and is licensed under the same terms as the main project.