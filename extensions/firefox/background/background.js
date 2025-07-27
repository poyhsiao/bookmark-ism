// Background script for Firefox extension
// Note: Firefox uses background scripts instead of service workers

// Import managers (these are loaded via manifest.json scripts array)
// const authManager, syncManager, storageManager are available globally

// Initialize extension
browser.runtime.onInstalled.addListener(async (details) => {
  console.log('Firefox extension installed:', details.reason);

  // Initialize managers
  await storageManager.init();
  await authManager.init();
  await syncManager.init();

  // Set up context menu
  browser.contextMenus.create({
    id: 'bookmark-current-page',
    title: 'Bookmark this page',
    contexts: ['page']
  });
});

// Handle extension startup
browser.runtime.onStartup.addListener(async () => {
  console.log('Firefox extension started');

  // Initialize managers
  await storageManager.init();
  await authManager.init();
  await syncManager.init();
});

// Handle context menu clicks
browser.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId === 'bookmark-current-page') {
    await handleBookmarkCurrentPage(tab);
  }
});

// Handle messages from popup and content scripts
browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
  handleMessage(message, sender, sendResponse);
  return true; // Keep message channel open for async response
});

// Handle bookmark current page
async function handleBookmarkCurrentPage(tab) {
  if (!authManager.isUserAuthenticated()) {
    // Open popup for authentication
    browser.browserAction.openPopup();
    return;
  }

  try {
    const metadata = await extractPageMetadata(tab);
    if (!metadata) {
      console.warn('Could not extract page metadata');
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
      browser.notifications.create({
        type: 'basic',
        iconUrl: 'icons/icon48.png',
        title: 'Bookmark Saved',
        message: `"${metadata.title}" has been bookmarked`
      });
    }
  } catch (error) {
    console.error('Failed to bookmark page:', error);

    browser.notifications.create({
      type: 'basic',
      iconUrl: 'icons/icon48.png',
      title: 'Bookmark Failed',
      message: 'Failed to save bookmark. Please try again.'
    });
  }
}

// Extract page metadata from tab
async function extractPageMetadata(tab) {
  if (!tab || !tab.url || tab.url.startsWith('about:') || tab.url.startsWith('moz-extension:')) {
    return null;
  }

  return {
    url: tab.url,
    title: tab.title || '',
    favicon: tab.favIconUrl || '',
    timestamp: new Date().toISOString()
  };
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
            console.error('Failed to refresh bookmarks:', error);
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
          sendResponse({ success: false, error: error.message });
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
          sendResponse({ success: false, error: error.message });
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
          sendResponse({ success: false, error: error.message });
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

      default:
        console.warn('Unknown message type:', message.type);
        sendResponse({ success: false, error: 'Unknown message type' });
    }
  } catch (error) {
    console.error('Error handling message:', error);
    sendResponse({ success: false, error: error.message });
  }
}

// Handle tab updates to detect navigation
browser.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete' && tab.url) {
    // Could implement auto-bookmark detection here
    console.log('Tab updated:', tab.url);
  }
});

// Periodic cleanup
setInterval(async () => {
  await storageManager.cleanup();
}, 3600000); // Every hour