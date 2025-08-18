# Requirements Document

## Introduction

This feature involves creating a self-hosted multi-user bookmark synchronization service that provides cross-browser bookmark management with a visual interface similar to Toby. The system will consist of a backend API service for bookmark storage, synchronization, sharing, and user management, along with browser extensions for Chrome, Firefox, and Safari that provide bookmark sync functionality, social features, and a Toby-like visual bookmark management interface.

## Requirements Priority Classification

### 🔴 Phase 1: Core MVP (Essential Features)

### Requirement 1 - User Authentication and Security

**User Story:** As a user, I want to secure my bookmark data with authentication, so that only I can access my personal bookmarks.

#### BDD Scenarios

**Scenario 1: First-time user authentication**
```gherkin
Given a user installs the browser extension for the first time
When they open the extension popup
Then they should see a login/register form
And they should not be able to access bookmark features without authentication
```

**Scenario 2: Successful user login**
```gherkin
Given a user has valid credentials
When they enter their email and password and click login
Then they should be authenticated with a secure JWT token
And they should be redirected to the main bookmark interface
And their session should be stored securely
```

**Scenario 3: Token expiration handling**
```gherkin
Given a user has an expired authentication token
When they try to perform any bookmark operation
Then they should be prompted to re-authenticate
And their pending operation should be queued for after login
```

**Scenario 4: User logout**
```gherkin
Given a user is logged in
When they click the logout button
Then their authentication token should be invalidated
And their local bookmark cache should be cleared
And they should be redirected to the login screen
```

**Scenario 5: Authentication failure handling**
```gherkin
Given a user enters invalid credentials
When they attempt to login
Then they should see a clear error message
And they should be offered password recovery options
And the system should not reveal whether the email exists
```

#### Implementation Status: ✅ COMPLETED
- ✅ Supabase Auth integration with JWT token validation
- ✅ User registration and login endpoints
- ✅ Session management with Redis storage
- ✅ Role-based access control (RBAC) middleware
- ✅ Password reset and account recovery workflows
- ✅ User profile management with preferences storage

### Requirement 2 - Core Bookmark Management

**User Story:** As a user, I want to create, edit, and delete bookmarks, so that I can manage my bookmark collection.

#### BDD Scenarios

**Scenario 1: Creating a new bookmark**
```gherkin
Given a user is logged in and on a webpage
When they click the "Save Bookmark" button in the extension
Then the system should capture the URL, title, and page description
And it should generate a favicon and screenshot
And the bookmark should be saved with a timestamp
And the user should see a success confirmation
```

**Scenario 2: Editing an existing bookmark**
```gherkin
Given a user has a saved bookmark
When they click edit and modify the title, description, or tags
Then the system should validate the changes
And update the bookmark with the new information
And preserve the original creation timestamp
And show the updated modification timestamp
```

**Scenario 3: Deleting a bookmark with recovery**
```gherkin
Given a user wants to delete a bookmark
When they click the delete button and confirm
Then the bookmark should be soft deleted (not permanently removed)
And it should be moved to a "Recently Deleted" section
And the user should be able to recover it within 30 days
And after 30 days it should be permanently deleted
```

**Scenario 4: Searching bookmarks**
```gherkin
Given a user has multiple bookmarks saved
When they enter a search term in the search box
Then the system should search across titles, descriptions, and URLs
And display results ranked by relevance
And highlight matching terms in the results
And show "no results found" if no matches exist
```

**Scenario 5: Filtering bookmarks by tags**
```gherkin
Given a user has bookmarks with various tags
When they select one or more tags from the filter menu
Then only bookmarks containing those tags should be displayed
And the filter should support multiple tag selection (AND/OR logic)
And they should be able to clear filters to see all bookmarks
```

**Scenario 6: Paginated bookmark listing**
```gherkin
Given a user has more than 50 bookmarks
When they view their bookmark list
Then bookmarks should be displayed in pages of 50 items
And they should see pagination controls at the bottom
And they should be able to sort by date, title, or URL
And the current page and total count should be visible
```

#### Implementation Status: ✅ COMPLETED
- ✅ Full CRUD operations (Create, Read, Update, Delete)
- ✅ URL format validation
- ✅ User authorization and isolation
- ✅ JSON-based tag storage and management
- ✅ Search functionality across multiple fields
- ✅ Pagination and sorting support
- ✅ Soft delete with recovery capability
- ✅ Comprehensive error handling and validation
- ✅ RESTful API endpoints with proper HTTP status codes

**Cross-References:**
- 📋 Implementation Task: [Task 6 in tasks.md](tasks.md#task-6)
- 🏗️ Technical Design: [Bookmark Management API in design.md](design.md#bookmark-management)
- 💻 Code Implementation: `backend/internal/bookmark/`
- 🧪 Test Coverage: `backend/internal/bookmark/service_test.go`

### Requirement 3 - Basic Collections and Organization

**User Story:** As a user, I want to organize my bookmarks in collections, so that I can group related bookmarks together.

#### BDD Scenarios

**Scenario 1: Creating a new collection**
```gherkin
Given a user wants to organize their bookmarks
When they click "Create Collection" and enter a name and description
Then a new collection should be created with the specified details
And it should appear in their collections list
And it should initially contain zero bookmarks
```

**Scenario 2: Adding bookmarks to a collection**
```gherkin
Given a user has bookmarks and collections
When they drag a bookmark onto a collection or use the "Add to Collection" menu
Then the bookmark should be associated with that collection
And the collection's bookmark count should increase
And the bookmark should appear when viewing the collection
```

**Scenario 3: Moving bookmarks between collections**
```gherkin
Given a bookmark exists in Collection A
When the user moves it to Collection B
Then the bookmark should be removed from Collection A
And added to Collection B
And both collections' bookmark counts should update accordingly
```

**Scenario 4: Viewing collection details**
```gherkin
Given a user has collections with bookmarks
When they view their collections list
Then each collection should show its name, description, and bookmark count
And they should be able to click to view the bookmarks within each collection
And collections should be sortable by name, date created, or bookmark count
```

**Scenario 5: Collection operation error handling**
```gherkin
Given a collection operation fails due to network or server issues
When the user attempts to create, modify, or delete a collection
Then the system should maintain data consistency
And show an appropriate error message
And allow the user to retry the operation
And not leave the collection in a corrupted state
```

#### Implementation Status: ⏳ PLANNED (Task 7)
- ⏳ Collection model with basic folder support
- ⏳ Collection CRUD operations with validation
- ⏳ Bookmark-to-collection associations
- ⏳ Basic collection sharing (public/private)
- ⏳ Collection listing and organization
- 🔗 Related to bookmark filtering by collection (partially implemented)

### Requirement 4 - Cross-Browser Synchronization

**User Story:** As a user, I want to sync my bookmarks across browsers, so that I have consistent access to my bookmarks.

#### BDD Scenarios

**Scenario 1: Real-time bookmark synchronization**
```gherkin
Given a user has the extension installed on Chrome and Firefox
When they save a bookmark in Chrome
Then the bookmark should appear in Firefox within 60 seconds
And both browsers should show the same bookmark data
And the sync should happen automatically without user intervention
```

**Scenario 2: Bookmark modification sync**
```gherkin
Given a user has the same bookmark on multiple devices
When they edit the bookmark title on Device A
Then Device B should receive the updated title within 60 seconds
And the modification timestamp should be updated on both devices
And no data should be lost during the sync process
```

**Scenario 3: Sync conflict resolution**
```gherkin
Given the same bookmark is modified on two devices while offline
When both devices come back online simultaneously
Then the system should detect the conflict
And resolve it using the most recent timestamp
And notify the user about the conflict resolution
And preserve both versions in the sync history
```

**Scenario 4: Offline sync queuing**
```gherkin
Given a user is working offline
When they create, edit, or delete bookmarks
Then the changes should be queued locally
And when internet connection is restored
Then all queued changes should sync automatically
And the user should see sync progress indicators
```

**Scenario 5: Sync failure recovery**
```gherkin
Given a sync operation fails due to server issues
When the system detects the failure
Then it should retry with exponential backoff (1s, 2s, 4s, 8s...)
And show the user the current sync status
And continue retrying until successful or maximum attempts reached
And allow manual retry if automatic retry fails
```

### Requirement 5 - Browser Extension Interface

**User Story:** As a user, I want browser extensions with basic bookmark management, so that I can access my bookmarks from any browser.

#### BDD Scenarios

**Scenario 1: Chrome extension installation and setup**
```gherkin
Given a user installs the Chrome extension
When they click the extension icon for the first time
Then they should see a welcome screen with login options
And after authentication, they should access bookmark management features
And the extension should integrate with Chrome's bookmark system
```

**Scenario 2: Firefox extension functionality**
```gherkin
Given a user installs the Firefox extension
When they use the extension features
Then they should have the same functionality as the Chrome version
And bookmarks should sync between Chrome and Firefox
And the UI should adapt to Firefox's design guidelines
```

**Scenario 3: Extension popup interface**
```gherkin
Given a user opens the extension popup
When they view their bookmarks
Then bookmarks should be displayed in a grid or list view
And they should be able to toggle between view modes
And search and filter options should be easily accessible
And the interface should be responsive and fast
```

**Scenario 4: Bookmark navigation**
```gherkin
Given a user sees bookmarks in the extension popup
When they click on a bookmark
Then it should open in a new tab
And the original tab should remain active
And the click should be tracked for usage analytics (if enabled)
```

**Scenario 5: Authentication requirement**
```gherkin
Given a user is not logged in
When they try to access bookmark features
Then they should be redirected to the login screen
And bookmark data should not be accessible
And they should see a clear message about authentication requirement
```

### 🟡 Phase 2: Enhanced Features

### Requirement 6 - Visual Grid Interface and Content Previews

**User Story:** As a user, I want a visual grid interface with previews, so that I can quickly identify and access my bookmarks.

#### BDD Scenarios

**Scenario 1: Visual grid bookmark display**
```gherkin
Given a user has saved bookmarks
When they view their bookmark collection
Then bookmarks should be displayed in a visual grid layout
And each bookmark should show a thumbnail, title, and URL
And the grid should be responsive to different screen sizes
```

**Scenario 2: Automatic screenshot capture**
```gherkin
Given a user saves a new bookmark
When the bookmark is being processed
Then the system should automatically capture a screenshot of the webpage
And store it in MinIO storage
And display the screenshot as the bookmark thumbnail
And the process should complete within 10 seconds
```

**Scenario 3: Bookmark hover interactions**
```gherkin
Given a user is viewing bookmarks in grid mode
When they hover over a bookmark thumbnail
Then additional information should appear (description, tags, date saved)
And the thumbnail should have a subtle hover effect
And action buttons (edit, delete, share) should become visible
```

**Scenario 4: Grid layout customization**
```gherkin
Given a user wants to customize their bookmark view
When they adjust grid size, spacing, or layout options
Then their preferences should be saved automatically
And applied consistently across all devices
And the changes should take effect immediately
```

**Scenario 5: Screenshot fallback handling**
```gherkin
Given a webpage screenshot cannot be captured (due to restrictions or errors)
When the bookmark is saved
Then the system should use the webpage's favicon as the thumbnail
And if no favicon is available, use a default placeholder image
And indicate to the user that screenshot capture failed
```

### Requirement 7 - Search and Discovery

**User Story:** As a user, I want to search my bookmarks effectively, so that I can find specific content quickly.

#### BDD Scenarios

**Scenario 1: Basic bookmark search**
```gherkin
Given a user has bookmarks with various titles and descriptions
When they enter a search term in the search box
Then the system should search across titles, URLs, and descriptions
And matching should be case-insensitive
And results should be displayed in real-time as they type
```

**Scenario 2: Search result ranking and sorting**
```gherkin
Given search results are returned
When the user views the results
Then they should be ranked by relevance (title matches first, then description, then URL)
And the user should be able to sort by date, title, or relevance
And matching terms should be highlighted in the results
```

**Scenario 3: Chinese language search support**
```gherkin
Given a user searches using Chinese characters (Traditional or Simplified)
When they enter Chinese text in the search box
Then the system should properly tokenize and search Chinese content
And support both Traditional and Simplified Chinese characters
And provide accurate results for Chinese bookmarks
```

**Scenario 4: No results found handling**
```gherkin
Given a user searches for a term that doesn't match any bookmarks
When the search completes
Then they should see a "No results found" message
And suggestions for refining their search
And the option to clear the search and view all bookmarks
```

**Scenario 5: Search service fallback**
```gherkin
Given the advanced search service (Typesense) is unavailable
When a user performs a search
Then the system should fall back to basic text matching
And inform the user that advanced search features are temporarily unavailable
And still provide functional search results using the fallback method
```

#### Implementation Status: 🟡 PARTIALLY COMPLETED
- ✅ Basic search across title, description, and URL
- ✅ Case-insensitive search functionality
- ✅ Search result pagination and sorting
- ⏳ Advanced search with Typesense (planned for Phase 7)
- ⏳ Chinese language support (planned for Phase 7)
- ⏳ Search suggestions and auto-complete (planned for Phase 9)

### Requirement 8 - Import/Export and Data Migration

**User Story:** As a user, I want to import existing bookmarks and export my data, so that I can migrate from other services.

#### Acceptance Criteria

1. WHEN importing bookmarks THEN the system SHALL support Chrome, Firefox, Safari formats
2. WHEN importing data THEN the system SHALL preserve folder structure and handle duplicates
3. WHEN exporting bookmarks THEN the system SHALL provide JSON and HTML formats
4. WHEN processing large imports THEN the system SHALL show progress indicators
5. IF import fails THEN the system SHALL provide detailed error information

### Requirement 9 - Offline Support

**User Story:** As a user, I want to access my bookmarks offline, so that I can work without internet connectivity.

#### Acceptance Criteria

1. WHEN the extension loads THEN it SHALL cache recent bookmarks locally
2. WHEN working offline THEN the system SHALL queue changes and show offline indicators
3. WHEN connectivity is restored THEN the system SHALL sync queued changes
4. WHEN offline storage is full THEN the system SHALL manage cache efficiently
5. IF sync conflicts occur THEN the system SHALL resolve using timestamp priority

### 🟢 Phase 3: Advanced Features

### Requirement 10 - Content Analysis and Smart Features

**User Story:** As a user, I want intelligent content analysis and suggestions, so that I can better organize my bookmarks.

#### Acceptance Criteria

1. WHEN a bookmark is saved THEN the system SHALL extract and analyze content metadata
2. WHEN content analysis completes THEN the system SHALL suggest relevant tags and categories
3. WHEN organizing bookmarks THEN the system SHALL recommend optimal folder structures
4. WHEN duplicates are detected THEN the system SHALL suggest merging options
5. IF content analysis fails THEN the system SHALL use basic URL and title information

### Requirement 11 - Basic Sharing and Collaboration

**User Story:** As a user, I want to share bookmark collections, so that I can collaborate with others.

#### Acceptance Criteria

1. WHEN a user creates a collection THEN they SHALL be able to set it as public or private
2. WHEN sharing collections THEN the system SHALL generate shareable links
3. WHEN viewing shared collections THEN users SHALL see owner information and permissions
4. WHEN copying shared collections THEN users SHALL be able to add them to their library
5. IF sharing is disabled THEN all collections SHALL remain private by default

### 🔵 Phase 4: Community and Social Features

### Requirement 12 - Community Discovery and Social Features

**User Story:** As a user, I want to discover popular bookmarks and engage with the community, so that I can find valuable content.

#### Acceptance Criteria

1. WHEN browsing community content THEN users SHALL see popular and trending bookmarks
2. WHEN viewing bookmarks THEN users SHALL see basic social metrics (save count)
3. WHEN users opt-in THEN they SHALL be able to follow other users
4. WHEN privacy is preferred THEN users SHALL be able to opt-out of all social features
5. IF community features are disabled THEN personal functionality SHALL remain unaffected

### Requirement 13 - Advanced Customization and Analytics

**User Story:** As a user, I want advanced customization and usage insights, so that I can optimize my bookmark management.

#### Acceptance Criteria

1. WHEN accessing settings THEN users SHALL have theme and layout customization options
2. WHEN viewing analytics THEN users SHALL see usage patterns and collection statistics
3. WHEN customizing interface THEN changes SHALL sync across all browser extensions
4. WHEN privacy is important THEN users SHALL be able to disable analytics tracking
5. IF analytics are disabled THEN basic functionality SHALL remain available

### 🟣 Phase 5: Enterprise and Advanced Features

### Requirement 14 - Link Monitoring and Maintenance

**User Story:** As a user, I want automatic link monitoring, so that I can maintain a healthy bookmark collection.

#### Acceptance Criteria

1. WHEN the system checks links THEN it SHALL detect broken URLs and redirects
2. WHEN issues are found THEN the system SHALL notify users with suggested fixes
3. WHEN maintenance is needed THEN the system SHALL provide cleanup suggestions
4. WHEN users review suggestions THEN bulk actions SHALL be available
5. IF monitoring is disabled THEN link status SHALL still be tracked passively

### Requirement 15 - Advanced Automation and Integration

**User Story:** As a user, I want automation features and external integrations, so that I can streamline my workflow.

#### Acceptance Criteria

1. WHEN setting up automation THEN users SHALL be able to create rules and triggers
2. WHEN integrating external services THEN the system SHALL support webhooks and APIs
3. WHEN generating feeds THEN the system SHALL provide RSS/Atom feeds for public collections
4. WHEN using bulk operations THEN the system SHALL support batch processing
5. IF automation fails THEN the system SHALL provide detailed error reporting

### Requirement 17 - Self-Hosted Deployment and Administration

**User Story:** As an administrator, I want to self-host the bookmark service with comprehensive management tools, so that I have full control over data and system operation.

#### Acceptance Criteria

1. WHEN deploying the service THEN the system SHALL provide Docker containers for self-hosted Supabase stack (PostgreSQL, Auth, Realtime, REST API), Redis, Typesense, MinIO, and Go backend services with configuration guides and initialization scripts
2. WHEN the service starts THEN the system SHALL initialize Supabase PostgreSQL database, MinIO storage, search indexes, API endpoints, and provide health monitoring for all Supabase components
3. WHEN managing the system THEN administrators SHALL have access to user management through Supabase Auth, system monitoring, MinIO storage management, and backup tools for Supabase PostgreSQL
4. WHEN errors occur THEN the system SHALL provide detailed logging, error reporting, and recovery procedures for both Go services and Supabase components
5. IF scaling is needed THEN the system SHALL support horizontal scaling with MinIO distributed storage, Supabase read replicas, and load balancing

### Requirement 18 - User and Permission Management

**User Story:** As an administrator, I want comprehensive user and permission management, so that I can control access, set quotas, and maintain system security.

#### Acceptance Criteria

1. WHEN managing users THEN administrators SHALL create, modify, suspend, and delete user accounts with proper data handling
2. WHEN setting permissions THEN the system SHALL support user groups, role-based access, and custom permission sets
3. WHEN enforcing quotas THEN the system SHALL limit collections, bookmarks, and storage per user/group with warnings
4. WHEN users exceed limits THEN the system SHALL prevent further actions and provide cleanup suggestions
5. IF administrative changes are made THEN the system SHALL apply them immediately and log all administrative actions

### Requirement 19 - Data Protection and Backup

**User Story:** As a user and administrator, I want comprehensive data protection and backup capabilities, so that bookmark data is safe and recoverable.

#### Acceptance Criteria

1. WHEN bookmarks are modified THEN the system SHALL automatically create versioned backups stored in MinIO with configurable retention
2. WHEN restoration is needed THEN users SHALL access backup versions with selective restoration options from MinIO storage
3. WHEN system backups run THEN administrators SHALL have automated backup scheduling for Supabase PostgreSQL database and MinIO file storage with monitoring
4. WHEN backup storage limits are reached THEN the system SHALL archive older versions to MinIO according to policies
5. IF data corruption occurs THEN the system SHALL provide multiple recovery options from MinIO backups and data integrity verification