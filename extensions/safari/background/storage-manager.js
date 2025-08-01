// Safari Extension Storage Manager
// Handles local storage and caching for Safari Web Extension

import { STORAGE_KEYS } from '../../shared/constants.js';

class SafariStorageManager {
  constructor() {
    this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
    this.maxCacheSize = 1000; // Maximum number of bookmarks to cache
  }

  /**
   * Initialize storage manager
   */
  async init() {
    try {
      // Check storage quota (Safari has limitations)
      const usage = await this.getStorageUsage();
      console.log('Safari Storage Manager: Current storage usage:', usage);

      // Clean up old data if needed
      await this.cleanup();

      console.log('Safari Storage Manager: Initialized successfully');
    } catch (error) {
      console.error('Safari Storage Manager: Failed to initialize:', error);
    }
  }

  /**
   * Get storage usage information
   */
  async getStorageUsage() {
    try {
      const allData = await browser.storage.local.get(null);
      const dataSize = JSON.stringify(allData).length;

      return {
        bytes: dataSize,
        items: Object.keys(allData).length,
        quota: 10 * 1024 * 1024 // Safari typically allows 10MB
      };
    } catch (error) {
      console.error('Safari Storage Manager: Failed to get storage usage:', error);
      return { bytes: 0, items: 0, quota: 0 };
    }
  }

  /**
   * Set bookmarks cache
   */
  async setBookmarks(bookmarks, total) {
    try {
      // Limit cache size for Safari
      const limitedBookmarks = bookmarks.slice(0, this.maxCacheSize);

      const cacheData = {
        bookmarks: limitedBookmarks,
        total: total,
        cached_at: Date.now()
      };

      await browser.storage.local.set({
        [STORAGE_KEYS.BOOKMARKS_CACHE]: JSON.stringify(cacheData)
      });

      console.log('Safari Storage Manager: Cached', limitedBookmarks.length, 'bookmarks');
    } catch (error) {
      console.error('Safari Storage Manager: Failed to cache bookmarks:', error);
    }
  }

  /**
   * Get bookmarks from cache
   */
  async getBookmarks() {
    try {
      const stored = await browser.storage.local.get([STORAGE_KEYS.BOOKMARKS_CACHE]);

      if (stored[STORAGE_KEYS.BOOKMARKS_CACHE]) {
        const cacheData = JSON.parse(stored[STORAGE_KEYS.BOOKMARKS_CACHE]);

        return {
          success: true,
          bookmarks: cacheData.bookmarks || [],
          total: cacheData.total || 0,
          cached_at: cacheData.cached_at,
          is_stale: this.isStale(cacheData.cached_at)
        };
      }

      return {
        success: true,
        bookmarks: [],
        total: 0,
        cached_at: null,
        is_stale: true
      };
    } catch (error) {
      console.error('Safari Storage Manager: Failed to get bookmarks:', error);
      return {
        success: false,
        bookmarks: [],
        total: 0,
        error: error.message
      };
    }
  }

  /**
   * Add bookmark to cache
   */
  async addBookmarkToCache(bookmark) {
    try {
      const cached = await this.getBookmarks();

      if (cached.success) {
        // Check if bookmark already exists
        const existingIndex = cached.bookmarks.findIndex(b => b.id === bookmark.id);

        if (existingIndex === -1) {
          // Add new bookmark
          cached.bookmarks.unshift(bookmark);
          cached.total += 1;
        } else {
          // Update existing bookmark
          cached.bookmarks[existingIndex] = bookmark;
        }

        // Limit cache size
        if (cached.bookmarks.length > this.maxCacheSize) {
          cached.bookmarks = cached.bookmarks.slice(0, this.maxCacheSize);
        }

        await this.setBookmarks(cached.bookmarks, cached.total);
        console.log('Safari Storage Manager: Added bookmark to cache:', bookmark.id);
      }
    } catch (error) {
      console.error('Safari Storage Manager: Failed to add bookmark to cache:', error);
    }
  }

  /**
   * Update bookmark in cache
   */
  async updateBookmarkInCache(bookmarkId, updates) {
    try {
      const cached = await this.getBookmarks();

      if (cached.success) {
        const bookmarkIndex = cached.bookmarks.findIndex(b => b.id === bookmarkId);

        if (bookmarkIndex !== -1) {
          // Update bookmark
          cached.bookmarks[bookmarkIndex] = {
            ...cached.bookmarks[bookmarkIndex],
            ...updates,
            updated_at: new Date().toISOString()
          };

          await this.setBookmarks(cached.bookmarks, cached.total);
          console.log('Safari Storage Manager: Updated bookmark in cache:', bookmarkId);
        }
      }
    } catch (error) {
      console.error('Safari Storage Manager: Failed to update bookmark in cache:', error);
    }
  }

  /**
   * Remove bookmark from cache
   */
  async removeBookmarkFromCache(bookmarkId) {
    try {
      const cached = await this.getBookmarks();

      if (cached.success) {
        const initialLength = cached.bookmarks.length;
        cached.bookmarks = cached.bookmarks.filter(b => b.id !== bookmarkId);

        if (cached.bookmarks.length < initialLength) {
          cached.total = Math.max(0, cached.total - 1);
          await this.setBookmarks(cached.bookmarks, cached.total);
          console.log('Safari Storage Manager: Removed bookmark from cache:', bookmarkId);
        }
      }
    } catch (error) {
      console.error('Safari Storage Manager: Failed to remove bookmark from cache:', error);
    }
  }

  /**
   * Check if cache is stale
   */
  isStale(cachedAt) {
    if (!cachedAt) return true;
    return (Date.now() - cachedAt) > this.cacheTimeout;
  }

  /**
   * Clear all cached data
   */
  async clearCache() {
    try {
      await browser.storage.local.remove([
        STORAGE_KEYS.BOOKMARKS_CACHE,
        STORAGE_KEYS.SYNC_STATE
      ]);

      console.log('Safari Storage Manager: Cache cleared');
    } catch (error) {
      console.error('Safari Storage Manager: Failed to clear cache:', error);
    }
  }

  /**
   * Set user preferences
   */
  async setPreferences(preferences) {
    try {
      await browser.storage.local.set({
        user_preferences: JSON.stringify(preferences)
      });

      console.log('Safari Storage Manager: Preferences saved');
    } catch (error) {
      console.error('Safari Storage Manager: Failed to save preferences:', error);
    }
  }

  /**
   * Get user preferences
   */
  async getPreferences() {
    try {
      const stored = await browser.storage.local.get(['user_preferences']);

      if (stored.user_preferences) {
        return JSON.parse(stored.user_preferences);
      }

      // Return default preferences
      return {
        theme: 'auto',
        viewMode: 'grid',
        gridSize: 'medium',
        notifications: true,
        autoSync: true
      };
    } catch (error) {
      console.error('Safari Storage Manager: Failed to get preferences:', error);
      return {};
    }
  }

  /**
   * Cleanup old data and optimize storage
   */
  async cleanup() {
    try {
      const usage = await this.getStorageUsage();

      // If storage is getting full (>80% of quota), clean up
      if (usage.bytes > usage.quota * 0.8) {
        console.log('Safari Storage Manager: Storage getting full, cleaning up...');

        // Remove old sync events
        const stored = await browser.storage.local.get([STORAGE_KEYS.SYNC_STATE]);
        if (stored[STORAGE_KEYS.SYNC_STATE]) {
          const syncState = JSON.parse(stored[STORAGE_KEYS.SYNC_STATE]);

          // Keep only recent events (last 24 hours)
          const oneDayAgo = Date.now() - (24 * 60 * 60 * 1000);
          syncState.queue = syncState.queue.filter(event =>
            new Date(event.timestamp).getTime() > oneDayAgo
          );

          await browser.storage.local.set({
            [STORAGE_KEYS.SYNC_STATE]: JSON.stringify(syncState)
          });
        }

        // Limit bookmark cache size
        const cached = await this.getBookmarks();
        if (cached.success && cached.bookmarks.length > this.maxCacheSize) {
          const limitedBookmarks = cached.bookmarks.slice(0, this.maxCacheSize);
          await this.setBookmarks(limitedBookmarks, cached.total);
        }
      }

      console.log('Safari Storage Manager: Cleanup completed');
    } catch (error) {
      console.error('Safari Storage Manager: Cleanup failed:', error);
    }
  }

  /**
   * Export data for backup
   */
  async exportData() {
    try {
      const allData = await browser.storage.local.get(null);

      return {
        timestamp: new Date().toISOString(),
        data: allData
      };
    } catch (error) {
      console.error('Safari Storage Manager: Failed to export data:', error);
      return null;
    }
  }

  /**
   * Import data from backup
   */
  async importData(backupData) {
    try {
      if (!backupData || !backupData.data) {
        throw new Error('Invalid backup data');
      }

      await browser.storage.local.clear();
      await browser.storage.local.set(backupData.data);

      console.log('Safari Storage Manager: Data imported successfully');
      return { success: true };
    } catch (error) {
      console.error('Safari Storage Manager: Failed to import data:', error);
      return { success: false, error: error.message };
    }
  }
}

// Export singleton instance
export const storageManager = new SafariStorageManager();