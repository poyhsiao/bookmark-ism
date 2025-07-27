// Sync manager for Firefox extension
// Adapted from Chrome version for Firefox compatibility

class SyncManager {
  constructor() {
    this.websocket = null;
    this.deviceId = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000; // Start with 1 second
    this.syncInProgress = false;
  }

  /**
   * Initialize sync manager
   */
  async init() {
    this.deviceId = await this.getDeviceId();
    await this.loadCachedBookmarks();

    // Start WebSocket connection if authenticated
    const stored = await browser.storage.local.get(['auth_access_token']);
    if (stored.auth_access_token) {
      await this.connect();
    }
  }

  /**
   * Get or create device ID
   */
  async getDeviceId() {
    const stored = await browser.storage.local.get(['device_id']);
    if (stored.device_id) {
      return stored.device_id;
    }

    const deviceId = 'ff_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
    await browser.storage.local.set({ device_id: deviceId });
    return deviceId;
  }

  /**
   * Connect to WebSocket server
   */
  async connect() {
    if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      const wsUrl = `ws://localhost:8080/api/v1/sync/ws`;
      this.websocket = new WebSocket(wsUrl);

      this.websocket.onopen = () => {
        console.log('WebSocket connected');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;

        // Send initial sync request
        this.requestSync();

        // Start heartbeat
        this.startHeartbeat();
      };

      this.websocket.onmessage = (event) => {
        this.handleMessage(JSON.parse(event.data));
      };

      this.websocket.onclose = () => {
        console.log('WebSocket disconnected');
        this.isConnected = false;
        this.stopHeartbeat();
        this.scheduleReconnect();
      };

      this.websocket.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Disconnect WebSocket
   */
  disconnect() {
    if (this.websocket) {
      this.websocket.close();
      this.websocket = null;
    }
    this.isConnected = false;
    this.stopHeartbeat();
  }

  /**
   * Handle WebSocket messages
   */
  handleMessage(message) {
    switch (message.type) {
      case 'pong':
        // Heartbeat response
        break;

      case 'sync_response':
        this.handleSyncResponse(message.data);
        break;

      case 'sync_event':
        this.handleSyncEvent(message.data);
        break;

      default:
        console.warn('Unknown message type:', message.type);
    }
  }

  /**
   * Handle sync response with delta changes
   */
  async handleSyncResponse(data) {
    if (!data.events || data.events.length === 0) {
      return;
    }

    console.log(`Received ${data.events.length} sync events`);

    for (const event of data.events) {
      await this.applySyncEvent(event);
    }

    // Update sync state
    await this.updateSyncState(data.last_sync_timestamp);

    // Update cached bookmarks
    await this.refreshBookmarksCache();

    // Notify popup of changes
    browser.runtime.sendMessage({
      type: 'SYNC_COMPLETED',
      eventsCount: data.events.length
    }).catch(() => {}); // Ignore if no listeners
  }

  /**
   * Handle individual sync event
   */
  async handleSyncEvent(event) {
    console.log('Received sync event:', event);
    await this.applySyncEvent(event);
    await this.refreshBookmarksCache();

    // Notify popup of changes
    browser.runtime.sendMessage({
      type: 'SYNC_EVENT_RECEIVED',
      event: event
    }).catch(() => {}); // Ignore if no listeners
  }

  /**
   * Apply sync event to local state
   */
  async applySyncEvent(event) {
    // Skip events from this device
    if (event.device_id === this.deviceId) {
      return;
    }

    try {
      switch (event.event_type) {
        case 'create':
          await this.handleCreateEvent(event);
          break;
        case 'update':
          await this.handleUpdateEvent(event);
          break;
        case 'delete':
          await this.handleDeleteEvent(event);
          break;
        default:
          console.warn('Unknown event type:', event.event_type);
      }
    } catch (error) {
      console.error('Failed to apply sync event:', error);
    }
  }

  /**
   * Handle create events
   */
  async handleCreateEvent(event) {
    console.log('Handling create event for:', event.resource_type, event.resource_id);
  }

  /**
   * Handle update events
   */
  async handleUpdateEvent(event) {
    console.log('Handling update event for:', event.resource_type, event.resource_id);
  }

  /**
   * Handle delete events
   */
  async handleDeleteEvent(event) {
    console.log('Handling delete event for:', event.resource_type, event.resource_id);
  }

  /**
   * Request sync from server
   */
  requestSync() {
    if (!this.isConnected || this.syncInProgress) {
      return;
    }

    this.syncInProgress = true;

    const message = {
      type: 'sync_request',
      data: {
        device_id: this.deviceId,
        timestamp: new Date().toISOString()
      }
    };

    this.websocket.send(JSON.stringify(message));

    // Reset sync flag after timeout
    setTimeout(() => {
      this.syncInProgress = false;
    }, 5000);
  }

  /**
   * Create sync event for local changes
   */
  async createSyncEvent(eventType, resourceType, resourceId, changes = {}) {
    try {
      const eventData = {
        device_id: this.deviceId,
        event_type: eventType,
        resource_type: resourceType,
        resource_id: resourceId,
        changes: changes,
        timestamp: new Date().toISOString()
      };

      await apiClient.createSyncEvent(eventData);

      // Also send via WebSocket if connected
      if (this.isConnected) {
        const message = {
          type: 'sync_event',
          data: eventData
        };
        this.websocket.send(JSON.stringify(message));
      }

    } catch (error) {
      console.error('Failed to create sync event:', error);
    }
  }

  /**
   * Update sync state
   */
  async updateSyncState(lastSyncTimestamp) {
    try {
      await apiClient.updateSyncState(this.deviceId, lastSyncTimestamp);

      await browser.storage.local.set({
        sync_state: {
          device_id: this.deviceId,
          last_sync_timestamp: lastSyncTimestamp,
          updated_at: new Date().toISOString()
        }
      });
    } catch (error) {
      console.error('Failed to update sync state:', error);
    }
  }

  /**
   * Refresh bookmarks cache
   */
  async refreshBookmarksCache() {
    try {
      const response = await apiClient.getBookmarks({ limit: 100 });

      if (response.success && response.data) {
        await browser.storage.local.set({
          bookmarks_cache: {
            bookmarks: response.data,
            total: response.total,
            cached_at: new Date().toISOString()
          }
        });
      }
    } catch (error) {
      console.error('Failed to refresh bookmarks cache:', error);
    }
  }

  /**
   * Load cached bookmarks
   */
  async loadCachedBookmarks() {
    const stored = await browser.storage.local.get(['bookmarks_cache']);
    return stored.bookmarks_cache || { bookmarks: [], total: 0 };
  }

  /**
   * Start heartbeat to keep connection alive
   */
  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        const message = {
          type: 'ping',
          data: { timestamp: new Date().toISOString() }
        };
        this.websocket.send(JSON.stringify(message));
      }
    }, 30000); // Send ping every 30 seconds
  }

  /**
   * Stop heartbeat
   */
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Scheduling reconnection attempt ${this.reconnectAttempts} in ${delay}ms`);

    setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Get connection status
   */
  getConnectionStatus() {
    return {
      connected: this.isConnected,
      deviceId: this.deviceId,
      reconnectAttempts: this.reconnectAttempts
    };
  }
}

// Create global instance for Firefox
const syncManager = new SyncManager();