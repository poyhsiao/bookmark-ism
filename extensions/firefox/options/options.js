// Options page script
import { formatTimestamp } from '../../shared/utils.js';
import { STORAGE_KEYS, UI_CONFIG } from '../../shared/constants.js';

class OptionsPage {
  constructor() {
    this.settings = {
      autoSync: true,
      syncOnStartup: true,
      syncInterval: 'realtime',
      viewMode: UI_CONFIG.VIEW_MODES.GRID,
      gridSize: UI_CONFIG.GRID_SIZES.MEDIUM,
      theme: UI_CONFIG.THEMES.AUTO,
      showFavicons: true,
      showDescriptions: true,
      privateByDefault: false,
      analyticsEnabled: true,
      cacheLimit: 100,
      apiUrl: 'http://localhost:8080',
      debugMode: false
    };

    this.init();
  }

  async init() {
    await this.loadSettings();
    await this.loadUserData();
    this.bindEvents();
    this.updateUI();
    this.updateSyncStatus();
  }

  async loadSettings() {
    try {
      const stored = await chrome.storage.sync.get(Object.keys(this.settings));
      this.settings = { ...this.settings, ...stored };
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  }

  async saveSettings() {
    try {
      await chrome.storage.sync.set(this.settings);
      this.showNotification('Settings saved successfully', 'success');
    } catch (error) {
      console.error('Failed to save settings:', error);
      this.showNotification('Failed to save settings', 'error');
    }
  }

  async loadUserData() {
    try {
      const response = await this.sendMessage({ type: 'GET_AUTH_STATE' });

      if (response.authenticated && response.user) {
        this.showUserInfo(response.user);
      } else {
        this.showLoginPrompt();
      }
    } catch (error) {
      console.error('Failed to load user data:', error);
      this.showLoginPrompt();
    }
  }

  showUserInfo(user) {
    document.getElementById('account-details').classList.remove('hidden');
    document.getElementById('login-prompt').classList.add('hidden');

    document.getElementById('user-name').textContent = user.name || 'User';
    document.getElementById('user-email').textContent = user.email;

    if (user.created_at) {
      const joinDate = new Date(user.created_at).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long'
      });
      document.getElementById('user-joined').textContent = joinDate;
    }

    if (user.avatar_url) {
      const avatar = document.getElementById('user-avatar');
      avatar.src = user.avatar_url;
      avatar.style.display = 'block';
      document.querySelector('.avatar-placeholder').style.display = 'none';
    }
  }

  showLoginPrompt() {
    document.getElementById('account-details').classList.add('hidden');
    document.getElementById('login-prompt').classList.remove('hidden');
  }

  bindEvents() {
    // Account actions
    document.getElementById('login-btn').addEventListener('click', () => {
      const api = typeof browser !== 'undefined' ? browser : chrome;
      api.runtime.openOptionsPage();
      // In a real implementation, this would open the login flow
    });

    document.getElementById('logout-btn').addEventListener('click', async () => {
      try {
        await this.sendMessage({ type: 'LOGOUT' });
        this.showLoginPrompt();
      } catch (error) {
        console.error('Logout failed:', error);
      }
    });

    // Sync settings
    document.getElementById('auto-sync').addEventListener('change', (e) => {
      this.settings.autoSync = e.target.checked;
      this.saveSettings();
    });

    document.getElementById('sync-on-startup').addEventListener('change', (e) => {
      this.settings.syncOnStartup = e.target.checked;
      this.saveSettings();
    });

    document.getElementById('sync-interval').addEventListener('change', (e) => {
      this.settings.syncInterval = e.target.value;
      this.saveSettings();
    });

    document.getElementById('force-sync-btn').addEventListener('click', async () => {
      try {
        await this.sendMessage({ type: 'FORCE_SYNC' });
        this.showNotification('Sync initiated', 'success');
        this.updateSyncStatus();
      } catch (error) {
        this.showNotification('Sync failed', 'error');
      }
    });

    // Display settings
    document.querySelectorAll('input[name="view-mode"]').forEach(radio => {
      radio.addEventListener('change', (e) => {
        if (e.target.checked) {
          this.settings.viewMode = e.target.value;
          this.saveSettings();
        }
      });
    });

    document.getElementById('grid-size').addEventListener('change', (e) => {
      this.settings.gridSize = e.target.value;
      this.saveSettings();
    });

    document.getElementById('theme').addEventListener('change', (e) => {
      this.settings.theme = e.target.value;
      this.saveSettings();
      this.applyTheme();
    });

    document.getElementById('show-favicons').addEventListener('change', (e) => {
      this.settings.showFavicons = e.target.checked;
      this.saveSettings();
    });

    document.getElementById('show-descriptions').addEventListener('change', (e) => {
      this.settings.showDescriptions = e.target.checked;
      this.saveSettings();
    });

    // Privacy settings
    document.getElementById('private-by-default').addEventListener('change', (e) => {
      this.settings.privateByDefault = e.target.checked;
      this.saveSettings();
    });

    document.getElementById('analytics-enabled').addEventListener('change', (e) => {
      this.settings.analyticsEnabled = e.target.checked;
      this.saveSettings();
    });

    // Data management
    document.getElementById('cache-limit').addEventListener('change', (e) => {
      this.settings.cacheLimit = e.target.value === 'unlimited' ? -1 : parseInt(e.target.value);
      this.saveSettings();
    });

    document.getElementById('export-btn').addEventListener('click', () => this.exportBookmarks());
    document.getElementById('import-btn').addEventListener('click', () => this.importBookmarks());
    document.getElementById('clear-cache-btn').addEventListener('click', () => this.clearCache());

    // Advanced settings
    document.getElementById('api-url').addEventListener('change', (e) => {
      this.settings.apiUrl = e.target.value;
      this.saveSettings();
    });

    document.getElementById('debug-mode').addEventListener('change', (e) => {
      this.settings.debugMode = e.target.checked;
      this.saveSettings();
    });

    document.getElementById('reset-settings-btn').addEventListener('click', () => this.resetSettings());

    // File import
    document.getElementById('import-file').addEventListener('change', (e) => {
      this.handleFileImport(e.target.files[0]);
    });

    // Footer links
    document.getElementById('help-link').addEventListener('click', (e) => {
      e.preventDefault();
      const api = typeof browser !== 'undefined' ? browser : chrome;
      api.tabs.create({ url: 'https://github.com/yourusername/bookmark-sync-service/wiki' });
    });

    document.getElementById('privacy-link').addEventListener('click', (e) => {
      e.preventDefault();
      const api = typeof browser !== 'undefined' ? browser : chrome;
      api.tabs.create({ url: 'https://github.com/yourusername/bookmark-sync-service/blob/main/PRIVACY.md' });
    });

    document.getElementById('feedback-link').addEventListener('click', (e) => {
      e.preventDefault();
      const api = typeof browser !== 'undefined' ? browser : chrome;
      api.tabs.create({ url: 'https://github.com/yourusername/bookmark-sync-service/issues' });
    });
  }

  updateUI() {
    // Sync settings
    document.getElementById('auto-sync').checked = this.settings.autoSync;
    document.getElementById('sync-on-startup').checked = this.settings.syncOnStartup;
    document.getElementById('sync-interval').value = this.settings.syncInterval;

    // Display settings
    document.querySelector(`input[name="view-mode"][value="${this.settings.viewMode}"]`).checked = true;
    document.getElementById('grid-size').value = this.settings.gridSize;
    document.getElementById('theme').value = this.settings.theme;
    document.getElementById('show-favicons').checked = this.settings.showFavicons;
    document.getElementById('show-descriptions').checked = this.settings.showDescriptions;

    // Privacy settings
    document.getElementById('private-by-default').checked = this.settings.privateByDefault;
    document.getElementById('analytics-enabled').checked = this.settings.analyticsEnabled;

    // Data management
    const cacheValue = this.settings.cacheLimit === -1 ? 'unlimited' : this.settings.cacheLimit.toString();
    document.getElementById('cache-limit').value = cacheValue;

    // Advanced settings
    document.getElementById('api-url').value = this.settings.apiUrl;
    document.getElementById('debug-mode').checked = this.settings.debugMode;

    // Apply theme
    this.applyTheme();

    // Update storage info
    this.updateStorageInfo();
  }

  async updateSyncStatus() {
    try {
      const status = await this.sendMessage({ type: 'GET_SYNC_STATUS' });
      const statusEl = document.getElementById('sync-status');

      if (status.connected) {
        statusEl.textContent = 'Online';
        statusEl.className = 'status-value online';
      } else {
        statusEl.textContent = 'Offline';
        statusEl.className = 'status-value offline';
      }
    } catch (error) {
      console.error('Failed to get sync status:', error);
      const statusEl = document.getElementById('sync-status');
      statusEl.textContent = 'Unknown';
      statusEl.className = 'status-value offline';
    }
  }

  async updateStorageInfo() {
    try {
      const api = typeof browser !== 'undefined' ? browser : chrome;
      const usage = await api.storage.local.getBytesInUse();
      const quota = api.storage.local.QUOTA_BYTES || 10485760; // 10MB default

      const usedMB = (usage / 1024 / 1024).toFixed(1);
      const quotaMB = (quota / 1024 / 1024).toFixed(1);
      const percentUsed = (usage / quota) * 100;

      document.getElementById('storage-used').style.width = `${Math.min(percentUsed, 100)}%`;
      document.getElementById('storage-text').textContent = `${usedMB} MB used of ${quotaMB} MB available`;
    } catch (error) {
      console.error('Failed to get storage info:', error);
    }
  }

  applyTheme() {
    const theme = this.settings.theme;
    const body = document.body;

    body.classList.remove('theme-light', 'theme-dark');

    if (theme === 'light') {
      body.classList.add('theme-light');
    } else if (theme === 'dark') {
      body.classList.add('theme-dark');
    } else {
      // Auto theme - use system preference
      if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
        body.classList.add('theme-dark');
      } else {
        body.classList.add('theme-light');
      }
    }
  }

  async exportBookmarks() {
    try {
      const response = await this.sendMessage({ type: 'GET_BOOKMARKS' });

      if (response.bookmarks) {
        const data = {
          version: '1.0',
          exported_at: new Date().toISOString(),
          bookmarks: response.bookmarks
        };

        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = `bookmarks-export-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);

        URL.revokeObjectURL(url);

        this.showNotification('Bookmarks exported successfully', 'success');
      }
    } catch (error) {
      console.error('Export failed:', error);
      this.showNotification('Export failed', 'error');
    }
  }

  importBookmarks() {
    document.getElementById('import-file').click();
  }

  async handleFileImport(file) {
    if (!file) return;

    try {
      const text = await file.text();
      let data;

      if (file.name.endsWith('.json')) {
        data = JSON.parse(text);
      } else if (file.name.endsWith('.html')) {
        // Parse HTML bookmark file (basic implementation)
        data = this.parseHtmlBookmarks(text);
      } else {
        throw new Error('Unsupported file format');
      }

      if (data.bookmarks && Array.isArray(data.bookmarks)) {
        let imported = 0;

        for (const bookmark of data.bookmarks) {
          try {
            await this.sendMessage({
              type: 'CREATE_BOOKMARK',
              data: {
                url: bookmark.url,
                title: bookmark.title,
                description: bookmark.description || '',
                tags: bookmark.tags || []
              }
            });
            imported++;
          } catch (error) {
            console.warn('Failed to import bookmark:', bookmark.url, error);
          }
        }

        this.showNotification(`Imported ${imported} bookmarks`, 'success');
      } else {
        throw new Error('Invalid bookmark file format');
      }
    } catch (error) {
      console.error('Import failed:', error);
      this.showNotification('Import failed: ' + error.message, 'error');
    }
  }

  parseHtmlBookmarks(html) {
    // Basic HTML bookmark parser
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const links = doc.querySelectorAll('a[href]');

    const bookmarks = Array.from(links).map(link => ({
      url: link.href,
      title: link.textContent.trim() || link.href,
      description: '',
      tags: []
    }));

    return { bookmarks };
  }

  async clearCache() {
    if (!confirm('Are you sure you want to clear the local cache? This will remove all cached bookmarks.')) {
      return;
    }

    try {
      const api = typeof browser !== 'undefined' ? browser : chrome;
      await api.storage.local.remove(['bookmarks_cache']);
      this.showNotification('Cache cleared successfully', 'success');
      this.updateStorageInfo();
    } catch (error) {
      console.error('Failed to clear cache:', error);
      this.showNotification('Failed to clear cache', 'error');
    }
  }

  async resetSettings() {
    if (!confirm('Are you sure you want to reset all settings to their default values?')) {
      return;
    }

    try {
      const api = typeof browser !== 'undefined' ? browser : chrome;
      await api.storage.sync.clear();

      // Reset to defaults
      this.settings = {
        autoSync: true,
        syncOnStartup: true,
        syncInterval: 'realtime',
        viewMode: UI_CONFIG.VIEW_MODES.GRID,
        gridSize: UI_CONFIG.GRID_SIZES.MEDIUM,
        theme: UI_CONFIG.THEMES.AUTO,
        showFavicons: true,
        showDescriptions: true,
        privateByDefault: false,
        analyticsEnabled: true,
        cacheLimit: 100,
        apiUrl: 'http://localhost:8080',
        debugMode: false
      };

      await this.saveSettings();
      this.updateUI();

      this.showNotification('Settings reset to defaults', 'success');
    } catch (error) {
      console.error('Failed to reset settings:', error);
      this.showNotification('Failed to reset settings', 'error');
    }
  }

  showNotification(message, type = 'info') {
    // Simple notification system
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;

    notification.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      padding: 12px 20px;
      border-radius: 4px;
      color: white;
      font-weight: 500;
      z-index: 1000;
      animation: slideIn 0.3s ease;
    `;

    if (type === 'success') {
      notification.style.background = '#28a745';
    } else if (type === 'error') {
      notification.style.background = '#dc3545';
    } else {
      notification.style.background = '#007bff';
    }

    document.body.appendChild(notification);

    setTimeout(() => {
      notification.style.animation = 'slideOut 0.3s ease';
      setTimeout(() => {
        document.body.removeChild(notification);
      }, 300);
    }, 3000);
  }

  sendMessage(message) {
    return new Promise((resolve, reject) => {
      const api = typeof browser !== 'undefined' ? browser : chrome;
      api.runtime.sendMessage(message, (response) => {
        if (api.runtime.lastError) {
          reject(api.runtime.lastError);
        } else {
          resolve(response);
        }
      });
    });
  }
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
  @keyframes slideIn {
    from { transform: translateX(100%); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
  }

  @keyframes slideOut {
    from { transform: translateX(0); opacity: 1; }
    to { transform: translateX(100%); opacity: 0; }
  }

  .theme-dark {
    background: #1a1a1a;
    color: #e0e0e0;
  }

  .theme-dark .container {
    background: #2d2d2d;
  }

  .theme-dark .section h2 {
    color: #e0e0e0;
  }

  .theme-dark .setting-input,
  .theme-dark .setting-select {
    background: #3d3d3d;
    border-color: #555;
    color: #e0e0e0;
  }
`;
document.head.appendChild(style);

// Initialize options page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  new OptionsPage();
});