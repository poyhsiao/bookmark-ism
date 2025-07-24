# Design Document

## Overview

The bookmark synchronization service is designed as a distributed system consisting of a self-hosted backend API service and cross-platform browser extensions. The architecture follows a client-server model with real-time synchronization, social features, and comprehensive bookmark management capabilities.

## Implementation Status Summary

### ✅ Completed Components
- **Core Backend Infrastructure**: Go + Gin framework with modular architecture
- **Database Layer**: Supabase PostgreSQL with GORM ORM and comprehensive schema
- **Authentication System**: Supabase Auth integration with JWT validation
- **Bookmark Management**: Full CRUD operations with advanced features
- **User Management**: Profile management and preferences storage
- **Testing Framework**: Comprehensive TDD approach with unit and integration tests

### ⏳ In Progress / Planned
- **Collection Management**: Basic folder support and organization (Task 7)
- **Real-time Synchronization**: WebSocket-based sync with conflict resolution (Task 8-9)
- **Browser Extensions**: Chrome, Firefox, Safari extensions (Task 10-11)
- **Visual Interface**: Grid layout with screenshots and previews (Task 12-13)
- **Advanced Search**: Typesense integration with multi-language support (Task 14)

### Key Design Principles

- **Privacy-First**: Self-hosted deployment ensures complete data control
- **Cross-Platform Compatibility**: Unified experience across Chrome, Firefox, and Safari
- **Real-Time Sync**: Sub-30-second synchronization across all connected devices
- **Scalable Architecture**: Microservices design supporting horizontal scaling
- **Offline-First**: Local-first approach with intelligent conflict resolution
- **Social Integration**: Community features with granular privacy controls

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "Frontend Applications"
        CE[Chrome Extension]
        FE[Firefox Extension]
        SE[Safari Extension]
        WA[Responsive Web App]
        MA[Mobile PWA]
    end

    subgraph "API Gateway"
        AG[Nginx Load Balancer]
    end

    subgraph "Backend Services (Go)"
        AS[Auth Service - Go + Gin]
        BS[Bookmark Service - Go + Gin]
        SS[Sync Service - Go + WebSocket]
        CS[Community Service - Go + Gin]
        NS[Notification Service - Go + Gin]
        FS[File Service - Go + Gin]
    end

    subgraph "Self-hosted Supabase Stack"
        SDB[(Supabase PostgreSQL)]
        SA[Supabase Auth]
        SRT[Supabase Realtime]
        SAPI[Supabase REST API]
    end

    subgraph "Additional Data Layer"
        RDB[(Redis Cache + Pub/Sub)]
        MS[(MinIO Storage)]
        TS[(Typesense Search)]
    end

    subgraph "External Services"
        WH[Webhooks]
        RSS[RSS Feeds]
        AI[AI/ML Services]
    end

    CE --> AG
    FE --> AG
    SE --> AG

    AG --> AS
    AG --> BS
    AG --> SS
    AG --> CS
    AG --> NS
    AG --> FS

    AS --> SA
    AS --> SDB
    BS --> SDB
    SS --> RDB
    CS --> SDB
    NS --> RDB
    FS --> RF

    BS --> TS
    CS --> TS

    SS --> WH
    BS --> RSS
    CS --> AI
```

### Service Architecture

#### 1. API Gateway Layer (Nginx)

- **Load Balancing**: Distributes requests across Go service instances using upstream blocks
- **Rate Limiting**: Nginx rate limiting with Redis backend for distributed rate limiting
- **SSL Termination**: TLS 1.3 encryption and certificate management with Let's Encrypt integration
- **Request Routing**: Location-based routing to Go microservices with path rewriting
- **CORS Handling**: Configurable CORS policies for browser extensions
- **Health Checks**: Nginx upstream health checks with automatic failover
- **Static Content**: Efficient serving of static assets and cached responses
- **Compression**: Gzip/Brotli compression for improved performance

#### 2. Core Services

**Authentication Service (Go + Supabase Auth)**

- Hybrid authentication: Supabase Auth + custom Go validation
- JWT token generation, validation, and refresh logic
- Password reset and account recovery workflows
- OAuth integration (Google, GitHub, etc.) via Supabase
- Role-based access control (RBAC) with custom Go middleware
- Session management with Redis storage

**Bookmark Service**

- CRUD operations for bookmarks and collections
- Metadata extraction and content analysis
- Tag management and auto-suggestion
- Import/export functionality
- Duplicate detection and merging

**Synchronization Service (Go + Gorilla WebSocket)**

- Real-time sync using Gorilla WebSocket library
- Redis Pub/Sub for multi-instance message broadcasting
- Conflict resolution with vector clocks and timestamp priority
- Change event propagation with efficient delta sync
- Offline queue management with persistent Redis storage
- Connection pooling and automatic reconnection handling

**Community Service**

- Public bookmark discovery and trending
- Social metrics and popularity tracking
- User following and feed generation
- Comment and discussion management
- Content moderation and reporting

**Notification Service**

- Real-time notifications via WebSockets
- Email notifications for important events
- Push notifications for mobile (future)
- Notification preferences and filtering
- Webhook delivery management

**File Service**

- Screenshot capture and storage
- Image optimization and thumbnail generation
- File upload and download management
- CDN integration for global delivery
- Backup and archival storage

#### 3. Data Layer

**Primary Database (Supabase PostgreSQL)**

- User accounts and authentication data managed through Supabase Auth
- Bookmark metadata and relationships with Supabase Row Level Security (RLS)
- Collections and folder structures with real-time subscriptions
- Social interactions and comments with Supabase Realtime updates
- System configuration and settings with Supabase REST API access

**Cache Layer (Redis)**

- Session storage and JWT blacklisting
- Frequently accessed bookmark data
- Real-time sync state management
- Rate limiting counters
- Temporary data for processing

**Search Engine (Typesense)**

- Full-text search with millisecond response times
- Multi-language support with Chinese (Traditional/Simplified) tokenization
- Real-time indexing and instant search results
- Typo tolerance and auto-complete features
- Faceted search and filtering capabilities
- Semantic search with vector embeddings support

**File Storage (MinIO S3-Compatible Storage)**

- High-performance object storage for webpage screenshots and system files implemented using MinIO
- S3-compatible API for easy integration with Go backend
- Optimized for concurrent read/write operations with proven MinIO performance
- Built-in image optimization and thumbnail generation capabilities
- Backup archives and data export storage
- Self-hosted solution for complete data control using MinIO as the storage engine

## Search Engine Configuration

### Typesense Multi-Language Setup

#### Chinese Language Support

```json
{
  "name": "bookmarks",
  "fields": [
    {
      "name": "title",
      "type": "string",
      "locale": "zh"
    },
    {
      "name": "description",
      "type": "string",
      "locale": "zh"
    },
    {
      "name": "content",
      "type": "string",
      "locale": "zh"
    },
    {
      "name": "tags",
      "type": "string[]",
      "locale": "zh"
    }
  ],
  "token_separators": ["，", "。", "！", "？", "；", "："],
  "symbols_to_index": ["@", "#", "$"],
  "default_sorting_field": "popularity_score"
}
```

#### Multi-Language Tokenization

- **Chinese (Simplified/Traditional)**: Built-in CJK tokenizer with word segmentation
- **English**: Standard tokenization with stemming
- **Japanese**: Kuromoji tokenizer support
- **Korean**: Nori tokenizer integration
- **European Languages**: ICU tokenizer with locale-specific rules

#### Search Features Configuration

```json
{
  "search_parameters": {
    "query_by": "title,description,content,tags",
    "query_by_weights": "4,3,2,1",
    "typo_tokens_threshold": 1,
    "drop_tokens_threshold": 1,
    "pinned_hits": [],
    "hidden_hits": [],
    "enable_overrides": true,
    "highlight_fields": "title,description",
    "snippet_threshold": 30,
    "num_typos": "2,1,0",
    "min_len_1typo": 4,
    "min_len_2typo": 7
  }
}
```

### Alternative Search Engine Comparison

#### If Using OpenSearch (Advanced Option)

```yaml
# OpenSearch Chinese Configuration
analysis:
  analyzer:
    ik_max_word:
      type: custom
      tokenizer: ik_max_word
      filter: [lowercase, stop]
    ik_smart:
      type: custom
      tokenizer: ik_smart
      filter: [lowercase, stop]
  tokenizer:
    ik_max_word:
      type: ik_max_word
    ik_smart:
      type: ik_smart
```

**Pros**: Most powerful Chinese text analysis, custom dictionaries
**Cons**: Higher resource usage, complex configuration

#### If Using Meilisearch (Simple Option)

```json
{
  "searchableAttributes": ["title", "description", "content", "tags"],
  "stopWords": {
    "zh": ["的", "了", "在", "是", "我", "有", "和", "就", "不", "人"]
  },
  "synonyms": {
    "書籤": ["书签", "bookmark"],
    "收藏": ["收藏夹", "favorites"]
  }
}
```

**Pros**: Simple setup, good performance
**Cons**: Limited Chinese tokenization, fewer advanced features

## Components and Interfaces

### Go Backend Implementation Details

#### Technology Stack

- **Framework**: Gin Web Framework for high-performance HTTP APIs
- **Database**: GORM for Supabase PostgreSQL ORM with connection pooling
- **Cache**: go-redis client for Redis operations and Pub/Sub
- **WebSocket**: Gorilla WebSocket + Supabase Realtime for real-time communication
- **Storage**: MinIO client for S3-compatible storage operations
- **Search**: Typesense Go client for search operations
- **Authentication**: Supabase Go client + custom JWT middleware for hybrid auth

#### Service Structure

```
backend/
├── cmd/
│   ├── api/              # API server entry point
│   ├── sync/             # Sync service entry point
│   └── worker/           # Background worker entry point
├── internal/
│   ├── auth/             # Authentication logic
│   ├── bookmark/         # Bookmark business logic
│   ├── sync/             # Synchronization logic
│   ├── community/        # Community features
│   ├── search/           # Search integration
│   └── storage/          # File storage logic
├── pkg/
│   ├── database/         # Database connection and models
│   ├── redis/            # Redis client and operations
│   ├── websocket/        # WebSocket connection management
│   └── utils/            # Shared utilities
├── api/
│   └── v1/               # API route definitions
├── config/               # Configuration management
└── migrations/           # Database migrations
```

#### Key Go Libraries

```go
// Core dependencies
"github.com/gin-gonic/gin"           // Web framework
"gorm.io/gorm"                       // ORM
"github.com/go-redis/redis/v8"       // Redis client
"github.com/gorilla/websocket"       // WebSocket
"github.com/typesense/typesense-go"  // Search client

// Authentication & Security
"github.com/golang-jwt/jwt/v4"       // JWT handling
"golang.org/x/crypto/bcrypt"         // Password hashing
"github.com/supabase-community/supabase-go" // Supabase client
"github.com/supabase-community/postgrest-go" // PostgREST client

// Storage & File handling
"github.com/minio/minio-go/v7"       // S3-compatible client for MinIO
"github.com/disintegration/imaging"  // Image processing

// Utilities
"github.com/spf13/viper"            // Configuration
"go.uber.org/zap"                   // Structured logging
"github.com/google/uuid"            // UUID generation
```

### Browser Extension Architecture

#### Extension Structure

```
extension/
├── manifest.json          # Extension configuration
├── background/            # Service worker/background scripts
│   ├── sync-manager.js   # Handles sync operations
│   ├── auth-manager.js   # Authentication handling
│   └── storage-manager.js # Local storage management
├── content/              # Content scripts
│   ├── bookmark-detector.js # Detects bookmarkable content
│   └── page-analyzer.js     # Extracts page metadata
├── popup/                # Extension popup UI
│   ├── popup.html       # Main popup interface
│   ├── popup.js         # Popup logic
│   └── popup.css        # Popup styling
├── options/              # Settings page
│   ├── options.html     # Settings interface
│   ├── options.js       # Settings logic
│   └── options.css      # Settings styling
└── shared/               # Shared utilities
    ├── api-client.js    # API communication
    ├── utils.js         # Common utilities
    └── constants.js     # Configuration constants
```

#### Key Extension Components

**Sync Manager**

- Manages bidirectional synchronization with backend
- Handles offline queue and conflict resolution
- Implements exponential backoff for failed requests
- Maintains local cache consistency

**Storage Manager**

- Abstracts browser-specific storage APIs
- Implements data encryption for sensitive information
- Manages cache eviction and cleanup
- Provides offline data access

**UI Components**

- Grid-based bookmark visualization
- Drag-and-drop organization interface
- Search and filtering capabilities
- Social features integration

### API Design

#### RESTful API Endpoints

**Authentication Endpoints**

```
POST /api/v1/auth/register     # User registration
POST /api/v1/auth/login        # User login
POST /api/v1/auth/refresh      # Token refresh
POST /api/v1/auth/logout       # User logout
POST /api/v1/auth/reset        # Password reset
```

**Bookmark Management** ✅ IMPLEMENTED

```
GET    /api/v1/bookmarks           # List user bookmarks with search, filtering, and pagination
POST   /api/v1/bookmarks           # Create bookmark with URL validation and metadata
GET    /api/v1/bookmarks/{id}      # Get bookmark details with user authorization
PUT    /api/v1/bookmarks/{id}      # Update bookmark with validation and conflict handling
DELETE /api/v1/bookmarks/{id}      # Soft delete bookmark with recovery capability
POST   /api/v1/bookmarks/batch     # Batch operations (planned for future implementation)
```

**Implemented Features:**
- ✅ Full CRUD operations with proper HTTP status codes
- ✅ User authorization and data isolation
- ✅ URL format validation and error handling
- ✅ JSON-based tag storage and management
- ✅ Search across title, description, and URL
- ✅ Filtering by tags, status, and collections
- ✅ Pagination with limit/offset and total count
- ✅ Sorting by created_at, updated_at, title, url
- ✅ Soft delete with recovery capability
- ✅ Comprehensive error responses with detailed messages

**Collection Management**

```
GET    /api/v1/collections         # List collections
POST   /api/v1/collections         # Create collection
GET    /api/v1/collections/{id}    # Get collection
PUT    /api/v1/collections/{id}    # Update collection
DELETE /api/v1/collections/{id}    # Delete collection
```

**Synchronization**

```
GET    /api/v1/sync/changes        # Get recent changes
POST   /api/v1/sync/push           # Push local changes
GET    /api/v1/sync/status         # Get sync status
WebSocket /api/v1/sync/ws          # Real-time sync
```

**Community Features**

```
GET    /api/v1/community/trending  # Get trending bookmarks
GET    /api/v1/community/popular   # Get popular collections
POST   /api/v1/community/follow    # Follow user
GET    /api/v1/community/feed      # Get user feed
POST   /api/v1/community/comment   # Add comment
```

**Search and Discovery**

```
GET    /api/v1/search/bookmarks    # Search bookmarks
GET    /api/v1/search/collections  # Search collections
GET    /api/v1/search/users        # Search users
GET    /api/v1/search/suggest      # Get suggestions
```

#### WebSocket Events

**Real-time Synchronization**

```javascript
// Client to Server
{
  "type": "bookmark_created",
  "data": { "bookmark": {...}, "timestamp": "..." }
}

{
  "type": "bookmark_updated",
  "data": { "id": "...", "changes": {...}, "timestamp": "..." }
}

{
  "type": "bookmark_deleted",
  "data": { "id": "...", "timestamp": "..." }
}

// Server to Client
{
  "type": "sync_update",
  "data": { "changes": [...], "last_sync": "..." }
}

{
  "type": "conflict_detected",
  "data": { "bookmark_id": "...", "conflicts": [...] }
}
```

## Data Models

### Core Data Models

#### User Model

```typescript
interface User {
  id: string;
  email: string;
  username: string;
  displayName: string;
  avatar?: string;
  preferences: UserPreferences;
  quotas: UserQuotas;
  createdAt: Date;
  updatedAt: Date;
  lastActiveAt: Date;
}

interface UserPreferences {
  theme: "light" | "dark" | "auto";
  gridSize: "small" | "medium" | "large";
  defaultView: "grid" | "list";
  notifications: NotificationSettings;
  privacy: PrivacySettings;
}

interface UserQuotas {
  maxBookmarks: number;
  maxCollections: number;
  storageLimit: number; // in bytes
  apiRateLimit: number; // requests per minute
}
```

#### Bookmark Model

```typescript
interface Bookmark {
  id: string;
  userId: string;
  url: string;
  title: string;
  description?: string;
  favicon?: string;
  screenshot?: string;
  tags: string[];
  collections: string[];
  metadata: BookmarkMetadata;
  socialMetrics: SocialMetrics;
  createdAt: Date;
  updatedAt: Date;
  lastAccessedAt?: Date;
}

interface BookmarkMetadata {
  extractedTitle?: string;
  extractedDescription?: string;
  keywords: string[];
  contentType?: string;
  language?: string;
  readingTime?: number;
  wordCount?: number;
  lastChecked?: Date;
  status: "active" | "broken" | "redirected" | "changed";
}

interface SocialMetrics {
  saveCount: number;
  likeCount: number;
  commentCount: number;
  shareCount: number;
  trendingScore: number;
  lastTrendingUpdate: Date;
}
```

#### Collection Model

```typescript
interface Collection {
  id: string;
  userId: string;
  name: string;
  description?: string;
  color?: string;
  icon?: string;
  parentId?: string; // for nested collections
  visibility: "private" | "public" | "shared";
  shareSettings?: ShareSettings;
  bookmarkIds: string[];
  metadata: CollectionMetadata;
  createdAt: Date;
  updatedAt: Date;
}

interface ShareSettings {
  allowComments: boolean;
  allowCopying: boolean;
  sharedWith: string[]; // user IDs
  shareLink?: string;
  shareExpiry?: Date;
}

interface CollectionMetadata {
  bookmarkCount: number;
  lastBookmarkAdded?: Date;
  tags: string[]; // aggregated from bookmarks
  categories: string[];
  averageRating?: number;
}
```

### Synchronization Models

#### Sync Event Model

```typescript
interface SyncEvent {
  id: string;
  userId: string;
  deviceId: string;
  eventType: "create" | "update" | "delete";
  resourceType: "bookmark" | "collection" | "tag";
  resourceId: string;
  changes: Record<string, any>;
  timestamp: Date;
  processed: boolean;
  conflictResolved?: boolean;
}

interface SyncState {
  userId: string;
  deviceId: string;
  lastSyncTimestamp: Date;
  pendingEvents: number;
  conflictCount: number;
  status: "synced" | "pending" | "conflict" | "error";
}
```

### Social Models

#### Comment Model

```typescript
interface Comment {
  id: string;
  bookmarkId: string;
  userId: string;
  content: string;
  parentId?: string; // for threaded comments
  reactions: Reaction[];
  createdAt: Date;
  updatedAt: Date;
  isModerated: boolean;
}

interface Reaction {
  userId: string;
  type: "like" | "dislike" | "helpful" | "bookmark";
  createdAt: Date;
}
```

#### Follow Model

```typescript
interface Follow {
  id: string;
  followerId: string;
  followingId: string;
  createdAt: Date;
  notificationSettings: {
    newCollections: boolean;
    updatedCollections: boolean;
    newBookmarks: boolean;
  };
}
```

## Error Handling

### Error Response Format

```typescript
interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, any>;
    timestamp: Date;
    requestId: string;
  };
}
```

### Error Categories

**Authentication Errors (4xx)**

- `AUTH_REQUIRED`: Authentication required
- `AUTH_INVALID`: Invalid credentials
- `AUTH_EXPIRED`: Token expired
- `AUTH_FORBIDDEN`: Insufficient permissions

**Validation Errors (4xx)**

- `VALIDATION_FAILED`: Request validation failed
- `RESOURCE_NOT_FOUND`: Resource not found
- `DUPLICATE_RESOURCE`: Resource already exists
- `QUOTA_EXCEEDED`: User quota exceeded

**Sync Errors (4xx/5xx)**

- `SYNC_CONFLICT`: Synchronization conflict
- `SYNC_FAILED`: Synchronization failed
- `OFFLINE_MODE`: Operating in offline mode
- `NETWORK_ERROR`: Network connectivity issue

**System Errors (5xx)**

- `INTERNAL_ERROR`: Internal server error
- `SERVICE_UNAVAILABLE`: Service temporarily unavailable
- `DATABASE_ERROR`: Database operation failed
- `EXTERNAL_SERVICE_ERROR`: External service error

### Error Recovery Strategies

**Client-Side Recovery**

- Exponential backoff for failed requests
- Automatic retry with jitter
- Graceful degradation to offline mode
- User notification with recovery options

**Server-Side Recovery**

- Circuit breaker pattern for external services
- Database connection pooling and failover
- Graceful service degradation
- Comprehensive logging and monitoring

## Testing Strategy

### Unit Testing

- **Backend Services**: Jest/Mocha for API endpoints and business logic
- **Browser Extensions**: Jest with browser API mocking
- **Database Layer**: Integration tests with test database
- **Utilities**: Pure function testing with comprehensive coverage

### Integration Testing

- **API Integration**: End-to-end API workflow testing
- **Database Integration**: Multi-service transaction testing
- **External Service Integration**: Mock external APIs
- **Cross-Browser Testing**: Automated extension testing

### End-to-End Testing

- **User Workflows**: Selenium/Playwright for complete user journeys
- **Sync Testing**: Multi-device synchronization scenarios
- **Performance Testing**: Load testing with realistic data volumes
- **Security Testing**: Authentication and authorization testing

### Testing Data Management

- **Test Fixtures**: Standardized test data sets
- **Database Seeding**: Automated test data generation
- **Cleanup Procedures**: Automated test environment reset
- **Mock Services**: Comprehensive external service mocking

### Performance Testing

- **Load Testing**: Concurrent user simulation
- **Stress Testing**: System breaking point identification
- **Sync Performance**: Large dataset synchronization testing
- **Browser Extension Performance**: Memory and CPU usage monitoring

## Security Considerations

### Authentication and Authorization

- **JWT Tokens**: Secure token-based authentication
- **Token Rotation**: Automatic token refresh and rotation
- **Role-Based Access**: Granular permission system
- **Multi-Factor Authentication**: Optional 2FA support

### Data Protection

- **Encryption at Rest**: Database and file storage encryption
- **Encryption in Transit**: TLS 1.3 for all communications
- **API Security**: Rate limiting and input validation
- **Privacy Controls**: Granular data sharing permissions

### Browser Extension Security

- **Content Security Policy**: Strict CSP implementation
- **Secure Storage**: Encrypted local storage
- **Permission Management**: Minimal required permissions
- **Code Integrity**: Extension signing and verification

### Infrastructure Security

- **Container Security**: Secure Docker configurations
- **Network Security**: VPC and firewall configurations
- **Monitoring**: Security event logging and alerting
- **Backup Security**: Encrypted backup storage

This design provides a comprehensive foundation for implementing the bookmark synchronization service with all required features while maintaining scalability, security, and user experience standards.

##

Technology Stack Details

### Storage Solution: MinIO S3-Compatible Storage + Supabase Storage

#### Primary MinIO Storage Architecture

**MinIO S3-Compatible Storage (Primary Storage)**

- **Use Cases**: All file storage including system files, user uploads, and backups
- **Files**: Webpage screenshots, user avatars, collection covers, export archives, system logs
- **Advantages**: High performance, full control, cost-effective, unified storage, proven reliability
- **Configuration**: Direct S3-compatible integration with Go backend services using MinIO client
- **Scalability**: Distributed storage with replication support and clustering capabilities

**Optional Supabase Storage (CDN Enhancement)**

- **Use Cases**: Optional CDN for frequently accessed user files
- **Files**: Cached user avatars and collection covers for global delivery
- **Advantages**: Global CDN distribution for improved user experience
- **Configuration**: Optional integration for enhanced performance

#### Architecture Benefits

- **Performance**: MinIO provides high-performance storage with optimized Go implementation
- **Data Sovereignty**: Complete control over all data with self-hosted solution
- **Cost Optimization**: No external storage costs, predictable infrastructure expenses
- **Scalability**: MinIO distributed storage scales horizontally with replication and clustering
- **Simplicity**: Unified storage solution reduces operational complexity

#### Implementation Strategy

```go
// Storage service interface
type StorageService interface {
    // All files stored in MinIO
    StoreScreenshot(bookmarkID string, image []byte) (string, error)
    StoreBackup(userID string, data []byte) (string, error)
    StoreUserAvatar(userID string, image []byte) (string, error)
    StoreCollectionCover(collectionID string, image []byte) (string, error)
    StoreFile(bucket, key string, data []byte) (string, error)
}

// MinIO storage implementation
type MinIOStorage struct {
    client   *minio.Client  // S3-compatible client for MinIO
    endpoint string
    buckets  map[string]string // Different buckets for different file types
}
```

### Search Engine Comparison

#### Typesense (Recommended)

**Advantages:**

- **Easy Deployment**: Single binary, minimal configuration
- **Excellent Performance**: Sub-millisecond search responses
- **Chinese Language Support**: Built-in Chinese tokenizer with Traditional/Simplified support
- **Real-time Indexing**: Supports live data updates
- **Resource Friendly**: Lower memory usage compared to alternatives

**Chinese Language Features:**

```json
{
  "fields": [
    {
      "name": "title",
      "type": "string",
      "locale": "zh"
    },
    {
      "name": "content",
      "type": "string",
      "locale": "zh"
    }
  ],
  "default_sorting_field": "popularity"
}
```

#### OpenSearch (Enterprise Alternative)

**Advantages:**

- **Powerful Features**: Complete Elasticsearch functionality
- **Superior Chinese Support**: IK analyzer with custom dictionaries
- **Mature and Stable**: Production-proven with extensive ecosystem
- **Rich Plugin Ecosystem**: Extensive plugins and tools

**Chinese Configuration:**

```json
{
  "settings": {
    "analysis": {
      "analyzer": {
        "ik_max_word": {
          "type": "ik_max_word"
        },
        "ik_smart": {
          "type": "ik_smart"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "content": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      }
    }
  }
}
```

**Considerations:**

- **Resource Intensive**: Higher memory and CPU requirements
- **Complex Configuration**: Requires tuning and maintenance
- **Deployment Complexity**: More complex cluster setup

#### Meilisearch

**Advantages:**

- **Developer Friendly**: Intuitive API design and excellent documentation
- **Quick Setup**: Simple configuration and deployment
- **Multi-language Support**: Built-in multi-language tokenization

**Limitations:**

- **Weaker Chinese Support**: Limited Chinese tokenization capabilities
- **Feature Constraints**: Limited complex query and aggregation features
- **Enterprise Features**: Some advanced features require paid version

### Multi-language Search Configuration

#### Typesense Multi-language Setup

```json
{
  "name": "bookmarks",
  "fields": [
    {
      "name": "title",
      "type": "string",
      "locale": "zh"
    },
    {
      "name": "title_en",
      "type": "string",
      "locale": "en"
    },
    {
      "name": "content",
      "type": "string",
      "locale": "zh"
    },
    {
      "name": "tags",
      "type": "string[]",
      "facet": true
    }
  ],
  "token_separators": ["-", "_", ".", " "],
  "symbols_to_index": ["@", "#"],
  "default_sorting_field": "popularity"
}
```

#### Search Query Examples

```javascript
// Chinese search with pinyin support
const searchParams = {
  q: "書籤管理",
  query_by: "title,content",
  facet_by: "tags,category",
  filter_by: "language:zh",
  sort_by: "popularity:desc",
};

// Multi-language search
const multiLangSearch = {
  q: "bookmark management",
  query_by: "title,title_en,content",
  facet_by: "language,tags",
  group_by: "language",
  group_limit: 10,
};
```

### Recommended Architecture Stack

**Final Technology Stack:**

- **Frontend**: Browser Extensions (Chrome, Firefox, Safari) + Responsive Web App
- **Backend**: Go + Gin Framework
- **Database**: Self-hosted Supabase PostgreSQL with Row Level Security
- **Cache**: Redis (with Pub/Sub for real-time sync)
- **Authentication**: Self-hosted Supabase Auth with OAuth providers
- **Storage**: MinIO S3-compatible storage (primary storage for all files)
- **Real-time**: Self-hosted Supabase Realtime + Gorilla WebSocket
- **Search**: Typesense with Chinese language support
- **Load Balancer**: Nginx
- **Containerization**: Docker + Docker Compose
- **Monitoring**: Prometheus + Grafana

**Deployment Configuration:**

````yaml
# docker-compose.yml
version: "3.8"
services:
  # Self-hosted Supabase Stack
  supabase-db:
    image: supabase/postgres:15.1.0.117
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: postgres
      POSTGRES_USER: postgres
      POSTGRES_HOST: supabase-db
      POSTGRES_PORT: 5432
    volumes:
      - supabase_db_data:/var/lib/postgresql/data
      - ./supabase/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    networks:
      - bookmark-network

  supabase-auth:
    image: supabase/gotrue:v2.99.0
    environment:
      GOTRUE_API_HOST: 0.0.0.0
      GOTRUE_API_PORT: 9999
      GOTRUE_DB_DRIVER: postgres
      GOTRUE_DB_DATABASE_URL: postgres://supabase_auth_admin:${AUTH_DB_PASSWORD}@supabase-db:5432/postgres
      GOTRUE_SITE_URL: ${SITE_URL}
      GOTRUE_URI_ALLOW_LIST: ${ADDITIONAL_REDIRECT_URLS}
      GOTRUE_JWT_ADMIN_ROLES: service_role
      GOTRUE_JWT_AUD: authenticated
      GOTRUE_JWT_DEFAULT_GROUP_NAME: authenticated
      GOTRUE_JWT_EXP: ${JWT_EXPIRY}
      GOTRUE_JWT_SECRET: ${JWT_SECRET}

      # OAuth Providers Configuration
      GOTRUE_EXTERNAL_GITHUB_ENABLED: "true"
      GOTRUE_EXTERNAL_GITHUB_CLIENT_ID: ${GITHUB_CLIENT_ID}
      GOTRUE_EXTERNAL_GITHUB_SECRET: ${GITHUB_CLIENT_SECRET}
      GOTRUE_EXTERNAL_GITHUB_REDIRECT_URI: ${SITE_URL}/auth/v1/callback

      GOTRUE_EXTERNAL_GOOGLE_ENABLED: "true"
      GOTRUE_EXTERNAL_GOOGLE_CLIENT_ID: ${GOOGLE_CLIENT_ID}
      GOTRUE_EXTERNAL_GOOGLE_SECRET: ${GOOGLE_CLIENT_SECRET}
      GOTRUE_EXTERNAL_GOOGLE_REDIRECT_URI: ${SITE_URL}/auth/v1/callback

      # Email Configuration
      GOTRUE_SMTP_HOST: ${SMTP_HOST}
      GOTRUE_SMTP_PORT: ${SMTP_PORT}
      GOTRUE_SMTP_USER: ${SMTP_USER}
      GOTRUE_SMTP_PASS: ${SMTP_PASS}
      GOTRUE_SMTP_ADMIN_EMAIL: ${ADMIN_EMAIL}
      GOTRUE_MAILER_AUTOCONFIRM: "false"
      GOTRUE_MAILER_SECURE_EMAIL_CHANGE_ENABLED: "true"
    depends_on:
      - supabase-db
    ports:
      - "9999:9999"
    networks:
      - bookmark-network

  supabase-rest:
    image: postgrest/postgrest:v11.2.0
    environment:
      PGRST_DB_URI: postgres://authenticator:${POSTGRES_PASSWORD}@supabase-db:5432/postgres
      PGRST_DB_SCHEMAS: ${PGRST_DB_SCHEMAS}
      PGRST_DB_ANON_ROLE: anon
      PGRST_JWT_SECRET: ${JWT_SECRET}
      PGRST_DB_USE_LEGACY_GUCS: "false"
    depends_on:
      - supabase-db
    ports:
      - "3000:3000"
    networks:
      - bookmark-network

  supabase-realtime:
    image: supabase/realtime:v2.25.35
    environment:
      PORT: 4000
      DB_HOST: supabase-db
      DB_PORT: 5432
      DB_USER: supabase_realtime_admin
      DB_PASSWORD: ${REALTIME_DB_PASSWORD}
      DB_NAME: postgres
      DB_AFTER_CONNECT_QUERY: 'SET search_path TO _realtime'
      DB_ENC_KEY: ${DB_ENC_KEY}
      API_JWT_SECRET: ${JWT_SECRET}
      FLY_ALLOC_ID: fly123
      FLY_APP_NAME: realtime
      SECRET_KEY_BASE: ${SECRET_KEY_BASE}
      ERL_AFLAGS: -proto_dist inet_tcp
      ENABLE_TAILSCALE: "false"
      DNS_NODES: "''"
    depends_on:
      - supabase-db
    ports:
      - "4000:4000"
    networks:
      - bookmark-network

  # Additional Services
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - bookmark-network

  typesense:
    image: typesense/typesense:0.25.1
    environment:
      TYPESENSE_DATA_DIR: /data
      TYPESENSE_API_KEY: ${TYPESENSE_API_KEY}
      TYPESENSE_ENABLE_CORS: "true"
    ports:
      - "8108:8108"
    volumes:
      - typesense_data:/data
    networks:
      - bookmark-network

  minio:
    image: minio/minio:RELEASE.2024-01-16T16-07-38Z
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD:-minioadmin}
      MINIO_BROWSER_REDIRECT_URL: http://localhost:9001
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data
    networks:
      - bookmark-network

  # Go Backend Services
  bookmark-api:
    build:
      context: ./backend
      dockerfile: Dockerfile
    environment:
      # Supabase Configuration
      SUPABASE_URL: http://supabase-rest:3000
      SUPABASE_ANON_KEY: ${SUPABASE_ANON_KEY}
      SUPABASE_SERVICE_ROLE_KEY: ${SUPABASE_SERVICE_ROLE_KEY}
      SUPABASE_DB_URL: postgres://postgres:${POSTGRES_PASSWORD}@supabase-db:5432/postgres

      # Other Services
      REDIS_URL: redis://redis:6379
      TYPESENSE_URL: http://typesense:8108
      MINIO_URL: http://minio:9000

      # JWT Configuration
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - supabase-db
      - supabase-auth
      - supabase-rest
      - redis
      - typesense
      - minio
    ports:
      - "8080:8080"
    networks:
      - bookmark-network

  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - bookmark-api
      - supabase-auth
      - supabase-rest
    networks:
      - bookmark-network

volumes:
  supabase_db_data:
  redis_data:
  typesense_data:
  minio_data:

networks:
  bookmark-network:
    driver: bridge
```OLE: anon
      PGRST_JWT_SECRET: ${JWT_SECRET}
      PGRST_DB_USE_LEGACY_GUCS: "false"
    depends_on:
      - supabase-db
    ports:
      - "3000:3000"

  supabase-realtime:
    image: supabase/realtime:v2.25.50
    environment:
      PORT: 4000
      DB_HOST: supabase-db
      DB_PORT: 5432
      DB_USER: supabase_realtime_admin
      DB_PASSWORD: ${REALTIME_DB_PASSWORD}
      DB_NAME: postgres
      DB_AFTER_CONNECT_QUERY: "SET search_path TO _realtime"
      DB_ENC_KEY: supabaserealtime
      API_JWT_SECRET: ${JWT_SECRET}
      FLY_ALLOC_ID: fly123
      FLY_APP_NAME: realtime
      SECRET_KEY_BASE: ${SECRET_KEY_BASE}
      ERL_AFLAGS: -proto_dist inet_tcp
      ENABLE_TAILSCALE: "false"
      DNS_NODES: "''"
    depends_on:
      - supabase-db
    ports:
      - "4000:4000"

  supabase-storage:
    image: supabase/storage-api:v0.40.4
    environment:
      ANON_KEY: ${ANON_KEY}
      SERVICE_KEY: ${SERVICE_KEY}
      POSTGREST_URL: http://supabase-rest:3000
      PGRST_JWT_SECRET: ${JWT_SECRET}
      DATABASE_URL: postgres://supabase_storage_admin:${STORAGE_DB_PASSWORD}@supabase-db:5432/postgres
      FILE_SIZE_LIMIT: 52428800
      STORAGE_BACKEND: file
      FILE_STORAGE_BACKEND_PATH: /var/lib/storage
      TENANT_ID: stub
      REGION: stub
      GLOBAL_S3_BUCKET: stub
    depends_on:
      - supabase-db
      - supabase-rest
    volumes:
      - supabase_storage_data:/var/lib/storage
    ports:
      - "5000:5000"

  # Additional Services
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  typesense:
    image: typesense/typesense:0.25.1
    environment:
      TYPESENSE_DATA_DIR: /data
      TYPESENSE_API_KEY: ${TYPESENSE_API_KEY}
    volumes:
      - typesense_data:/data
    ports:
      - "8108:8108"

  minio:
    image: minio/minio:RELEASE.2024-01-16T16-07-38Z
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
      MINIO_BROWSER_REDIRECT_URL: ${MINIO_BROWSER_REDIRECT_URL}
      MINIO_SERVER_URL: ${MINIO_SERVER_URL}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000" # API port
      - "9001:9001" # Console port

  # Main Application
  bookmark-api:
    build: ./backend
    environment:
      # Supabase Configuration
      SUPABASE_URL: http://supabase-rest:3000
      SUPABASE_ANON_KEY: ${ANON_KEY}
      SUPABASE_SERVICE_KEY: ${SERVICE_KEY}
      SUPABASE_AUTH_URL: http://supabase-auth:9999
      SUPABASE_REALTIME_URL: ws://supabase-realtime:4000/socket

      # Database
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@supabase-db:5432/postgres

      # Other Services
      REDIS_URL: redis://redis:6379
      TYPESENSE_URL: http://typesense:8108
      TYPESENSE_API_KEY: ${TYPESENSE_API_KEY}
      MINIO_ENDPOINT: http://minio:9000
      MINIO_ACCESS_KEY: ${MINIO_ROOT_USER}
      MINIO_SECRET_KEY: ${MINIO_ROOT_PASSWORD}

      # JWT Configuration
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - supabase-db
      - supabase-auth
      - supabase-rest
      - supabase-storage
      - redis
      - typesense
      - minio
    ports:
      - "8080:8080"

volumes:
  supabase_db_data:
  supabase_storage_data:
  redis_data:
  typesense_data:
  minio_data:
````

## Hybrid Storage Architecture Details

### Storage Strategy

#### MinIO (Primary Storage) - All Files

**Use Cases:**

- Webpage screenshots and thumbnails
- User profile avatars and collection covers
- System backup files and exports
- Log files and analytics data
- User-uploaded content and shared media files
- Temporary processing files

**Configuration:**

```go
type MinIOConfig struct {
    Endpoint    string `env:"MINIO_ENDPOINT"`
    AccessKey   string `env:"MINIO_ACCESS_KEY"`
    SecretKey   string `env:"MINIO_SECRET_KEY"`
    Region      string `env:"MINIO_REGION" default:"us-east-1"`
    UseSSL      bool   `env:"MINIO_USE_SSL" default:"false"`
    SystemBucket string `env:"MINIO_SYSTEM_BUCKET" default:"bookmarks-system"`
    UserBucket   string `env:"MINIO_USER_BUCKET" default:"bookmarks-users"`
}

// MinIO client for all file operations
func (s *FileService) StoreScreenshot(bookmarkID string, imageData []byte) (string, error) {
    key := fmt.Sprintf("screenshots/%s/%s.webp",
        bookmarkID[:2], bookmarkID)

    return s.minioClient.PutObject(context.Background(),
        s.config.SystemBucket, key, bytes.NewReader(imageData),
        int64(len(imageData)), minio.PutObjectOptions{
            ContentType: "image/webp",
        })
}

func (s *FileService) StoreUserAvatar(userID string, imageData []byte) (string, error) {
    key := fmt.Sprintf("avatars/%s.webp", userID)

    _, err := s.minioClient.PutObject(context.Background(),
        s.config.UserBucket, key, bytes.NewReader(imageData),
        int64(len(imageData)), minio.PutObjectOptions{
            ContentType: "image/webp",
        })

    if err != nil {
        return "", err
    }

    // Return public URL
    return fmt.Sprintf("%s/%s/%s", s.config.Endpoint, s.config.UserBucket, key), nil
}
```

### Benefits of MinIO Primary Storage

#### Cost Optimization

- **No External Costs**: Complete elimination of external storage fees
- **Predictable Expenses**: Only infrastructure costs, no variable pricing
- **Efficient Storage**: Rust's performance optimizations reduce resource usage
- **Long-term Savings**: Significant cost reduction for high-usage scenarios

#### Performance Benefits

- **Rust Performance**: Native Rust implementation provides superior speed
- **Direct Access**: No network latency for server-to-server operations
- **Optimized I/O**: Efficient concurrent read/write operations
- **Local Caching**: Built-in caching mechanisms for frequently accessed files

#### Security and Compliance

- **Complete Control**: Full ownership and control over all data
- **Data Sovereignty**: No third-party access to sensitive information
- **Custom Security**: Implement custom security policies as needed
- **Compliance**: Easier to meet data residency and compliance requirements

#### Operational Advantages

- **Simplified Architecture**: Single storage solution reduces complexity
- **Self-Contained**: No external dependencies for storage operations
- **Unified Backup**: Single backup strategy for all data types
- **Scalability**: Distributed storage with built-in replication

### Implementation Example

```go
// Unified MinIO storage interface
type StorageManager struct {
    minio *MinIOClient
    config *MinIOConfig
}

func (sm *StorageManager) StoreFile(fileType FileType, data []byte, metadata FileMetadata) (string, error) {
    bucket := sm.getBucketForFileType(fileType)
    key := sm.generateKey(fileType, metadata)

    return sm.minio.Store(bucket, key, data, metadata)
}

func (sm *StorageManager) getBucketForFileType(fileType FileType) string {
    switch fileType {
    case FileTypeScreenshot, FileTypeBackup, FileTypeExport:
        return sm.config.SystemBucket
    case FileTypeAvatar, FileTypeCollectionCover, FileTypeUserUpload:
        return sm.config.UserBucket
    default:
        return sm.config.SystemBucket
    }
}

// File type enumeration
type FileType int

const (
    FileTypeScreenshot FileType = iota
    FileTypeBackup
    FileTypeExport
    FileTypeAvatar
    FileTypeCollectionCover
    FileTypeUserUpload
)
```

## Self-hosted Supabase Architecture Benefits

### Complete Data Sovereignty

- **Full Control**: All data remains on your infrastructure
- **Privacy**: No third-party access to sensitive user data
- **Compliance**: Easier to meet data residency requirements
- **Customization**: Can modify and extend Supabase functionality

### Cost Optimization

- **No Usage Limits**: No per-request or storage limitations
- **Predictable Costs**: Only infrastructure costs, no variable pricing
- **Scaling Freedom**: Scale resources based on actual needs
- **Long-term Savings**: Significant cost reduction for high-usage scenarios

### Technical Advantages

- **Integration**: Better integration with other self-hosted services
- **Performance**: Reduced latency with local deployment
- **Reliability**: No dependency on external service availability
- **Security**: Complete control over security configurations

### Self-hosted Supabase Components

#### Database (PostgreSQL)

```sql
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pgjwt";

-- Row Level Security setup
ALTER TABLE bookmarks ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only see their own bookmarks" ON bookmarks
    FOR ALL USING (auth.uid() = user_id);
```

#### Authentication Service

```go
// Go integration with self-hosted Supabase Auth
type SupabaseAuth struct {
    AuthURL    string
    JWTSecret  string
    AnonKey    string
    ServiceKey string
}

func (s *SupabaseAuth) ValidateToken(token string) (*User, error) {
    // Validate JWT token against self-hosted auth service
    claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.JWTSecret), nil
    })

    if err != nil {
        return nil, err
    }

    // Extract user information from claims
    userID := claims.Claims.(jwt.MapClaims)["sub"].(string)
    email := claims.Claims.(jwt.MapClaims)["email"].(string)

    return &User{ID: userID, Email: email}, nil
}
```

#### Storage Integration

```go
// Hybrid storage with self-hosted Supabase Storage
type HybridStorageManager struct {
    supabaseStorage *SupabaseStorageClient
    minio          *MinIOClient
}

func (hsm *HybridStorageManager) StoreUserFile(userID string, file []byte, fileName string) (string, error) {
    // Use self-hosted Supabase Storage for user files
    bucket := "user-uploads"
    path := fmt.Sprintf("%s/%s", userID, fileName)

    return hsm.supabaseStorage.Upload(bucket, path, file)
}

func (hsm *HybridStorageManager) StoreSystemFile(file []byte, fileName string) (string, error) {
    // Use MinIO for system files
    return hsm.minio.Store(fileName, file)
}
```

### Environment Configuration

```bash
# .env file for self-hosted Supabase
# Database
POSTGRES_PASSWORD=your_secure_password
AUTH_DB_PASSWORD=auth_secure_password
REALTIME_DB_PASSWORD=realtime_secure_password
STORAGE_DB_PASSWORD=storage_secure_password

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-token-with-at-least-32-characters
JWT_EXPIRY=3600
SECRET_KEY_BASE=your-secret-key-base-for-realtime

# API Keys
ANON_KEY=your-anon-key
SERVICE_KEY=your-service-role-key

# Application URLs
SITE_URL=https://your-domain.com
ADDITIONAL_REDIRECT_URLS=https://your-domain.com/auth/callback,chrome-extension://your-extension-id/auth/callback

# OAuth Providers
# GitHub OAuth
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_google_client_secret

# Apple ID OAuth
APPLE_CLIENT_ID=your.apple.service.id
APPLE_CLIENT_SECRET=your_apple_client_secret

# Email Configuration (for email/password auth)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
ADMIN_EMAIL=admin@your-domain.com

# PostgREST Configuration
PGRST_DB_SCHEMAS=public,storage,graphql_public

# Other Services
TYPESENSE_API_KEY=your-typesense-api-key
MINIO_ROOT_PASSWORD=your-minio-password
```

### Deployment Considerations

#### Resource Requirements

- **CPU**: 4+ cores recommended for full stack
- **Memory**: 8GB+ RAM for optimal performance
- **Storage**: SSD recommended, 100GB+ for production
- **Network**: Stable internet connection for external integrations

#### Security Setup

```nginx
# Nginx configuration for self-hosted Supabase
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # Supabase REST API
    location /rest/v1/ {
        proxy_pass http://localhost:3000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Supabase Auth
    location /auth/v1/ {
        proxy_pass http://localhost:9999/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Supabase Storage
    location /storage/v1/ {
        proxy_pass http://localhost:5000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Supabase Realtime
    location /realtime/v1/ {
        proxy_pass http://localhost:4000/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

#### Backup Strategy

```bash
#!/bin/bash
# Automated backup script
DATE=$(date +%Y%m%d_%H%M%S)

# Database backup
docker exec supabase-db pg_dump -U postgres postgres > backup_db_$DATE.sql

# MinIO backup (primary storage)
docker exec minio tar -czf - /data > backup_minio_$DATE.tar.gz

# Optional: Create incremental backups using MinIO built-in features
docker exec minio mc mirror --overwrite /data /backup/backup_$DATE

# Verify backup integrity
docker exec minio mc ls /backup/backup_$DATE

# Upload to remote backup location (optional)
# aws s3 cp backup_* s3://your-backup-bucket/
```

This self-hosted architecture provides complete control, optimal performance, and cost-effectiveness while maintaining all the benefits of the Supabase ecosystem for the bookmark synchronization service.##
Multi-Provider Authentication Implementation

### Supported Authentication Methods

#### 1. Email/Password Authentication

- Traditional registration and login
- Email verification required
- Password reset functionality
- Secure password hashing with bcrypt

#### 2. GitHub OAuth

- Perfect for developer-focused users
- Access to GitHub profile information
- Automatic avatar and name import
- Repository access (optional, for bookmark import)

#### 3. Google OAuth

- Most widely used OAuth provider
- Access to Google profile information
- Gmail integration for notifications
- Google Drive integration (future feature)

#### 4. Apple ID

- Required for iOS users
- Privacy-focused authentication
- Minimal data sharing
- Sign in with Apple compliance

### Browser Extension Authentication Flow

#### OAuth Flow for Extensions

```javascript
// Browser extension OAuth implementation
class AuthManager {
  constructor(supabaseUrl, anonKey) {
    this.supabaseUrl = supabaseUrl;
    this.anonKey = anonKey;
  }

  // Initiate OAuth flow
  async signInWithProvider(provider) {
    const redirectURL = chrome.identity.getRedirectURL();
    const authURL = `${
      this.supabaseUrl
    }/auth/v1/authorize?provider=${provider}&redirect_to=${encodeURIComponent(
      redirectURL
    )}`;

    return new Promise((resolve, reject) => {
      chrome.identity.launchWebAuthFlow(
        {
          url: authURL,
          interactive: true,
        },
        (responseUrl) => {
          if (chrome.runtime.lastError) {
            reject(chrome.runtime.lastError);
            return;
          }

          // Extract tokens from callback URL
          const urlParams = new URLSearchParams(responseUrl.split("#")[1]);
          const accessToken = urlParams.get("access_token");
          const refreshToken = urlParams.get("refresh_token");

          if (accessToken) {
            this.storeTokens(accessToken, refreshToken);
            resolve({ accessToken, refreshToken });
          } else {
            reject(new Error("Authentication failed"));
          }
        }
      );
    });
  }

  // Email/Password authentication
  async signInWithEmail(email, password) {
    const response = await fetch(
      `${this.supabaseUrl}/auth/v1/token?grant_type=password`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          apikey: this.anonKey,
        },
        body: JSON.stringify({ email, password }),
      }
    );

    if (!response.ok) {
      throw new Error("Authentication failed");
    }

    const data = await response.json();
    this.storeTokens(data.access_token, data.refresh_token);
    return data;
  }

  // Store tokens securely
  async storeTokens(accessToken, refreshToken) {
    await chrome.storage.local.set({
      auth_access_token: accessToken,
      auth_refresh_token: refreshToken,
      auth_timestamp: Date.now(),
    });
  }

  // Get current user
  async getCurrentUser() {
    const tokens = await chrome.storage.local.get(["auth_access_token"]);
    if (!tokens.auth_access_token) {
      return null;
    }

    const response = await fetch(`${this.supabaseUrl}/auth/v1/user`, {
      headers: {
        Authorization: `Bearer ${tokens.auth_access_token}`,
        apikey: this.anonKey,
      },
    });

    if (!response.ok) {
      return null;
    }

    return await response.json();
  }
}
```

### Go Backend Authentication Integration

```go
// Multi-provider authentication service
type MultiAuthService struct {
    supabaseAuth *SupabaseAuthClient
    config       *AuthConfig
}

type AuthConfig struct {
    SupabaseURL    string
    AnonKey        string
    ServiceKey     string
    JWTSecret      string

    // OAuth Provider Configs
    GitHubClientID     string
    GitHubClientSecret string
    GoogleClientID     string
    GoogleClientSecret string
    AppleClientID      string
    AppleClientSecret  string
}

// User model with provider support
type User struct {
    ID           string                 `json:"id" db:"id"`
    Email        string                 `json:"email" db:"email"`
    Name         string                 `json:"name" db:"name"`
    Avatar       string                 `json:"avatar_url" db:"avatar_url"`
    Provider     string                 `json:"provider" db:"provider"`
    ProviderID   string                 `json:"provider_id" db:"provider_id"`
    Metadata     map[string]interface{} `json:"user_metadata" db:"user_metadata"`
    CreatedAt    time.Time             `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time             `json:"updated_at" db:"updated_at"`
    LastSignIn   *time.Time            `json:"last_sign_in" db:"last_sign_in"`
}

// Authentication middleware
func (m *MultiAuthService) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Extract Bearer token
        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        token := tokenParts[1]
        user, err := m.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Store user in context
        c.Set("user", user)
        c.Next()
    }
}

// Token validation
func (m *MultiAuthService) ValidateToken(tokenString string) (*User, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(m.config.JWTSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid token claims")
    }

    // Extract user information
    user := &User{
        ID:    claims["sub"].(string),
        Email: claims["email"].(string),
    }

    // Extract provider information
    if appMetadata, ok := claims["app_metadata"].(map[string]interface{}); ok {
        if provider, ok := appMetadata["provider"].(string); ok {
            user.Provider = provider
        }
        if providerID, ok := appMetadata["provider_id"].(string); ok {
            user.ProviderID = providerID
        }
    }

    // Extract user metadata
    if userMetadata, ok := claims["user_metadata"].(map[string]interface{}); ok {
        user.Metadata = userMetadata
        if name, ok := userMetadata["full_name"].(string); ok {
            user.Name = name
        }
        if avatar, ok := userMetadata["avatar_url"].(string); ok {
            user.Avatar = avatar
        }
    }

    return user, nil
}

// OAuth provider configuration
func (m *MultiAuthService) GetOAuthConfig(provider string) (*oauth2.Config, error) {
    var config *oauth2.Config

    switch provider {
    case "github":
        config = &oauth2.Config{
            ClientID:     m.config.GitHubClientID,
            ClientSecret: m.config.GitHubClientSecret,
            Scopes:       []string{"user:email"},
            Endpoint:     github.Endpoint,
            RedirectURL:  fmt.Sprintf("%s/auth/callback/github", m.config.SupabaseURL),
        }
    case "google":
        config = &oauth2.Config{
            ClientID:     m.config.GoogleClientID,
            ClientSecret: m.config.GoogleClientSecret,
            Scopes:       []string{"openid", "profile", "email"},
            Endpoint:     google.Endpoint,
            RedirectURL:  fmt.Sprintf("%s/auth/callback/google", m.config.SupabaseURL),
        }
    case "apple":
        // Apple ID configuration is more complex and requires special handling
        return nil, fmt.Errorf("apple oauth config requires special implementation")
    default:
        return nil, fmt.Errorf("unsupported provider: %s", provider)
    }

    return config, nil
}
```

### Database Schema for Multi-Provider Auth

```sql
-- Users table with provider support
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(50) NOT NULL DEFAULT 'email',
    provider_id VARCHAR(255),
    user_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_sign_in TIMESTAMP WITH TIME ZONE
);

-- User identities for multiple providers per user
CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider, provider_id)
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider, provider_id);
CREATE INDEX idx_user_identities_user_id ON user_identities(user_id);
CREATE INDEX idx_user_identities_provider ON user_identities(provider, provider_id);

-- Row Level Security policies
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_identities ENABLE ROW LEVEL SECURITY;

-- Users can only see their own data
CREATE POLICY "Users can view own profile" ON users
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON users
    FOR UPDATE USING (auth.uid() = id);

-- User identities policies
CREATE POLICY "Users can view own identities" ON user_identities
    FOR SELECT USING (auth.uid() = user_id);
```

### OAuth Provider Setup Instructions

#### GitHub OAuth App Setup

1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Create new OAuth App with:
   - Application name: "Your Bookmark Service"
   - Homepage URL: "https://your-domain.com"
   - Authorization callback URL: "https://your-domain.com/auth/v1/callback"

#### Google OAuth Setup

1. Go to Google Cloud Console > APIs & Services > Credentials
2. Create OAuth 2.0 Client ID with:
   - Application type: Web application
   - Authorized redirect URIs: "https://your-domain.com/auth/v1/callback"

#### Apple ID Setup

1. Go to Apple Developer > Certificates, Identifiers & Profiles
2. Create new Services ID with:
   - Description: "Your Bookmark Service"
   - Identifier: "com.yourdomain.bookmarks"
   - Configure Sign in with Apple with your domain

This multi-provider authentication system provides users with flexible sign-in options while maintaining security and user experience consistency across all platforms.

##

Updated Architecture with MinIO + Typesense

### Final Technology Stack Selection

**Selected Components:**

- **Storage**: MinIO (primary) + Supabase Storage (hybrid)
- **Search**: Typesense with Chinese language optimization
- **Backend**: Go + Gin Framework
- **Database**: Self-hosted PostgreSQL
- **Cache**: Redis with Pub/Sub
- **Authentication**: Supabase Auth

### Hybrid Storage Strategy

#### MinIO Configuration

```yaml
# MinIO deployment configuration
minio:
  image: minio/minio:RELEASE.2024-01-16T16-07-38Z
  environment:
    MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}
    MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    MINIO_BROWSER_REDIRECT_URL: http://localhost:9001
  volumes:
    - minio_data:/data
  ports:
    - "9000:9000"
    - "9001:9001"
  command: server /data --console-address ":9001"
```

#### Supabase Storage Integration

```go
// Go integration example
type StorageService struct {
    minioClient    *minio.Client
    supabaseClient *supabase.Client
}

func (s *StorageService) StoreFile(file File) error {
    if file.Size > LARGE_FILE_THRESHOLD {
        // Use MinIO for large files
        return s.minioClient.PutObject(file)
    } else {
        // Use Supabase Storage for small files with CDN
        return s.supabaseClient.Storage.Upload(file)
    }
}
```

#### Storage Decision Matrix

| File Type           | Size   | Storage  | Reason                 |
| ------------------- | ------ | -------- | ---------------------- |
| User Avatars        | < 1MB  | Supabase | CDN + Auth integration |
| Webpage Screenshots | 1-10MB | MinIO    | Performance + Cost     |
| Backup Archives     | > 10MB | MinIO    | Cost efficiency        |
| Static Assets       | < 5MB  | Supabase | CDN distribution       |

### Typesense Chinese Language Configuration

#### Schema Definition

```json
{
  "name": "bookmarks",
  "fields": [
    {
      "name": "title",
      "type": "string",
      "locale": "zh",
      "infix": true
    },
    {
      "name": "description",
      "type": "string",
      "locale": "zh",
      "optional": true
    },
    {
      "name": "content",
      "type": "string",
      "locale": "zh",
      "optional": true
    },
    {
      "name": "tags",
      "type": "string[]",
      "facet": true,
      "locale": "zh"
    },
    {
      "name": "url",
      "type": "string",
      "index": false
    },
    {
      "name": "created_at",
      "type": "int64"
    },
    {
      "name": "popularity_score",
      "type": "float"
    }
  ],
  "default_sorting_field": "popularity_score",
  "token_separators": ["-", "_", ".", " ", "，", "。", "！", "？"],
  "symbols_to_index": ["@", "#", "$"]
}
```

#### Multi-language Search Implementation

```go
// Go search service implementation
type SearchService struct {
    client *typesense.Client
}

func (s *SearchService) SearchBookmarks(query string, lang string) (*SearchResult, error) {
    searchParams := &api.SearchCollectionParams{
        Q:         query,
        QueryBy:   "title,description,content,tags",
        FilterBy:  fmt.Sprintf("language:%s", lang),
        SortBy:    "popularity_score:desc",
        FacetBy:   "tags,category",
        MaxFacetValues: 20,
        NumTypos:  "2,1,0",
        TypoTokensThreshold: 1,
    }

    // Handle Chinese-specific search optimizations
    if lang == "zh" {
        searchParams.QueryByWeights = "4,3,2,1" // Prioritize title for Chinese
        searchParams.PrefixMatch = false        // Disable prefix for better Chinese matching
    }

    return s.client.Collection("bookmarks").Documents().Search(searchParams)
}
```

#### Chinese Language Features

- **Traditional/Simplified Support**: Automatic detection and conversion
- **Pinyin Search**: Support for pinyin input searching Chinese content
- **Word Segmentation**: Intelligent Chinese word boundary detection
- **Synonym Support**: Built-in Chinese synonym handling
- **Typo Tolerance**: Optimized for Chinese character input errors

### Performance Optimizations

#### MinIO Optimizations

```yaml
# MinIO configuration
environment:
  MINIO_CACHE_DRIVES: "/tmp/cache1,/tmp/cache2"
  MINIO_CACHE_EXCLUDE: "*.pdf"
  MINIO_CACHE_QUOTA: "80"
  MINIO_CACHE_AFTER: "3"
  MINIO_CACHE_WATERMARK_LOW: "70"
  MINIO_CACHE_WATERMARK_HIGH: "90"
```

#### Typesense Performance Tuning

```json
{
  "memory_num_shards": 4,
  "search_only_mode": false,
  "num_memory_shards": 4,
  "log_slow_requests_time_ms": 2000,
  "healthy_read_lag": 1000,
  "healthy_write_lag": 500
}
```

### Deployment Architecture

#### Docker Compose Configuration

```yaml
version: "3.8"
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: bookmarks
      POSTGRES_USER: bookmarks_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  typesense:
    image: typesense/typesense:0.25.1
    environment:
      TYPESENSE_DATA_DIR: /data
      TYPESENSE_API_KEY: ${TYPESENSE_API_KEY}
      TYPESENSE_ENABLE_CORS: true
    volumes:
      - typesense_data:/data
    ports:
      - "8108:8108"

  minio:
    image: minio/minio:RELEASE.2024-01-16T16-07-38Z
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"

  bookmark-api:
    build: ./backend
    environment:
      DATABASE_URL: postgres://bookmarks_user:${DB_PASSWORD}@postgres:5432/bookmarks
      REDIS_URL: redis://redis:6379
      TYPESENSE_URL: http://typesense:8108
      TYPESENSE_API_KEY: ${TYPESENSE_API_KEY}
      MINIO_URL: http://minio:9000
      SUPABASE_URL: ${SUPABASE_URL}
      SUPABASE_ANON_KEY: ${SUPABASE_ANON_KEY}
    depends_on:
      - postgres
      - redis
      - typesense
      - minio
    ports:
      - "8080:8080"

  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - bookmark-api

volumes:
  postgres_data:
  redis_data:
  typesense_data:
  minio_data:
```

### Self-hosted Supabase Configuration

#### Database Setup

The self-hosted Supabase PostgreSQL instance provides:

- **Row Level Security (RLS)**: Automatic data isolation per user
- **Real-time Subscriptions**: Live data updates via WebSocket
- **REST API**: Automatic API generation from database schema
- **Authentication Integration**: Seamless user management

#### Required Environment Variables

```bash
# Database Configuration
POSTGRES_PASSWORD=your_secure_password
AUTH_DB_PASSWORD=your_auth_db_password
REALTIME_DB_PASSWORD=your_realtime_db_password

# JWT Configuration
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY=3600

# Site Configuration
SITE_URL=http://localhost:3000
ADDITIONAL_REDIRECT_URLS=http://localhost:8080

# OAuth Providers
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_password
ADMIN_EMAIL=admin@yourdomain.com

# Supabase Keys
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key

# Other Services
TYPESENSE_API_KEY=your_typesense_key
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_minio_password
DB_ENC_KEY=your_db_encryption_key
SECRET_KEY_BASE=your_secret_key_base
```

#### Database Schema with RLS

```sql
-- Enable Row Level Security
ALTER TABLE bookmarks ENABLE ROW LEVEL SECURITY;
ALTER TABLE collections ENABLE ROW LEVEL SECURITY;
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

-- RLS Policies for bookmarks
CREATE POLICY "Users can view own bookmarks" ON bookmarks
  FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own bookmarks" ON bookmarks
  FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own bookmarks" ON bookmarks
  FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own bookmarks" ON bookmarks
  FOR DELETE USING (auth.uid() = user_id);

-- Public bookmarks policy
CREATE POLICY "Anyone can view public bookmarks" ON bookmarks
  FOR SELECT USING (visibility = 'public');
```

### Advantages of This Architecture

1. **Performance**: MinIO + Typesense both optimized for speed and efficiency
2. **Cost-Effective**: Self-hosted Supabase eliminates external service costs
3. **Scalability**: Each component can scale independently with Supabase read replicas
4. **Security**: Built-in RLS and authentication with Supabase Auth
5. **Chinese Language**: Excellent Chinese search capabilities with Typesense
6. **Maintainability**: Simple deployment and configuration with Docker Compose
7. **Data Sovereignty**: Complete control over all data with self-hosted stack
8. **Real-time Features**: Built-in real-time capabilities with Supabase Realtime

### Nginx Configuration

#### Main Nginx Configuration

```nginx
# nginx/nginx.conf
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
    access_log /var/log/nginx/access.log main;

    # Performance
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    # Compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/s;

    # Upstream definitions
    upstream bookmark_api {
        least_conn;
        server bookmark-api:8080 max_fails=3 fail_timeout=30s;
        # Add more instances for scaling
        # server bookmark-api-2:8080 max_fails=3 fail_timeout=30s;
        # server bookmark-api-3:8080 max_fails=3 fail_timeout=30s;
    }

    upstream supabase_auth {
        server supabase-auth:9999 max_fails=3 fail_timeout=30s;
    }

    upstream supabase_rest {
        server supabase-rest:3000 max_fails=3 fail_timeout=30s;
    }

    upstream supabase_realtime {
        server supabase-realtime:4000 max_fails=3 fail_timeout=30s;
    }

    # Main server block
    server {
        listen 80;
        listen [::]:80;
        server_name localhost;

        # Redirect HTTP to HTTPS in production
        # return 301 https://$server_name$request_uri;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Referrer-Policy "strict-origin-when-cross-origin";

        # CORS headers for browser extensions
        add_header Access-Control-Allow-Origin "*";
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept, Authorization";

        # Handle preflight requests
        location ~* \.(OPTIONS)$ {
            add_header Access-Control-Allow-Origin "*";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Origin, X-Requested-With, Content-Type, Accept, Authorization";
            add_header Access-Control-Max-Age 86400;
            return 204;
        }

        # API routes with rate limiting
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://bookmark_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_connect_timeout 30s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }

        # WebSocket for real-time sync
        location /api/v1/sync/ws {
            proxy_pass http://bookmark_api;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_read_timeout 86400;
        }

        # Supabase Auth
        location /auth/ {
            limit_req zone=auth burst=10 nodelay;
            proxy_pass http://supabase_auth/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Supabase REST API
        location /rest/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://supabase_rest/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Supabase Realtime
        location /realtime/ {
            proxy_pass http://supabase_realtime/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_read_timeout 86400;
        }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        # Static files (if serving web interface)
        location /static/ {
            alias /var/www/static/;
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # Default location
        location / {
            return 404;
        }
    }

    # HTTPS server block (for production)
    server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;
        server_name localhost;

        # SSL configuration
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;

        # Include the same location blocks as HTTP server
        include /etc/nginx/conf.d/locations.conf;
    }
}
```

#### Docker Compose Nginx Service

```yaml
nginx:
  image: nginx:alpine
  volumes:
    - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    - ./nginx/conf.d:/etc/nginx/conf.d:ro
    - ./nginx/ssl:/etc/nginx/ssl:ro
    - ./web/dist:/var/www/static:ro
  ports:
    - "80:80"
    - "443:443"
  depends_on:
    - bookmark-api
    - supabase-auth
    - supabase-rest
    - supabase-realtime
  networks:
    - bookmark-network
  restart: unless-stopped
  healthcheck:
    test: ["CMD", "nginx", "-t"]
    interval: 30s
    timeout: 10s
    retries: 3
```

#### Nginx Advantages

1. **High Performance**: Excellent performance for static content and reverse proxy
2. **Low Resource Usage**: Minimal memory footprint compared to alternatives
3. **Mature and Stable**: Battle-tested in production environments
4. **Rich Feature Set**: Built-in load balancing, SSL termination, rate limiting
5. **Easy Configuration**: Simple, readable configuration syntax
6. **Wide Adoption**: Extensive documentation and community support
7. **WebSocket Support**: Native support for WebSocket proxying
8. **Health Checks**: Built-in upstream health monitoring

This architecture provides the optimal balance of performance, cost, security, and functionality for your bookmark synchronization service with excellent Chinese language support and complete data control through self-hosted Supabase.
