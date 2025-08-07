// API client for browser extensions
import { API_CONFIG, STORAGE_KEYS, ERROR_CODES } from './constants.js';
import { handleApiError } from './utils.js';

class ApiClient {
  constructor() {
    this.baseUrl = API_CONFIG.BASE_URL;
    this.token = null;
  }

  /**
   * Initialize API client with stored token
   */
  async init() {
    const stored = await chrome.storage.local.get([STORAGE_KEYS.AUTH_TOKEN]);
    this.token = stored[STORAGE_KEYS.AUTH_TOKEN];
  }

  /**
   * Set authentication token
   */
  setToken(token) {
    this.token = token;
  }

  /**
   * Get authentication headers
   */
  getHeaders() {
    const headers = {
      'Content-Type': 'application/json'
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    return headers;
  }

  /**
   * Make HTTP request with error handling
   */
  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const config = {
      headers: this.getHeaders(),
      ...options
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      }

      return await response.text();
    } catch (error) {
      console.error('API Request failed:', error);
      throw handleApiError(error);
    }
  }

  /**
   * Authentication methods
   */
  async login(email, password) {
    const response = await this.request(`${API_CONFIG.ENDPOINTS.AUTH}/login`, {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });

    if (response.token) {
      this.setToken(response.token);
      await chrome.storage.local.set({
        [STORAGE_KEYS.AUTH_TOKEN]: response.token,
        [STORAGE_KEYS.USER_DATA]: response.user
      });
    }

    return response;
  }

  async register(email, password, name) {
    const response = await this.request(`${API_CONFIG.ENDPOINTS.AUTH}/register`, {
      method: 'POST',
      body: JSON.stringify({ email, password, name })
    });

    if (response.token) {
      this.setToken(response.token);
      await chrome.storage.local.set({
        [STORAGE_KEYS.AUTH_TOKEN]: response.token,
        [STORAGE_KEYS.USER_DATA]: response.user
      });
    }

    return response;
  }

  async logout() {
    try {
      await this.request(`${API_CONFIG.ENDPOINTS.AUTH}/logout`, {
        method: 'POST'
      });
    } catch (error) {
      console.warn('Logout request failed:', error);
    }

    // Clear local storage regardless of API response
    await chrome.storage.local.remove([
      STORAGE_KEYS.AUTH_TOKEN,
      STORAGE_KEYS.REFRESH_TOKEN,
      STORAGE_KEYS.USER_DATA,
      STORAGE_KEYS.BOOKMARKS_CACHE
    ]);

    this.token = null;
  }

  /**
   * Bookmark methods
   */
  async getBookmarks(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ?
      `${API_CONFIG.ENDPOINTS.BOOKMARKS}?${queryString}` :
      API_CONFIG.ENDPOINTS.BOOKMARKS;

    return await this.request(endpoint);
  }

  async createBookmark(bookmarkData) {
    return await this.request(API_CONFIG.ENDPOINTS.BOOKMARKS, {
      method: 'POST',
      body: JSON.stringify(bookmarkData)
    });
  }

  async updateBookmark(id, bookmarkData) {
    return await this.request(`${API_CONFIG.ENDPOINTS.BOOKMARKS}/${id}`, {
      method: 'PUT',
      body: JSON.stringify(bookmarkData)
    });
  }

  async deleteBookmark(id) {
    return await this.request(`${API_CONFIG.ENDPOINTS.BOOKMARKS}/${id}`, {
      method: 'DELETE'
    });
  }

  /**
   * Collection methods
   */
  async getCollections(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ?
      `${API_CONFIG.ENDPOINTS.COLLECTIONS}?${queryString}` :
      API_CONFIG.ENDPOINTS.COLLECTIONS;

    return await this.request(endpoint);
  }

  async createCollection(collectionData) {
    return await this.request(API_CONFIG.ENDPOINTS.COLLECTIONS, {
      method: 'POST',
      body: JSON.stringify(collectionData)
    });
  }

  async updateCollection(id, collectionData) {
    return await this.request(`${API_CONFIG.ENDPOINTS.COLLECTIONS}/${id}`, {
      method: 'PUT',
      body: JSON.stringify(collectionData)
    });
  }

  async deleteCollection(id) {
    return await this.request(`${API_CONFIG.ENDPOINTS.COLLECTIONS}/${id}`, {
      method: 'DELETE'
    });
  }

  /**
   * Sync methods
   */
  async getSyncState(deviceId) {
    return await this.request(`${API_CONFIG.ENDPOINTS.SYNC}/state?device_id=${deviceId}`);
  }

  async updateSyncState(deviceId, lastSyncTimestamp) {
    return await this.request(`${API_CONFIG.ENDPOINTS.SYNC}/state`, {
      method: 'PUT',
      body: JSON.stringify({
        device_id: deviceId,
        last_sync_timestamp: lastSyncTimestamp
      })
    });
  }

  async getDeltaSync(deviceId, since) {
    const params = new URLSearchParams({ device_id: deviceId, since });
    return await this.request(`${API_CONFIG.ENDPOINTS.SYNC}/delta?${params}`);
  }

  async createSyncEvent(eventData) {
    return await this.request(`${API_CONFIG.ENDPOINTS.SYNC}/events`, {
      method: 'POST',
      body: JSON.stringify(eventData)
    });
  }
}

// Export singleton instance
export const apiClient = new ApiClient();