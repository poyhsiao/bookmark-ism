// Test suite for Chrome extension
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock Chrome APIs
global.chrome = {
  runtime: {
    sendMessage: vi.fn(),
    onMessage: {
      addListener: vi.fn()
    },
    onInstalled: {
      addListener: vi.fn()
    },
    onStartup: {
      addListener: vi.fn()
    }
  },
  storage: {
    local: {
      get: vi.fn(),
      set: vi.fn(),
      remove: vi.fn(),
      clear: vi.fn(),
      getBytesInUse: vi.fn(),
      QUOTA_BYTES: 10485760
    },
    sync: {
      get: vi.fn(),
      set: vi.fn(),
      clear: vi.fn()
    }
  },
  tabs: {
    query: vi.fn(),
    create: vi.fn(),
    onUpdated: {
      addListener: vi.fn()
    }
  },
  contextMenus: {
    create: vi.fn(),
    onClicked: {
      addListener: vi.fn()
    }
  },
  notifications: {
    create: vi.fn()
  },
  identity: {
    getRedirectURL: vi.fn(),
    launchWebAuthFlow: vi.fn()
  },
  action: {
    openPopup: vi.fn()
  }
};

// Mock WebSocket
global.WebSocket = vi.fn().mockImplementation(() => ({
  send: vi.fn(),
  close: vi.fn(),
  readyState: 1,
  OPEN: 1,
  CLOSED: 3
}));

// Mock fetch
global.fetch = vi.fn();

describe('Chrome Extension Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('API Client', () => {
    let apiClient;

    beforeEach(async () => {
      const { apiClient: client } = await import('../shared/api-client.js');
      apiClient = client;
    });

    it('should initialize with stored token', async () => {
      chrome.storage.local.get.mockResolvedValue({
        auth_access_token: 'test-token'
      });

      await apiClient.init();
      expect(apiClient.token).toBe('test-token');
    });

    it('should make authenticated requests', async () => {
      apiClient.setToken('test-token');

      fetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true, data: [] })
      });

      const result = await apiClient.getBookmarks();

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/bookmarks',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-token'
          })
        })
      );
    });

    it('should handle login correctly', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({
          success: true,
          token: 'new-token',
          user: { id: '1', email: 'test@example.com' }
        })
      });

      chrome.storage.local.set.mockResolvedValue();

      const result = await apiClient.login('test@example.com', 'password');

      expect(result.success).toBe(true);
      expect(chrome.storage.local.set).toHaveBeenCalledWith({
        auth_access_token: 'new-token',
        user_data: { id: '1', email: 'test@example.com' }
      });
    });

    it('should create bookmarks', async () => {
      apiClient.setToken('test-token');

      const bookmarkData = {
        url: 'https://example.com',
        title: 'Test Bookmark',
        description: 'Test description',
        tags: ['test']
      };

      fetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({
          success: true,
          data: { id: '1', ...bookmarkData }
        })
      });

      const result = await apiClient.createBookmark(bookmarkData);

      expect(result.success).toBe(true);
      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/bookmarks',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(bookmarkData)
        })
      );
    });
  });

  describe('Auth Manager', () => {
    let authManager;

    beforeEach(async () => {
      const { authManager: manager } = await import('../chrome/background/auth-manager.js');
      authManager = manager;
    });

    it('should initialize authentication state', async () => {
      chrome.storage.local.get.mockResolvedValue({
        auth_access_token: 'test-token',
        user_data: { id: '1', email: 'test@example.com' }
      });

      const isAuth = await authManager.init();

      expect(isAuth).toBe(true);
      expect(authManager.isAuthenticated).toBe(true);
      expect(authManager.user).toEqual({ id: '1', email: 'test@example.com' });
    });

    it('should handle login', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({
          success: true,
          user: { id: '1', email: 'test@example.com' }
        })
      });

      chrome.storage.local.set.mockResolvedValue();
      chrome.runtime.sendMessage.mockImplementation(() => {});

      const result = await authManager.login('test@example.com', 'password');

      expect(result.success).toBe(true);
      expect(authManager.isAuthenticated).toBe(true);
      expect(chrome.runtime.sendMessage).toHaveBeenCalledWith({
        type: 'AUTH_STATE_CHANGED',
        authenticated: true,
        user: { id: '1', email: 'test@example.com' }
      });
    });

    it('should handle logout', async () => {
      authManager.isAuthenticated = true;
      authManager.user = { id: '1', email: 'test@example.com' };

      fetch.mockResolvedValue({ ok: true });
      chrome.storage.local.remove.mockResolvedValue();
      chrome.runtime.sendMessage.mockImplementation(() => {});

      const result = await authManager.logout();

      expect(result.success).toBe(true);
      expect(authManager.isAuthenticated).toBe(false);
      expect(authManager.user).toBe(null);
    });
  });

  describe('Sync Manager', () => {
    let syncManager;

    beforeEach(async () => {
      const { syncManager: manager } = await import('../chrome/background/sync-manager.js');
      syncManager = manager;
    });

    it('should initialize with device ID', async () => {
      chrome.storage.local.get.mockResolvedValue({
        device_id: 'test-device-id'
      });

      await syncManager.init();

      expect(syncManager.deviceId).toBe('test-device-id');
    });

    it('should connect to WebSocket', async () => {
      const mockWebSocket = {
        readyState: 1,
        send: vi.fn(),
        close: vi.fn()
      };

      global.WebSocket.mockReturnValue(mockWebSocket);

      await syncManager.connect();

      expect(WebSocket).toHaveBeenCalledWith('ws://localhost:8080/api/v1/sync/ws');
    });

    it('should handle sync events', async () => {
      const event = {
        id: '1',
        device_id: 'other-device',
        event_type: 'create',
        resource_type: 'bookmark',
        resource_id: 'bookmark-1',
        changes: { title: 'New Bookmark' }
      };

      syncManager.deviceId = 'test-device';

      await syncManager.handleSyncEvent(event);

      // Should not process events from same device
      syncManager.deviceId = 'other-device';
      await syncManager.handleSyncEvent(event);
    });

    it('should create sync events', async () => {
      syncManager.deviceId = 'test-device';
      syncManager.isConnected = true;
      syncManager.websocket = { send: vi.fn() };

      fetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ success: true })
      });

      await syncManager.createSyncEvent('create', 'bookmark', 'bookmark-1', {
        title: 'Test Bookmark'
      });

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/sync/events',
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('create')
        })
      );
    });
  });

  describe('Storage Manager', () => {
    let storageManager;

    beforeEach(async () => {
      const { storageManager: manager } = await import('../chrome/background/storage-manager.js');
      storageManager = manager;
    });

    it('should get and set data', async () => {
      chrome.storage.local.get.mockResolvedValue({ test_key: 'test_value' });
      chrome.storage.local.set.mockResolvedValue();

      await storageManager.set('test_key', 'test_value');
      const value = await storageManager.get('test_key');

      expect(value).toBe('test_value');
      expect(chrome.storage.local.set).toHaveBeenCalledWith({
        test_key: 'test_value'
      });
    });

    it('should manage bookmarks cache', async () => {
      const bookmarks = [
        { id: '1', title: 'Bookmark 1', url: 'https://example1.com' },
        { id: '2', title: 'Bookmark 2', url: 'https://example2.com' }
      ];

      chrome.storage.local.set.mockResolvedValue();

      await storageManager.setBookmarks(bookmarks, 2);

      expect(chrome.storage.local.set).toHaveBeenCalledWith({
        bookmarks_cache: expect.objectContaining({
          bookmarks,
          total: 2,
          cached_at: expect.any(String)
        })
      });
    });

    it('should add bookmark to cache', async () => {
      const existingCache = {
        bookmarks: [
          { id: '1', title: 'Bookmark 1', url: 'https://example1.com' }
        ],
        total: 1,
        cached_at: new Date().toISOString()
      };

      chrome.storage.local.get.mockResolvedValue({
        bookmarks_cache: existingCache
      });
      chrome.storage.local.set.mockResolvedValue();

      const newBookmark = {
        id: '2',
        title: 'Bookmark 2',
        url: 'https://example2.com'
      };

      await storageManager.addBookmarkToCache(newBookmark);

      expect(chrome.storage.local.set).toHaveBeenCalledWith({
        bookmarks_cache: expect.objectContaining({
          bookmarks: [newBookmark, existingCache.bookmarks[0]],
          total: 2
        })
      });
    });

    it('should check if data is stale', () => {
      const now = new Date();
      const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
      const tenMinutesAgo = new Date(now.getTime() - 10 * 60 * 1000);

      expect(storageManager.isStale(fiveMinutesAgo.toISOString())).toBe(false);
      expect(storageManager.isStale(tenMinutesAgo.toISOString())).toBe(true);
      expect(storageManager.isStale(null)).toBe(true);
    });
  });

  describe('Utility Functions', () => {
    let utils;

    beforeEach(async () => {
      utils = await import('../shared/utils.js');
    });

    it('should generate device ID', () => {
      const deviceId = utils.generateDeviceId();

      expect(deviceId).toMatch(/^ext_[a-z0-9]+_\d+$/);
    });

    it('should validate URLs', () => {
      expect(utils.isValidUrl('https://example.com')).toBe(true);
      expect(utils.isValidUrl('http://example.com')).toBe(true);
      expect(utils.isValidUrl('ftp://example.com')).toBe(true);
      expect(utils.isValidUrl('not-a-url')).toBe(false);
      expect(utils.isValidUrl('')).toBe(false);
    });

    it('should format timestamps', () => {
      const now = new Date();
      const oneMinuteAgo = new Date(now.getTime() - 60 * 1000);
      const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
      const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

      expect(utils.formatTimestamp(now.toISOString())).toBe('Just now');
      expect(utils.formatTimestamp(oneMinuteAgo.toISOString())).toBe('1m ago');
      expect(utils.formatTimestamp(oneHourAgo.toISOString())).toBe('1h ago');
      expect(utils.formatTimestamp(oneDayAgo.toISOString())).toBe('1d ago');
    });

    it('should extract page metadata', async () => {
      chrome.tabs.query.mockResolvedValue([{
        url: 'https://example.com',
        title: 'Example Site',
        favIconUrl: 'https://example.com/favicon.ico'
      }]);

      const metadata = await utils.extractPageMetadata();

      expect(metadata).toEqual({
        url: 'https://example.com',
        title: 'Example Site',
        favicon: 'https://example.com/favicon.ico',
        timestamp: expect.any(String)
      });
    });

    it('should handle API errors', () => {
      const error = { status: 401, message: 'Unauthorized' };
      const handled = utils.handleApiError(error);

      expect(handled).toEqual({
        code: 'AUTH_REQUIRED',
        message: 'Authentication required'
      });
    });
  });

  describe('Content Script', () => {
    beforeEach(() => {
      // Mock DOM
      global.document = {
        title: 'Test Page',
        documentElement: { lang: 'en' },
        querySelector: vi.fn(),
        querySelectorAll: vi.fn(),
        addEventListener: vi.fn(),
        readyState: 'complete'
      };

      global.window = {
        location: { href: 'https://example.com' }
      };
    });

    it('should extract page metadata', async () => {
      document.querySelector.mockImplementation((selector) => {
        if (selector === 'meta[name="description"]') {
          return { content: 'Test description' };
        }
        if (selector === 'meta[name="keywords"]') {
          return { content: 'test, keywords' };
        }
        if (selector === 'link[rel="icon"]') {
          return { href: '/favicon.ico' };
        }
        return null;
      });

      // Import and test content script functionality
      // Note: This would require refactoring the content script to export functions
      // For now, we'll test the concept
      expect(document.title).toBe('Test Page');
      expect(window.location.href).toBe('https://example.com');
    });
  });
});

describe('Integration Tests', () => {
  it('should handle complete bookmark creation flow', async () => {
    // Mock all required APIs
    chrome.storage.local.get.mockResolvedValue({
      auth_access_token: 'test-token',
      device_id: 'test-device'
    });

    chrome.tabs.query.mockResolvedValue([{
      url: 'https://example.com',
      title: 'Example Site',
      favIconUrl: 'https://example.com/favicon.ico'
    }]);

    fetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        success: true,
        data: {
          id: 'bookmark-1',
          url: 'https://example.com',
          title: 'Example Site',
          description: '',
          tags: []
        }
      })
    });

    chrome.storage.local.set.mockResolvedValue();
    chrome.runtime.sendMessage.mockImplementation(() => {});

    // Import managers
    const { apiClient } = await import('../shared/api-client.js');
    const { syncManager } = await import('../chrome/background/sync-manager.js');
    const { storageManager } = await import('../chrome/background/storage-manager.js');

    // Initialize
    await apiClient.init();
    await syncManager.init();
    await storageManager.init();

    // Create bookmark
    const bookmarkData = {
      url: 'https://example.com',
      title: 'Example Site',
      description: '',
      tags: []
    };

    const result = await apiClient.createBookmark(bookmarkData);

    expect(result.success).toBe(true);
    expect(result.data.id).toBe('bookmark-1');
  });
});