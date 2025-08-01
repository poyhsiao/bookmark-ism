// Safari Extension Background Script
// Main background script for Safari Web Extension

import { authManager } from './auth-manager.js';
import { syncManager } from './sync-manager.js';
import { storageManager } from './storage-manager.js';
import { safariImporter } from './safari-importer.js';
import { errorHandler } from './error-handler.js';
import { apiClient } from '../../shared/api-client.js';
import { extractPageMetadata } from '../../shared/utils.js';

// Initialize extension
browser.runtime.onInstalled.addListener(async (details) => {
  try {
    console.log('Safari Extension installed:', details.reason);

    // Setup global error handlers
    errorHandler.setupGlobalErrorHandlers();

    // Initialize managers
    await storageManager.init();
    await authManager.init();
    await syncManager.init();
    await apiClient.init();

    // Set up context menu
    if (browser.contextMenus) {
      browser.contextMenus.create({
        id: 'bookmark-current-page',
        title: 'Bookmark this page',
        contexts: ['page']
      });

      browser.contextMenus.create({
        id: 'import-safari-bookmarks',
        title: 'Import Safari bookmarks',
        contexts: ['browser_action']
      });
    }

    console.log('Safari Extension: Initialization completed');

  } catch (error) {
    await errorHandler.handleSafariError(error, 'extension_install');
  }
});

// Handle extension startup
browser.runtime.onStartup.addListener(async () => {
  try {
    console.log('Safari Extension started');

    // Initialize managers
    await storageManager.init();
    await authManager.init();
    await syncManager.init();
    await apiClient.init();

    // Connect to sync if user is authenticated
    if (authManager.isUserAuthenticated()) {
      await syncManager.connect();
    }

  } catch (error) {
    await errorHandler.handleSafariError(error, 'extension_startup');
  }
});

// Handle context menu clicks
if (browser.contextMenus) {
  browser.contextMenus.onClicked.addListener(async (info, tab) => {
    try {
      if (info.menuItemId === 'bookmark-current-page') {
        await handleBookmarkCurrentPage(tab);
      } else if (info.menuItemId === 'import-safari-bookmarks') {
        await handleImportSafariBookmarks();
      }
    } catch (error) {
      await errorHandler.handleSafariError(error, 'context_menu');
    }
  });
}

// Handle messages from popup and content scripts
browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
  handleMessage(message, sender, sendResponse);
  return true; // Keep message channel open for async response
});

// Handle bookmark current page
async function handleBookmarkCurrentPage(tab) {
  try {
    if (!authManager.isUserAuthenticated()) {
      // Open popup for authentication
      if (browser.browserAction && browser.browserAction.openPopup) {
        browser.browserAction.openPopup();
      }
      return;
    }

    const metadata = await extractPageMetadata();
    if (!metadata) {
      console.warn('Safari Extension: Could not extract page metadata');
      return;
    }

    // Create bookmark via API
    const response = await apiClient.createBookmark({
      url: metadata.url,
      title: metadata.title,
      description: '',
      tags: [],
      favicon: metadata.favicon
    });

    if (response.success) {
      // Create sync event
      await syncManager.createSyncEvent(
        'create',
        'bookmark',
        response.data.id,
        response.data
      );

      // Update cache
      await storageManager.addBookmarkToCache(response.data);

      // Show notification
      if (browser.notifications) {
        browser.notifications.create({
          type: 'basic',
          iconUrl: 'icons/icon48.png',
          title: 'Bookmark Saved',
          message: `"${metadata.title}" has been bookmarked`
        });
      }
    }
  } catch (error) {
    await errorHandler.handleSafariError(error, 'bookmark_current_page');

    if (browser.notifications) {
      browser.notifications.create({
        type: 'basic',
        iconUrl: 'icons/icon48.png',
        title: 'Bookmark Failed',
        message: 'Failed to save bookmark. Please try again.'
      });
    }
  }
}

// Handle Safari bookmark import
async function handleImportSafariBookmarks() {
  try {
    if (!authManager.isUserAuthenticated()) {
      console.warn('Safari Extension: User not authenticated for import');
      return;
    }

    console.log('Safari Extension: Starting Safari bookmark import');

    const result = await safariImporter.importSafariBookmarks();

    if (result.success) {
      // Show success notification
      if (browser.notifications) {
        browser.notifications.create({
          type: 'basic',
          iconUrl: 'icons/icon48.png',
          title: 'Import Completed',
          message: `Imported ${result.stats.processed} bookmarks from Safari`
        });
      }

      // Refresh bookmark cache
      const bookmarks = await apiClient.getBookmarks();
      if (bookmarks.success) {
        await storageManager.setBookmarks(bookmarks.data, bookmarks.total);
      }
    } else {
      throw new Error(result.error || 'Import failed');
    }

  } catch (error) {
    await errorHandler.handleSafariError(error, 'safari_import');

    if (browser.notifications) {
      browser.notifications.create({
        type: 'basic',
        iconUrl: 'icons/icon48.png',
        title: 'Import Failed',
        message: 'Failed to import Safari bookmarks. Please try again.'
      });
    }
  }
}

// Handle messages from other parts of the extension
async function handleMessage(message, sender, sendResponse) {
  try {
    switch (message.type) {
      case 'GET_AUTH_STATE':
        sendResponse({
          authenticated: authManager.isUserAuthenticated(),
          user: authManager.getCurrentUser()
        });
        break;

      case 'LOGIN':
        const loginResult = await authManager.login(message.email, message.password);
        if (loginResult.success) {
          await syncManager.connect();
        }
        sendResponse(loginResult);
        break;

      case 'REGISTER':
        const registerResult = await authManager.register(
          message.email,
          message.password,
          message.name
        );
        if (registerResult.success) {
          await syncManager.connect();
        }
        sendResponse(registerResult);
        break;

      case 'LOGOUT':
        await syncManager.disconnect();
        const logoutResult = await authManager.logout();
        sendResponse(logoutResult);
        break;

      case 'GET_BOOKMARKS':
        const bookmarks = await storageManager.getBookmarks();

        // Refresh from API if cache is stale
        if (storageManager.isStale(bookmarks.cached_at)) {
          try {
            const response = await apiClient.getBookmarks(message.params || {});
            if (response.success) {
              await storageManager.setBookmarks(response.data, response.total);
              sendResponse({ success: true, data: response.data, total: response.total });
            } else {
              sendResponse(bookmarks);
            }
          } catch (error) {
            console.error('Safari Extension: Failed to refresh bookmarks:', error);
            sendResponse(bookmarks);
          }
        } else {
          sendResponse(bookmarks);
        }
        break;

      case 'CREATE_BOOKMARK':
        if (!authManager.isUserAuthenticated()) {
          sendResponse({ success: false, error: 'Not authenticated' });
          break;
        }

        try {
          const response = await apiClient.createBookmark(message.data);
          if (response.success) {
            await syncManager.createSyncEvent(
              'create',
              'bookmark',
              response.data.id,
              response.data
            );
            await storageManager.addBookmarkToCache(response.data);
          }
          sendResponse(response);
        } catch (error) {
          const errorResult = await errorHandler.handleSafariError(error, 'create_bookmark');
          sendResponse(errorResult);
        }
        break;

      case 'UPDATE_BOOKMARK':
        if (!authManager.isUserAuthenticated()) {
          sendResponse({ success: false, error: 'Not authenticated' });
          break;
        }

        try {
          const response = await apiClient.updateBookmark(message.id, message.data);
          if (response.success) {
            await syncManager.createSyncEvent(
              'update',
              'bookmark',
              message.id,
              message.data
            );
            await storageManager.updateBookmarkInCache(message.id, message.data);
          }
          sendResponse(response);
        } catch (error) {
          const errorResult = await errorHandler.handleSafariError(error, 'update_bookmark');
          sendResponse(errorResult);
        }
        break;

      case 'DELETE_BOOKMARK':
        if (!authManager.isUserAuthenticated()) {
          sendResponse({ success: false, error: 'Not authenticated' });
          break;
        }

        try {
          const response = await apiClient.deleteBookmark(message.id);
          if (response.success) {
            await syncManager.createSyncEvent('delete', 'bookmark', message.id);
            await storageManager.removeBookmarkFromCache(message.id);
          }
          sendResponse(response);
        } catch (error) {
          const errorResult = await errorHandler.handleSafariError(error, 'delete_bookmark');
          sendResponse(errorResult);
        }
        break;

      case 'GET_SYNC_STATUS':
        sendResponse(syncManager.getConnectionStatus());
        break;

      case 'FORCE_SYNC':
        if (syncManager.isConnected) {
          syncManager.requestSync();
          sendResponse({ success: true });
        } else {
          sendResponse({ success: false, error: 'Not connected' });
        }
        break;

      case 'IMPORT_SAFARI_BOOKMARKS':
        if (!authManager.isUserAuthenticated()) {
          sendResponse({ success: false, error: 'Not authenticated' });
          break;
        }

        try {
          const importResult = await safariImporter.importSafariBookmarks();
          sendResponse(importResult);
        } catch (error) {
          const errorResult = await errorHandler.handleSafariError(error, 'import_safari_bookmarks');
          sendResponse(errorResult);
        }
        break;

      case 'GET_IMPORT_PROGRESS':
        sendResponse(safariImporter.getImportProgress());
        break;

      case 'GET_ERROR_LOG':
        sendResponse(errorHandler.getErrorLog());
        break;

      case 'CLEAR_ERROR_LOG':
        errorHandler.clearErrorLog();
        sendResponse({ success: true });
        break;

      case 'GET_PREFERENCES':
        const preferences = await storageManager.getPreferences();
        sendResponse({ success: true, data: preferences });
        break;

      case 'SET_PREFERENCES':
        await storageManager.setPreferences(message.data);
        sendResponse({ success: true });
        break;

      default:
        console.warn('Safari Extension: Unknown message type:', message.type);
        sendResponse({ success: false, error: 'Unknown message type' });
    }
  } catch (error) {
    console.error('Safari Extension: Error handling message:', error);
    const errorResult = await errorHandler.handleSafariError(error, 'message_handling');
    sendResponse(errorResult);
  }
}

// Handle tab updates to detect navigation
if (browser.tabs && browser.tabs.onUpdated) {
  browser.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (changeInfo.status === 'complete' && tab.url) {
      // Could implement auto-bookmark detection here
      console.log('Safari Extension: Tab updated:', tab.url);
    }
  });
}

// Periodic cleanup and maintenance
setInterval(async () => {
  try {
    await storageManager.cleanup();
  } catch (error) {
    await errorHandler.handleSafariError(error, 'periodic_cleanup');
  }
}, 3600000); // Every hour

// Handle storage changes for cross-tab communication
if (browser.storage && browser.storage.onChanged) {
  browser.storage.onChanged.addListener((changes, namespace) => {
    if (namespace === 'local' && changes.sync_notification) {
      // Handle sync notifications from other parts of the extension
      try {
        const notification = JSON.parse(changes.sync_notification.newValue);
        console.log('Safari Extension: Received sync notification:', notification);
      } catch (error) {
        console.warn('Safari Extension: Failed to parse sync notification:', error);
      }
    }
  });
}

console.log('Safari Extension: Background script loaded');