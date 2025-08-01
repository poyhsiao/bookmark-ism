// Safari Extension Sync Manager
// Handles real-time synchronization for Safari Web Extension

import { apiClient } from '../../shared/api-client.js';
import { STORAGE_KEYS, MESSAGE_TYPES, SYNC_EVENTS } from '../../shared/constants.js';
import { getDeviceId } from '../../shared/utils.js';

class SafariSyncManager {
  constructor() {
    this.websocket = null;
    this.isConnected = false;
    this.deviceId = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.pingInterval = null;
    this.syncQueue = [];
  }

  /**
   * Initialize sync manager
   */
  async init() {
    try {
      this.deviceId = await getDeviceId();

      // Load sync queue from storage
      const stored = await browser.storage.local.get([STORAGE_KEYS.SYNC_STATE]);
      if (stored[STORAGE_KEYS.SYNC_STATE]) {
        const syncState = JSON.parse(stored[STORAGE_KEYS.SYNC_STATE]);
        this.syncQueue = syncState.queue || [];
      }

      console.log('Safari Sync Manager: Initialized with device ID:', this.deviceId);
    } catch (error) {
      console.error('Safari Sync Manager: Failed to initialize:', error);
    }
  }

  /**
   * Connect to WebSocket server
   */
  async connect() {
    if (this.isConnected || !this.deviceId) {
      return;
    }

    try {
      const wsUrl = `ws://localhost:8080/api/v1/sync/ws?device_id=${this.deviceId}`;
      this.websocket = new WebSocket(wsUrl);

      this.websocket.addEventListener('open', () => {
        console.log('Safari Sync Manager: WebSocket connected');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.startPingInterval();
        this.processSyncQueue();
      });

      this.websocket.addEventListener('message', (event) => {
        this.handleMessage(event.data);
      });

      this.websocket.addEventListener('close', () => {
        console.log('Safari Sync Manager: WebSocket disconnected');
        this.isConnected = false;
        this.stopPingInterval();
        this.scheduleReconnect();
      });

      this.websocket.addEventListener('error', (error) => {
        console.error('Safari Sync Manager: WebSocket error:', error);
        this.isConnected = false;
      });

    } catch (error) {
      console.error('Safari Sync Manager: Failed to connect:', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect() {
    if (this.websocket) {
      this.websocket.close();
      this.websocket = null;
    }
    this.isConnected = false;
    this.stopPingInterval();
  }

  /**
   * Handle incoming WebSocket messages
   */
  handleMessage(data) {
    try {
      const message = JSON.parse(data);

      switch (message.type) {
        case MESSAGE_TYPES.PONG:
          // Pong received, connection is alive
          break;

        case MESSAGE_TYPES.SYNC_EVENT:
          this.handleSyncEvent(message.data);
          break;

        case MESSAGE_TYPES.SYNC_RESPONSE:
          this.handleSyncResponse(message.data);
          break;

        default:
          console.warn('Safari Sync Manager: Unknown message type:', message.type);
      }
    } catch (error) {
      console.error('Safari Sync Manager: Failed to handle message:', error);
    }
  }

  /**
   * Handle sync events from server
   */
  async handleSyncEvent(eventData) {
    try {
      console.log('Safari Sync Manager: Received sync event:', eventData);

      // Don't process events from this device
      if (eventData.device_id === this.deviceId) {
        return;
      }

      // Update local cache based on event type
      switch (eventData.event_type) {
        case 'create':
          if (eventData.resource_type === 'bookmark') {
            await this.handleBookmarkCreated(eventData);
          }
          break;

        case 'update':
          if (eventData.resource_type === 'bookmark') {
            await this.handleBookmarkUpdated(eventData);
          }
          break;

        case 'delete':
          if (eventData.resource_type === 'bookmark') {
            await this.handleBookmarkDeleted(eventData);
          }
          break;
      }

      // Notify popup/options page of changes
      this.notifyExtensionPages('SYNC_UPDATE', eventData);

    } catch (error) {
      console.error('Safari Sync Manager: Failed to handle sync event:', error);
    }
  }

  /**
   * Handle bookmark created event
   */
  async handleBookmarkCreated(eventData) {
    const storageManager = (await import('./storage-manager.js')).storageManager;
    await storageManager.addBookmarkToCache(eventData.changes);
  }

  /**
   * Handle bookmark updated event
   */
  async handleBookmarkUpdated(eventData) {
    const storageManager = (await import('./storage-manager.js')).storageManager;
    await storageManager.updateBookmarkInCache(eventData.resource_id, eventData.changes);
  }

  /**
   * Handle bookmark deleted event
   */
  async handleBookmarkDeleted(eventData) {
    const storageManager = (await import('./storage-manager.js')).storageManager;
    await storageManager.removeBookmarkFromCache(eventData.resource_id);
  }

  /**
   * Create sync event
   */
  async createSyncEvent(eventType, resourceType, resourceId, changes = {}) {
    const event = {
      device_id: this.deviceId,
      event_type: eventType,
      resource_type: resourceType,
      resource_id: resourceId,
      changes: changes,
      timestamp: new Date().toISOString()
    };

    if (this.isConnected) {
      try {
        // Send immediately via WebSocket
        this.websocket.send(JSON.stringify({
          type: MESSAGE_TYPES.SYNC_EVENT,
          data: event
        }));

        console.log('Safari Sync Manager: Sent sync event:', event);
      } catch (error) {
        console.error('Safari Sync Manager: Failed to send sync event:', error);
        this.queueSyncEvent(event);
      }
    } else {
      // Queue for later
      this.queueSyncEvent(event);
    }
  }

  /**
   * Queue sync event for later processing
   */
  async queueSyncEvent(event) {
    this.syncQueue.push(event);

    // Persist queue to storage
    await browser.storage.local.set({
      [STORAGE_KEYS.SYNC_STATE]: JSON.stringify({
        queue: this.syncQueue,
        last_sync: new Date().toISOString()
      })
    });

    console.log('Safari Sync Manager: Queued sync event:', event);
  }

  /**
   * Process queued sync events
   */
  async processSyncQueue() {
    if (!this.isConnected || this.syncQueue.length === 0) {
      return;
    }

    console.log('Safari Sync Manager: Processing sync queue:', this.syncQueue.length, 'events');

    const eventsToProcess = [...this.syncQueue];
    this.syncQueue = [];

    for (const event of eventsToProcess) {
      try {
        this.websocket.send(JSON.stringify({
          type: MESSAGE_TYPES.SYNC_EVENT,
          data: event
        }));
      } catch (error) {
        console.error('Safari Sync Manager: Failed to process queued event:', error);
        this.syncQueue.push(event); // Re-queue failed event
      }
    }

    // Update storage
    await browser.storage.local.set({
      [STORAGE_KEYS.SYNC_STATE]: JSON.stringify({
        queue: this.syncQueue,
        last_sync: new Date().toISOString()
      })
    });
  }

  /**
   * Request full sync
   */
  requestSync() {
    if (!this.isConnected) {
      console.warn('Safari Sync Manager: Cannot request sync - not connected');
      return;
    }

    this.websocket.send(JSON.stringify({
      type: MESSAGE_TYPES.SYNC_REQUEST,
      data: {
        device_id: this.deviceId,
        timestamp: new Date().toISOString()
      }
    }));

    console.log('Safari Sync Manager: Requested full sync');
  }

  /**
   * Start ping interval to keep connection alive
   */
  startPingInterval() {
    this.pingInterval = setInterval(() => {
      if (this.isConnected && this.websocket) {
        this.websocket.send(JSON.stringify({
          type: MESSAGE_TYPES.PING,
          data: { timestamp: new Date().toISOString() }
        }));
      }
    }, 30000); // Ping every 30 seconds
  }

  /**
   * Stop ping interval
   */
  stopPingInterval() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  /**
   * Schedule reconnection attempt
   */
  scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Safari Sync Manager: Max reconnection attempts reached');
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;

    console.log(`Safari Sync Manager: Scheduling reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`);

    setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Notify extension pages of sync updates
   */
  notifyExtensionPages(type, data) {
    // Safari doesn't have chrome.runtime.sendMessage to all pages
    // We'll use storage events instead
    browser.storage.local.set({
      sync_notification: JSON.stringify({
        type,
        data,
        timestamp: Date.now()
      })
    });
  }

  /**
   * Get connection status
   */
  getConnectionStatus() {
    return {
      connected: this.isConnected,
      device_id: this.deviceId,
      queue_length: this.syncQueue.length,
      reconnect_attempts: this.reconnectAttempts
    };
  }
}

// Export singleton instance
export const syncManager = new SafariSyncManager();