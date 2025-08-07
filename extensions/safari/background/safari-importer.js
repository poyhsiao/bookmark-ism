// Safari Bookmark Importer
// Handles importing bookmarks from Safari's native bookmark system

import { apiClient } from '../../shared/api-client.js';

class SafariBookmarkImporter {
  constructor() {
    this.importProgress = {
      total: 0,
      processed: 0,
      errors: 0,
      duplicates: 0
    };
  }

  /**
   * Import bookmarks from Safari
   */
  async importSafariBookmarks() {
    try {
      console.log('Safari Importer: Starting Safari bookmark import');

      // Get all Safari bookmarks using the bookmarks API
      const safariBookmarks = await this.getSafariBookmarks();

      if (!safariBookmarks || safariBookmarks.length === 0) {
        return {
          success: true,
          message: 'No Safari bookmarks found to import',
          stats: this.importProgress
        };
      }

      this.importProgress.total = safariBookmarks.length;
      console.log('Safari Importer: Found', safariBookmarks.length, 'Safari bookmarks');

      // Process bookmarks in batches to avoid overwhelming the API
      const batchSize = 10;
      const results = [];

      for (let i = 0; i < safariBookmarks.length; i += batchSize) {
        const batch = safariBookmarks.slice(i, i + batchSize);
        const batchResults = await this.processBatch(batch);
        results.push(...batchResults);

        // Small delay between batches
        await this.delay(100);
      }

      console.log('Safari Importer: Import completed');

      return {
        success: true,
        message: `Imported ${results.length} bookmarks from Safari`,
        stats: this.importProgress,
        bookmarks: results
      };

    } catch (error) {
      console.error('Safari Importer: Import failed:', error);
      return {
        success: false,
        error: error.message,
        stats: this.importProgress
      };
    }
  }

  /**
   * Get Safari bookmarks using the browser bookmarks API
   */
  async getSafariBookmarks() {
    try {
      // Search for all bookmarks
      const bookmarks = await browser.bookmarks.search({});

      // Filter out folders and organize bookmarks
      const validBookmarks = [];

      for (const bookmark of bookmarks) {
        if (bookmark.url && this.isValidUrl(bookmark.url)) {
          validBookmarks.push({
            id: bookmark.id,
            title: bookmark.title || 'Untitled',
            url: bookmark.url,
            dateAdded: bookmark.dateAdded,
            parentId: bookmark.parentId
          });
        }
      }

      return validBookmarks;
    } catch (error) {
      console.error('Safari Importer: Failed to get Safari bookmarks:', error);
      return [];
    }
  }

  /**
   * Process a batch of bookmarks
   */
  async processBatch(bookmarks) {
    const results = [];

    for (const bookmark of bookmarks) {
      try {
        const result = await this.importSingleBookmark(bookmark);
        if (result) {
          results.push(result);
        }
        this.importProgress.processed++;
      } catch (error) {
        console.error('Safari Importer: Failed to import bookmark:', bookmark.url, error);
        this.importProgress.errors++;
      }
    }

    return results;
  }

  /**
   * Import a single bookmark
   */
  async importSingleBookmark(safariBookmark) {
    try {
      // Check if bookmark already exists
      const existingBookmarks = await apiClient.getBookmarks({
        search: safariBookmark.url,
        limit: 1
      });

      if (existingBookmarks.data && existingBookmarks.data.length > 0) {
        console.log('Safari Importer: Bookmark already exists:', safariBookmark.url);
        this.importProgress.duplicates++;
        return null;
      }

      // Create bookmark data
      const bookmarkData = {
        url: safariBookmark.url,
        title: safariBookmark.title,
        description: '',
        tags: ['imported-from-safari'],
        favicon: await this.getFaviconUrl(safariBookmark.url),
        created_at: safariBookmark.dateAdded ? new Date(safariBookmark.dateAdded).toISOString() : undefined
      };

      // Create bookmark via API
      const response = await apiClient.createBookmark(bookmarkData);

      if (response.success) {
        console.log('Safari Importer: Successfully imported:', safariBookmark.title);
        return response.data;
      } else {
        throw new Error(response.error || 'Failed to create bookmark');
      }

    } catch (error) {
      console.error('Safari Importer: Failed to import single bookmark:', error);
      throw error;
    }
  }

  /**
   * Get favicon URL for a website
   */
  async getFaviconUrl(url) {
    try {
      const domain = new URL(url).hostname;
      return `https://www.google.com/s2/favicons?domain=${domain}&sz=32`;
    } catch (error) {
      return null;
    }
  }

  /**
   * Validate URL format
   */
  isValidUrl(string) {
    try {
      const url = new URL(string);
      return url.protocol === 'http:' || url.protocol === 'https:';
    } catch (_) {
      return false;
    }
  }

  /**
   * Get import progress
   */
  getImportProgress() {
    return {
      ...this.importProgress,
      percentage: this.importProgress.total > 0
        ? Math.round((this.importProgress.processed / this.importProgress.total) * 100)
        : 0
    };
  }

  /**
   * Reset import progress
   */
  resetProgress() {
    this.importProgress = {
      total: 0,
      processed: 0,
      errors: 0,
      duplicates: 0
    };
  }

  /**
   * Utility delay function
   */
  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Export Safari bookmarks to JSON format
   */
  async exportSafariBookmarks() {
    try {
      const bookmarks = await this.getSafariBookmarks();

      const exportData = {
        timestamp: new Date().toISOString(),
        source: 'Safari',
        version: '1.0',
        bookmarks: bookmarks.map(bookmark => ({
          title: bookmark.title,
          url: bookmark.url,
          dateAdded: bookmark.dateAdded,
          tags: ['exported-from-safari']
        }))
      };

      return {
        success: true,
        data: exportData,
        filename: `safari-bookmarks-${new Date().toISOString().split('T')[0]}.json`
      };

    } catch (error) {
      console.error('Safari Importer: Export failed:', error);
      return {
        success: false,
        error: error.message
      };
    }
  }

  /**
   * Get Safari bookmark folders structure
   */
  async getSafariBookmarkFolders() {
    try {
      const tree = await browser.bookmarks.getTree();
      const folders = [];

      const extractFolders = (nodes, path = '') => {
        for (const node of nodes) {
          if (!node.url) { // It's a folder
            const folderPath = path ? `${path}/${node.title}` : node.title;
            folders.push({
              id: node.id,
              title: node.title,
              path: folderPath,
              parentId: node.parentId
            });

            if (node.children) {
              extractFolders(node.children, folderPath);
            }
          }
        }
      };

      extractFolders(tree);
      return folders;

    } catch (error) {
      console.error('Safari Importer: Failed to get bookmark folders:', error);
      return [];
    }
  }
}

// Export singleton instance
export const safariImporter = new SafariBookmarkImporter();