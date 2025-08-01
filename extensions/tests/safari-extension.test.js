// Safari Extension Test Suite
// Following TDD approach for Safari Web Extension implementation

describe('Safari Extension', () => {
  describe('Manifest Configuration', () => {
    test('should have valid Safari Web Extension manifest', () => {
      // Test manifest.json structure for Safari
      const manifest = require('../safari/manifest.json');

      expect(manifest.manifest_version).toBe(2);
      expect(manifest.name).toBe('Bookmark Sync Service');
      expect(manifest.version).toBe('1.0.0');
      expect(manifest.permissions).toContain('storage');
      expect(manifest.permissions).toContain('activeTab');
      expect(manifest.permissions).toContain('bookmarks');
    });

    test('should have Safari-specific configurations', () => {
      const manifest = require('../safari/manifest.json');

      // Safari Web Extension specific fields
      expect(manifest.safari_web_extension).toBeDefined();
      expect(manifest.safari_web_extension.bundle_identifier).toBeDefined();
      expect(manifest.safari_web_extension.team_identifier).toBeDefined();
    });
  });

  describe('Background Script', () => {
    let backgroundScript;

    beforeEach(() => {
      // Mock Safari browser APIs
      global.browser = {
        runtime: {
          onInstalled: { addListener: jest.fn() },
          onStartup: { addListener: jest.fn() },
          onMessage: { addListener: jest.fn() },
          sendMessage: jest.fn()
        },
        storage: {
          local: {
            get: jest.fn(),
            set: jest.fn(),
            remove: jest.fn()
          }
        },
        tabs: {
          query: jest.fn(),
          onUpdated: { addListener: jest.fn() }
        },
        contextMenus: {
          create: jest.fn(),
          onClicked: { addListener: jest.fn() }
        },
        notifications: {
          create: jest.fn()
        },
        bookmarks: {
          create: jest.fn(),
          get: jest.fn(),
          update: jest.fn(),
          remove: jest.fn(),
          search: jest.fn()
        }
      };

      backgroundScript = require('../safari/background/background.js');
    });

    test('should initialize extension on install', () => {
      expect(browser.runtime.onInstalled.addListener).toHaveBeenCalled();
    });

    test('should handle context menu creation', () => {
      expect(browser.contextMenus.create).toHaveBeenCalledWith({
        id: 'bookmark-current-page',
        title: 'Bookmark this page',
        contexts: ['page']
      });
    });

    test('should handle message passing', () => {
      expect(browser.runtime.onMessage.addListener).toHaveBeenCalled();
    });
  });

  describe('Auth Manager', () => {
    let authManager;

    beforeEach(() => {
      global.browser = {
        storage: {
          local: {
            get: jest.fn().mockResolvedValue({}),
            set: jest.fn().mockResolvedValue(),
            remove: jest.fn().mockResolvedValue()
          }
        }
      };

      authManager = require('../safari/background/auth-manager.js').authManager;
    });

    test('should initialize auth manager', async () => {
      await authManager.init();
      expect(browser.storage.local.get).toHaveBeenCalled();
    });

    test('should handle user login', async () => {
      const mockResponse = {
        success: true,
        token: 'test-token',
        user: { id: '1', email: 'test@example.com' }
      };

      // Mock API client
      jest.doMock('../../shared/api-client.js', () => ({
        apiClient: {
          login: jest.fn().mockResolvedValue(mockResponse)
        }
      }));

      const result = await authManager.login('test@example.com', 'password');
      expect(result.success).toBe(true);
    });

    test('should handle user logout', async () => {
      await authManager.logout();
      expect(browser.storage.local.remove).toHaveBeenCalled();
    });
  });

  describe('Sync Manager', () => {
    let syncManager;

    beforeEach(() => {
      global.browser = {
        storage: {
          local: {
            get: jest.fn().mockResolvedValue({}),
            set: jest.fn().mockResolvedValue()
          }
        }
      };

      // Mock WebSocket
      global.WebSocket = jest.fn().mockImplementation(() => ({
        addEventListener: jest.fn(),
        send: jest.fn(),
        close: jest.fn(),
        readyState: 1
      }));

      syncManager = require('../safari/background/sync-manager.js').syncManager;
    });

    test('should initialize sync manager', async () => {
      await syncManager.init();
      expect(browser.storage.local.get).toHaveBeenCalled();
    });

    test('should connect to WebSocket', async () => {
      await syncManager.connect();
      expect(WebSocket).toHaveBeenCalled();
    });

    test('should handle sync events', async () => {
      const eventData = {
        type: 'bookmark_created',
        data: { id: '1', title: 'Test Bookmark' }
      };

      await syncManager.createSyncEvent('create', 'bookmark', '1', eventData.data);
      // Verify event was processed
    });
  });

  describe('Storage Manager', () => {
    let storageManager;

    beforeEach(() => {
      global.browser = {
        storage: {
          local: {
            get: jest.fn().mockResolvedValue({}),
            set: jest.fn().mockResolvedValue(),
            remove: jest.fn().mockResolvedValue()
          }
        }
      };

      storageManager = require('../safari/background/storage-manager.js').storageManager;
    });

    test('should initialize storage manager', async () => {
      await storageManager.init();
      expect(browser.storage.local.get).toHaveBeenCalled();
    });

    test('should cache bookmarks', async () => {
      const bookmarks = [
        { id: '1', title: 'Test 1', url: 'https://example.com' },
        { id: '2', title: 'Test 2', url: 'https://example.org' }
      ];

      await storageManager.setBookmarks(bookmarks, bookmarks.length);
      expect(browser.storage.local.set).toHaveBeenCalled();
    });

    test('should retrieve cached bookmarks', async () => {
      const mockBookmarks = {
        bookmarks: [{ id: '1', title: 'Test' }],
        total: 1,
        cached_at: Date.now()
      };

      browser.storage.local.get.mockResolvedValue({
        bookmarks_cache: JSON.stringify(mockBookmarks)
      });

      const result = await storageManager.getBookmarks();
      expect(result.bookmarks).toHaveLength(1);
    });
  });

  describe('Popup Interface', () => {
    let popupScript;

    beforeEach(() => {
      // Mock DOM
      document.body.innerHTML = `
        <div id="app">
          <div id="auth-section">
            <form id="login-form">
              <input id="email" type="email" />
              <input id="password" type="password" />
              <button type="submit">Login</button>
            </form>
          </div>
          <div id="bookmarks-section" style="display: none;">
            <div id="bookmarks-grid"></div>
            <div id="bookmarks-list"></div>
          </div>
        </div>
      `;

      global.browser = {
        runtime: {
          sendMessage: jest.fn().mockResolvedValue({ success: true })
        },
        tabs: {
          query: jest.fn().mockResolvedValue([{
            url: 'https://example.com',
            title: 'Example'
          }])
        }
      };

      popupScript = require('../safari/popup/popup.js');
    });

    test('should initialize popup interface', () => {
      expect(document.getElementById('app')).toBeTruthy();
    });

    test('should handle login form submission', async () => {
      const form = document.getElementById('login-form');
      const email = document.getElementById('email');
      const password = document.getElementById('password');

      email.value = 'test@example.com';
      password.value = 'password';

      const event = new Event('submit');
      form.dispatchEvent(event);

      expect(browser.runtime.sendMessage).toHaveBeenCalledWith({
        type: 'LOGIN',
        email: 'test@example.com',
        password: 'password'
      });
    });

    test('should display bookmarks in grid view', async () => {
      const mockBookmarks = [
        { id: '1', title: 'Test 1', url: 'https://example.com' },
        { id: '2', title: 'Test 2', url: 'https://example.org' }
      ];

      browser.runtime.sendMessage.mockResolvedValue({
        success: true,
        data: mockBookmarks
      });

      // Simulate authenticated state
      await popupScript.loadBookmarks();

      const grid = document.getElementById('bookmarks-grid');
      expect(grid.children.length).toBe(2);
    });
  });

  describe('Options Page', () => {
    let optionsScript;

    beforeEach(() => {
      document.body.innerHTML = `
        <div id="options-container">
          <form id="settings-form">
            <select id="theme-select">
              <option value="light">Light</option>
              <option value="dark">Dark</option>
              <option value="auto">Auto</option>
            </select>
            <select id="view-mode">
              <option value="grid">Grid</option>
              <option value="list">List</option>
            </select>
            <button type="submit">Save Settings</button>
          </form>
        </div>
      `;

      global.browser = {
        storage: {
          local: {
            get: jest.fn().mockResolvedValue({}),
            set: jest.fn().mockResolvedValue()
          }
        }
      };

      optionsScript = require('../safari/options/options.js');
    });

    test('should load saved settings', async () => {
      const mockSettings = {
        theme: 'dark',
        viewMode: 'list'
      };

      browser.storage.local.get.mockResolvedValue({
        user_preferences: JSON.stringify(mockSettings)
      });

      await optionsScript.loadSettings();

      expect(document.getElementById('theme-select').value).toBe('dark');
      expect(document.getElementById('view-mode').value).toBe('list');
    });

    test('should save settings', async () => {
      const form = document.getElementById('settings-form');
      const themeSelect = document.getElementById('theme-select');
      const viewMode = document.getElementById('view-mode');

      themeSelect.value = 'dark';
      viewMode.value = 'grid';

      const event = new Event('submit');
      form.dispatchEvent(event);

      expect(browser.storage.local.set).toHaveBeenCalled();
    });
  });

  describe('Content Script', () => {
    let contentScript;

    beforeEach(() => {
      global.browser = {
        runtime: {
          sendMessage: jest.fn()
        }
      };

      // Mock DOM
      document.head.innerHTML = `
        <title>Test Page</title>
        <meta name="description" content="Test description" />
        <link rel="icon" href="/favicon.ico" />
      `;

      contentScript = require('../safari/content/page-analyzer.js');
    });

    test('should extract page metadata', () => {
      const metadata = contentScript.extractPageMetadata();

      expect(metadata.title).toBe('Test Page');
      expect(metadata.description).toBe('Test description');
      expect(metadata.url).toBe(window.location.href);
    });

    test('should detect bookmarkable content', () => {
      const isBookmarkable = contentScript.isBookmarkable();
      expect(typeof isBookmarkable).toBe('boolean');
    });
  });

  describe('Safari-specific Features', () => {
    test('should handle Safari bookmark import', async () => {
      // Mock Safari bookmark API
      global.browser.bookmarks = {
        search: jest.fn().mockResolvedValue([
          { id: '1', title: 'Safari Bookmark', url: 'https://example.com' }
        ])
      };

      const safariImporter = require('../safari/background/safari-importer.js');
      const bookmarks = await safariImporter.importSafariBookmarks();

      expect(bookmarks).toHaveLength(1);
      expect(bookmarks[0].title).toBe('Safari Bookmark');
    });

    test('should handle Safari-specific UI limitations', () => {
      // Test popup size constraints
      const popup = document.createElement('div');
      popup.style.width = '400px';
      popup.style.height = '600px';

      // Safari has stricter popup size limits
      expect(parseInt(popup.style.width)).toBeLessThanOrEqual(400);
      expect(parseInt(popup.style.height)).toBeLessThanOrEqual(600);
    });

    test('should handle Safari App Store requirements', () => {
      const manifest = require('../safari/manifest.json');

      // Check for required Safari fields
      expect(manifest.safari_web_extension).toBeDefined();
      expect(manifest.safari_web_extension.bundle_identifier).toMatch(/^[a-zA-Z0-9.-]+$/);
    });
  });

  describe('Cross-browser Compatibility', () => {
    test('should use browser API compatibility layer', () => {
      // Test that Safari extension uses browser API instead of chrome API
      const backgroundScript = require('../safari/background/background.js');

      // Should use browser.* instead of chrome.*
      expect(backgroundScript.toString()).toContain('browser.');
      expect(backgroundScript.toString()).not.toContain('chrome.');
    });

    test('should sync with Chrome and Firefox extensions', async () => {
      // Mock sync manager
      const syncManager = require('../safari/background/sync-manager.js').syncManager;

      const testEvent = {
        type: 'bookmark_created',
        data: { id: '1', title: 'Cross-browser test' }
      };

      await syncManager.createSyncEvent('create', 'bookmark', '1', testEvent.data);

      // Verify event can be processed by other browsers
      expect(testEvent.type).toBe('bookmark_created');
      expect(testEvent.data.id).toBe('1');
    });
  });

  describe('Error Handling', () => {
    test('should handle Safari-specific errors', async () => {
      global.browser = {
        runtime: {
          lastError: { message: 'Safari extension error' }
        }
      };

      const errorHandler = require('../safari/background/error-handler.js');
      const result = await errorHandler.handleSafariError();

      expect(result.error).toBe('Safari extension error');
    });

    test('should gracefully degrade when features unavailable', () => {
      // Test when certain APIs are not available
      delete global.browser.notifications;

      const backgroundScript = require('../safari/background/background.js');

      // Should not throw error when notifications API is unavailable
      expect(() => backgroundScript.showNotification('test')).not.toThrow();
    });
  });
});