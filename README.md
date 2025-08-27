# iOS App Store Reviews Viewer

A Go-based backend service that polls iOS App Store RSS feeds to collect and serve app reviews, with a React frontend for viewing and managing review data.

## High-Level Architecture

This application follows a layered architecture pattern with clear separation of concerns between different components. The system is designed to continuously poll iOS App Store RSS feeds for app reviews and provide a RESTful API for frontend consumption.

### Core Components

#### 1. **Entry Point (`cmd/server/main.go`)**
- **Purpose**: Application bootstrap and lifecycle management
- **Responsibilities**:
  - Load configuration from environment variables
  - Initialize database connection
  - Start background polling services
  - Setup HTTP server with graceful shutdown
  - Handle OS signals for clean termination

#### 2. **Configuration Layer (`internal/config/`)**
- **Purpose**: Centralized application configuration management
- **Features**:
  - Environment variable-based configuration
  - Default values for server port, database path, and polling settings
  - Configurable logging levels and concurrent polling limits

#### 3. **Data Layer (`internal/repository/`)**
- **Purpose**: Data persistence and retrieval abstraction
- **Components**:
  - **Interface**: Defines contract for data operations (`Repository`)
  - **SQLite Implementation**: Concrete implementation using SQLite database
  - **Operations**: CRUD operations for reviews and app configurations
  - **Database Schema**: Automatic migration with indexes for performance

#### 4. **Business Logic Layer (`internal/services/`)**
- **Purpose**: Core business logic and external service integration
- **Services**:
  - **RSS Service**: Fetches and parses iOS App Store RSS feeds
  - **Polling Manager**: Orchestrates background polling for multiple apps
  - **App Poller**: Individual app polling with configurable intervals

#### 5. **API Layer (`internal/api/`)**
- **Purpose**: HTTP API endpoints and request handling
- **Components**:
  - **Handlers**: Business logic for HTTP requests
  - **Routes**: API endpoint definitions and middleware setup
  - **Middleware**: CORS, rate limiting, and request logging

#### 6. **Models (`internal/models/`)**
- **Purpose**: Data structures and domain models
- **Types**:
  - **Review**: App review data structure
  - **AppConfig**: App polling configuration
  - **RSSFeed/RSSEntry**: RSS feed parsing structures

#### 7. **Utilities (`pkg/logger/`)**
- **Purpose**: Structured logging with configurable levels
- **Features**: JSON-formatted logs with contextual information



## Data Flow

### 1. **Application Startup**
```
main() → config.Load() → repository.NewSQLiteRepository() → 
services.NewPollingManager() → pollingManager.StartAll()
```

### 2. **Background Polling**
```
PollingManager → GetActiveApps() → GetAppConfig() → 
StartPolling() → AppPoller → RSSService.FetchReviews() → 
Parse & Store Reviews
```

### 3. **API Request Flow**
```
HTTP Request → Middleware → Route Handler → 
Repository → Database → Response
```

## Key Design Patterns

### 1. **Dependency Injection**
- Services receive dependencies through constructors
- Repository interface allows for different database implementations
- Logger is injected throughout the system

### 2. **Repository Pattern**
- Abstract data access through interfaces
- SQLite implementation handles database specifics
- Clean separation between business logic and data persistence

### 3. **Service Layer Pattern**
- Business logic encapsulated in service objects
- RSS service handles external API communication
- Polling manager orchestrates background tasks

### 4. **Middleware Pattern**
- Cross-cutting concerns handled through middleware
- CORS, rate limiting, and logging applied consistently
- Easy to add new middleware without changing handlers

## Configuration

The application is configured through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8000` | HTTP server port |
| `DB_PATH` | `./reviews.db` | SQLite database path |
| `POLL_INTERVAL` | `5m` | Default polling interval |
| `MAX_CONCURRENT_POLLS` | `10` | Maximum concurrent RSS fetches |
| `LOG_LEVEL` | `info` | Logging verbosity |

## Database Schema

### Reviews Table
- **id**: Unique review identifier
- **app_id**: iOS App Store app ID
- **author**: Review author name
- **rating**: 1-5 star rating
- **title**: Review title (optional)
- **content**: Review text content
- **submitted_date**: When review was submitted
- **created_at**: When review was stored

### App Configs Table
- **app_id**: iOS App Store app ID (primary key)
- **poll_interval**: Polling frequency in nanoseconds
- **last_poll**: Last successful poll timestamp
- **is_active**: Whether polling is enabled

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/reviews/:appId` | Retrieve reviews for an app |
| `POST` | `/api/apps/:appId/configure` | Configure app polling settings |
| `GET` | `/api/polling/status` | Get polling service status |
| `GET` | `/health` | Health check endpoint |

## Background Processing

The system maintains active polling for configured apps:

1. **Startup**: Loads all active app configurations and starts polling
2. **Runtime**: Each app has its own poller with configurable intervals
3. **Shutdown**: Gracefully stops all pollers and saves state
4. **Error Handling**: Logs errors and continues operation for other apps

## Scalability Considerations

- **Concurrent Polling**: Configurable limit on simultaneous RSS fetches
- **Database Indexing**: Optimized queries for app_id and date ranges
- **Graceful Shutdown**: Proper cleanup of background goroutines
- **Error Isolation**: Individual app failures don't affect others

## Development

### Prerequisites
- Go 1.21+
- SQLite3
- Node.js 18+ (for frontend)

### Running the Application
```bash
# Backend
make dev
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Project Structure

```
review-app/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── api/            # HTTP API layer
│   ├── config/         # Configuration management
│   ├── models/         # Data structures
│   ├── repository/     # Data access layer
│   └── services/       # Business logic services
├── pkg/                # Public packages
│   └── logger/         # Logging utilities
└── web/                # React frontend
```

This architecture provides a solid foundation for a scalable review collection system with clear separation of concerns, testable components, and maintainable code structure. 