// Sync manager for Chrome extension
import { apiClient } from '../../shared/api-client.js';
import { getDeviceId } from '../../shared/utils.js';
import { STORAGE_KEYS, MESSAGE_TYPES, SYNC_EVENTS } from '../../shared/constants.js';

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
    this.deviceId = await getDeviceId();
    await this.loadCachedBookmarks();

    // Start WebSocket connection if authenticated
    const stored = await chrome.storage.local.get([STORAGE_KEYS.AUTH_TOKEN]);
    if (stored[STORAGE_KEYS.AUTH_TOKEN]) {
      await this.connect();
    }
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
      case MESSAGE_TYPES.PONG:
        // Heartbeat response
        break;

      case MESSAGE_TYPES.SYNC_RESPONSE:
        this.handleSyncResponse(message.data);
        break;

      case MESSAGE_TYPES.SYNC_EVENT:
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
    chrome.runtime.sendMessage({
      type: 'SYNC_COMPLETED',
      eventsCount: data.events.length
    });
  }

  /**
   * Handle individual sync event
   */
  async handleSyncEvent(event) {
    console.log('Received sync event:', event);
    await this.applySyncEvent(event);
    await this.refreshBookmarksCache();

    // Notify popup of changes
    chrome.runtime.sendMessage({
      type: 'SYNC_EVENT_RECEIVED',
      event: event
    });
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
    // For extensions, we don't need to create local bookmarks
    // The API handles the creation, we just need to refresh our cache
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
      type: MESSAGE_TYPES.SYNC_REQUEST,
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
          type: MESSAGE_TYPES.SYNC_EVENT,
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

      await chrome.storage.local.set({
        [STORAGE_KEYS.SYNC_STATE]: {
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
        await chrome.storage.local.set({
          [STORAGE_KEYS.BOOKMARKS_CACHE]: {
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
    const stored = await chrome.storage.local.get([STORAGE_KEYS.BOOKMARKS_CACHE]);
    return stored[STORAGE_KEYS.BOOKMARKS_CACHE] || { bookmarks: [], total: 0 };
  }

  /**
   * Start heartbeat to keep connection alive
   */
  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        const message = {
          type: MESSAGE_TYPES.PING,
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

// Export singleton instance
export const syncManager = new SyncManager();