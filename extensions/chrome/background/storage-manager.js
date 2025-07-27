// Storage manager for Chrome extension
import { STORAGE_KEYS } from '../../shared/constants.js';

class StorageManager {
  constructor() {
    this.cache = new Map();
    this.maxCacheSize = 1000; // Maximum number of cached items
  }

  /**
   * Initialize storage manager
   */
  async init() {
    // Load frequently accessed data into memory cache
    await this.loadCache();
  }

  /**
   * Load cache from storage
   */
  async loadCache() {
    try {
      const stored = await chrome.storage.local.get([
        STORAGE_KEYS.BOOKMARKS_CACHE,
        STORAGE_KEYS.USER_DATA,
        STORAGE_KEYS.SYNC_STATE
      ]);

      // Cache bookmarks
      if (stored[STORAGE_KEYS.BOOKMARKS_CACHE]) {
        this.cache.set('bookmarks', stored[STORAGE_KEYS.BOOKMARKS_CACHE]);
      }

      // Cache user data
      if (stored[STORAGE_KEYS.USER_DATA]) {
        this.cache.set('user', stored[STORAGE_KEYS.USER_DATA]);
      }

      // Cache sync state
      if (stored[STORAGE_KEYS.SYNC_STATE]) {
        this.cache.set('syncState', stored[STORAGE_KEYS.SYNC_STATE]);
      }

    } catch (error) {
      console.error('Failed to load cache:', error);
    }
  }

  /**
   * Get data from cache or storage
   */
  async get(key) {
    // Check memory cache first
    if (this.cache.has(key)) {
      return this.cache.get(key);
    }

    // Fallback to chrome storage
    try {
      const stored = await chrome.storage.local.get([key]);
      const value = stored[key];

      // Cache the value if it exists
      if (value !== undefined) {
        this.set(key, value, false); // Don't persist to storage again
      }

      return value;
    } catch (error) {
      console.error('Failed to get from storage:', error);
      return undefined;
    }
  }

  /**
   * Set data in cache and storage
   */
  async set(key, value, persist = true) {
    // Update memory cache
    this.cache.set(key, value);

    // Manage cache size
    if (this.cache.size > this.maxCacheSize) {
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }

    // Persist to chrome storage if requested
    if (persist) {
      try {
        await chrome.storage.local.set({ [key]: value });
      } catch (error) {
        console.error('Failed to set in storage:', error);
      }
    }
  }

  /**
   * Remove data from cache and storage
   */
  async remove(key) {
    // Remove from memory cache
    this.cache.delete(key);

    // Remove from chrome storage
    try {
      await chrome.storage.local.remove([key]);
    } catch (error) {
      console.error('Failed to remove from storage:', error);
    }
  }

  /**
   * Clear all data
   */
  async clear() {
    // Clear memory cache
    this.cache.clear();

    // Clear chrome storage
    try {
      await chrome.storage.local.clear();
    } catch (error) {
      console.error('Failed to clear storage:', error);
    }
  }

  /**
   * Get bookmarks from cache
   */
  async getBookmarks() {
    const cached = await this.get(STORAGE_KEYS.BOOKMARKS_CACHE);
    return cached || { bookmarks: [], total: 0, cached_at: null };
  }

  /**
   * Update bookmarks cache
   */
  async setBookmarks(bookmarks, total) {
    const cacheData = {
      bookmarks,
      total,
      cached_at: new Date().toISOString()
    };

    await this.set(STORAGE_KEYS.BOOKMARKS_CACHE, cacheData);
  }

  /**
   * Add bookmark to cache
   */
  async addBookmarkToCache(bookmark) {
    const cached = await this.getBookmarks();
    cached.bookmarks.unshift(bookmark); // Add to beginning
    cached.total += 1;
    cached.cached_at = new Date().toISOString();

    await this.set(STORAGE_KEYS.BOOKMARKS_CACHE, cached);
  }

  /**
   * Update bookmark in cache
   */
  async updateBookmarkInCache(bookmarkId, updates) {
    const cached = await this.getBookmarks();
    const index = cached.bookmarks.findIndex(b => b.id === bookmarkId);

    if (index !== -1) {
      cached.bookmarks[index] = { ...cached.bookmarks[index], ...updates };
      cached.cached_at = new Date().toISOString();
      await this.set(STORAGE_KEYS.BOOKMARKS_CACHE, cached);
    }
  }

  /**
   * Remove bookmark from cache
   */
  async removeBookmarkFromCache(bookmarkId) {
    const cached = await this.getBookmarks();
    const index = cached.bookmarks.findIndex(b => b.id === bookmarkId);

    if (index !== -1) {
      cached.bookmarks.splice(index, 1);
      cached.total -= 1;
      cached.cached_at = new Date().toISOString();
      await this.set(STORAGE_KEYS.BOOKMARKS_CACHE, cached);
    }
  }

  /**
   * Get user data
   */
  async getUserData() {
    return await this.get(STORAGE_KEYS.USER_DATA);
  }

  /**
   * Set user data
   */
  async setUserData(userData) {
    await this.set(STORAGE_KEYS.USER_DATA, userData);
  }

  /**
   * Get sync state
   */
  async getSyncState() {
    return await this.get(STORAGE_KEYS.SYNC_STATE);
  }

  /**
   * Set sync state
   */
  async setSyncState(syncState) {
    await this.set(STORAGE_KEYS.SYNC_STATE, syncState);
  }

  /**
   * Get authentication token
   */
  async getAuthToken() {
    return await this.get(STORAGE_KEYS.AUTH_TOKEN);
  }

  /**
   * Set authentication token
   */
  async setAuthToken(token) {
    await this.set(STORAGE_KEYS.AUTH_TOKEN, token);
  }

  /**
   * Check if data is stale
   */
  isStale(cachedAt, maxAgeMs = 300000) { // Default 5 minutes
    if (!cachedAt) return true;

    const age = Date.now() - new Date(cachedAt).getTime();
    return age > maxAgeMs;
  }

  /**
   * Get storage usage statistics
   */
  async getStorageStats() {
    try {
      const usage = await chrome.storage.local.getBytesInUse();
      const quota = chrome.storage.local.QUOTA_BYTES;

      return {
        used: usage,
        quota: quota,
        available: quota - usage,
        percentUsed: (usage / quota) * 100
      };
    } catch (error) {
      console.error('Failed to get storage stats:', error);
      return null;
    }
  }

  /**
   * Cleanup old cache entries
   */
  async cleanup() {
    try {
      const bookmarksCache = await this.getBookmarks();

      // Remove old bookmarks cache if stale
      if (this.isStale(bookmarksCache.cached_at, 3600000)) { // 1 hour
        await this.remove(STORAGE_KEYS.BOOKMARKS_CACHE);
      }

      // Clear memory cache
      this.cache.clear();

      console.log('Storage cleanup completed');
    } catch (error) {
      console.error('Failed to cleanup storage:', error);
    }
  }
}

// Export singleton instance
export const storageManager = new StorageManager();