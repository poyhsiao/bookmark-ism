// Safari Extension Options Page Script
// Settings and configuration interface for Safari Web Extension

class SafariOptions {
  constructor() {
    this.preferences = {};
    this.currentUser = null;
    this.isImporting = false;
  }

  /**
   * Initialize options page
   */
  async init() {
    try {
      // Load current preferences
      await this.loadPreferences();

      // Load user info
      await this.loadUserInfo();

      // Setup event listeners
      this.setupEventListeners();

      // Load data stats
      await this.loadDataStats();

      // Update sync status
      await this.updateSyncStatus();

      console.log('Safari Options: Initialized successfully');

    } catch (error) {
      console.error('Safari Options: Failed to initialize:', error);
      this.showToast('Failed to initialize settings', 'error');
    }
  }

  /**
   * Setup event listeners
   */
  setupEventListeners() {
    // Theme selection
    document.getElementById('theme-select').addEventListener('change', (e) => {
      this.setPreference('theme', e.target.value);
    });

    // View mode selection
    document.getElementById('view-mode-select').addEventListener('change', (e) => {
      this.setPreference('viewMode', e.target.value);
    });

    // Grid size selection
    document.getElementById('grid-size-select').addEventListener('change', (e) => {
      this.setPreference('gridSize', e.target.value);
    });

    // Sync preferences
    document.getElementById('auto-sync-checkbox').addEventListener('change', (e) => {
      this.setPreference('autoSync', e.target.checked);
    });

    document.getElementById('real-time-sync-checkbox').addEventListener('change', (e) => {
      this.setPreference('realTimeSync', e.target.checked);
    });

    document.getElementById('sync-interval-select').addEventListener('change', (e) => {
      this.setPreference('syncInterval', parseInt(e.target.value));
    });

    // Notification preferences
    document.getElementById('notifications-checkbox').addEventListener('change', (e) => {
      this.setPreference('notifications', e.target.checked);
    });

    document.getElementById('sync-notifications-checkbox').addEventListener('change', (e) => {
      this.setPreference('syncNotifications', e.target.checked);
    });

    document.getElementById('error-notifications-checkbox').addEventListener('change', (e) => {
      this.setPreference('errorNotifications', e.target.checked);
    });

    // Safari integration
    document.getElementById('context-menu-checkbox').addEventListener('change', (e) => {
      this.setPreference('contextMenu', e.target.checked);
    });

    document.getElementById('auto-import-checkbox').addEventListener('change', (e) => {
      this.setPreference('autoImport', e.target.checked);
    });

    // Advanced settings
    document.getElementById('api-endpoint-input').addEventListener('change', (e) => {
      this.setPreference('apiEndpoint', e.target.value);
    });

    document.getElementById('debug-mode-checkbox').addEventListener('change', (e) => {
      this.setPreference('debugMode', e.target.checked);
    });

    // Action buttons
    document.getElementById('logout-button').addEventListener('click', () => {
      this.handleLogout();
    });

    document.getElementById('login-button').addEventListener('click', () => {
      this.handleLogin();
    });

    document.getElementById('force-sync-button').addEventListener('click', () => {
      this.handleForceSync();
    });

    document.getElementById('import-safari-button').addEventListener('click', () => {
      this.handleImportSafari();
    });

    document.getElementById('export-data-button').addEventListener('click', () => {
      this.handleExportData();
    });

    document.getElementById('import-data-button').addEventListener('click', () => {
      this.handleImportData();
    });

    document.getElementById('clear-cache-button').addEventListener('click', () => {
      this.handleClearCache();
    });

    document.getElementById('reset-settings-button').addEventListener('click', () => {
      this.handleResetSettings();
    });

    document.getElementById('view-logs-button').addEventListener('click', () => {
      this.showLogsModal();
    });

    // Modal controls
    document.getElementById('close-import-modal').addEventListener('click', () => {
      this.hideImportModal();
    });

    document.getElementById('cancel-import').addEventListener('click', () => {
      this.cancelImport();
    });

    document.getElementById('close-logs-modal').addEventListener('click', () => {
      this.hideLogsModal();
    });

    document.getElementById('clear-logs-button').addEventListener('click', () => {
      this.clearLogs();
    });

    document.getElementById('export-logs-button').addEventListener('click', () => {
      this.exportLogs();
    });

    // Overlay click to close modals
    document.getElementById('overlay').addEventListener('click', () => {
      this.hideAllModals();
    });

    // Links
    document.getElementById('help-link').addEventListener('click', (e) => {
      e.preventDefault();
      this.openHelpPage();
    });

    document.getElementById('privacy-link').addEventListener('click', (e) => {
      e.preventDefault();
      this.openPrivacyPage();
    });

    document.getElementById('github-link').addEventListener('click', (e) => {
      e.preventDefault();
      this.openGitHubPage();
    });
  }

  /**
   * Load user preferences
   */
  async loadPreferences() {
    try {
      const result = await this.sendMessage({ type: 'GET_PREFERENCES' });

      if (result.success) {
        this.preferences = result.data;
      } else {
        // Use defaults
        this.preferences = this.getDefaultPreferences();
      }

      this.updatePreferencesUI();

    } catch (error) {
      console.error('Safari Options: Failed to load preferences:', error);
      this.preferences = this.getDefaultPreferences();
      this.updatePreferencesUI();
    }
  }

  /**
   * Get default preferences
   */
  getDefaultPreferences() {
    return {
      theme: 'auto',
      viewMode: 'grid',
      gridSize: 'medium',
      autoSync: true,
      realTimeSync: true,
      syncInterval: 60,
      notifications: true,
      syncNotifications: true,
      errorNotifications: true,
      contextMenu: true,
      autoImport: false,
      apiEndpoint: 'http://localhost:8080',
      debugMode: false
    };
  }

  /**
   * Update preferences UI
   */
  updatePreferencesUI() {
    // Appearance
    document.getElementById('theme-select').value = this.preferences.theme || 'auto';
    document.getElementById('view-mode-select').value = this.preferences.viewMode || 'grid';
    document.getElementById('grid-size-select').value = this.preferences.gridSize || 'medium';

    // Sync
    document.getElementById('auto-sync-checkbox').checked = this.preferences.autoSync !== false;
    document.getElementById('real-time-sync-checkbox').checked = this.preferences.realTimeSync !== false;
    document.getElementById('sync-interval-select').value = this.preferences.syncInterval || 60;

    // Notifications
    document.getElementById('notifications-checkbox').checked = this.preferences.notifications !== false;
    document.getElementById('sync-notifications-checkbox').checked = this.preferences.syncNotifications !== false;
    document.getElementById('error-notifications-checkbox').checked = this.preferences.errorNotifications !== false;

    // Safari integration
    document.getElementById('context-menu-checkbox').checked = this.preferences.contextMenu !== false;
    document.getElementById('auto-import-checkbox').checked = this.preferences.autoImport === true;

    // Advanced
    document.getElementById('api-endpoint-input').value = this.preferences.apiEndpoint || 'http://localhost:8080';
    document.getElementById('debug-mode-checkbox').checked = this.preferences.debugMode === true;
  }

  /**
   * Set preference value
   */
  async setPreference(key, value) {
    try {
      this.preferences[key] = value;

      const result = await this.sendMessage({
        type: 'SET_PREFERENCES',
        data: this.preferences
      });

      if (result.success) {
        this.showToast(`${key} updated successfully`, 'success');
      } else {
        throw new Error(result.error || 'Failed to save preference');
      }

    } catch (error) {
      console.error('Safari Options: Failed to set preference:', error);
      this.showToast('Failed to save setting', 'error');
    }
  }

  /**
   * Load user information
   */
  async loadUserInfo() {
    try {
      const authState = await this.sendMessage({ type: 'GET_AUTH_STATE' });

      if (authState.authenticated) {
        this.currentUser = authState.user;
        this.updateUserInfoUI();
      } else {
        this.showLoginState();
      }

    } catch (error) {
      console.error('Safari Options: Failed to load user info:', error);
      this.showLoginState();
    }
  }

  /**
   * Update user info UI
   */
  updateUserInfoUI() {
    if (this.currentUser) {
      const name = this.currentUser.name || this.currentUser.email;
      const email = this.currentUser.email;
      const initial = name.charAt(0).toUpperCase();

      document.getElementById('user-name').textContent = name;
      document.getElementById('user-email').textContent = email;
      document.getElementById('user-initial').textContent = initial;

      document.getElementById('logout-button').style.display = 'block';
      document.getElementById('login-button').style.display = 'none';
    }
  }

  /**
   * Show login state
   */
  showLoginState() {
    document.getElementById('user-name').textContent = 'Not signed in';
    document.getElementById('user-email').textContent = '';
    document.getElementById('user-initial').textContent = '?';

    document.getElementById('logout-button').style.display = 'none';
    document.getElementById('login-button').style.display = 'block';
  }

  /**
   * Load data statistics
   */
  async loadDataStats() {
    try {
      // Get bookmark count
      const bookmarksResult = await this.sendMessage({ type: 'GET_BOOKMARKS' });
      const bookmarkCount = bookmarksResult.success ? bookmarksResult.total || 0 : 0;

      // Get collection count (placeholder)
      const collectionCount = 0; // Would need API endpoint

      // Get storage usage (placeholder)
      const storageUsed = '0 KB'; // Would need storage API

      document.getElementById('bookmark-count').textContent = bookmarkCount;
      document.getElementById('collection-count').textContent = collectionCount;
      document.getElementById('storage-used').textContent = storageUsed;

    } catch (error) {
      console.error('Safari Options: Failed to load data stats:', error);
    }
  }

  /**
   * Update sync status
   */
  async updateSyncStatus() {
    try {
      const status = await this.sendMessage({ type: 'GET_SYNC_STATUS' });

      const statusDot = document.getElementById('sync-status-dot');
      const statusText = document.getElementById('sync-status-text');

      if (status.connected) {
        statusDot.className = 'status-dot online';
        statusText.textContent = 'Connected';
      } else {
        statusDot.className = 'status-dot';
        statusText.textContent = 'Offline';
      }

    } catch (error) {
      console.error('Safari Options: Failed to get sync status:', error);

      const statusDot = document.getElementById('sync-status-dot');
      const statusText = document.getElementById('sync-status-text');

      statusDot.className = 'status-dot';
      statusText.textContent = 'Unknown';
    }
  }

  /**
   * Handle logout
   */
  async handleLogout() {
    if (!confirm('Are you sure you want to logout?')) {
      return;
    }

    try {
      const result = await this.sendMessage({ type: 'LOGOUT' });

      if (result.success) {
        this.currentUser = null;
        this.showLoginState();
        this.showToast('Logged out successfully', 'success');
      } else {
        throw new Error(result.error || 'Logout failed');
      }

    } catch (error) {
      console.error('Safari Options: Logout failed:', error);
      this.showToast('Logout failed', 'error');
    }
  }

  /**
   * Handle login (redirect to popup)
   */
  handleLogin() {
    // Open popup for login
    browser.browserAction.openPopup();
  }

  /**
   * Handle force sync
   */
  async handleForceSync() {
    try {
      const statusDot = document.getElementById('sync-status-dot');
      const statusText = document.getElementById('sync-status-text');

      statusDot.className = 'status-dot syncing';
      statusText.textContent = 'Syncing...';

      const result = await this.sendMessage({ type: 'FORCE_SYNC' });

      if (result.success) {
        this.showToast('Sync completed successfully', 'success');

        // Refresh data stats
        setTimeout(() => {
          this.loadDataStats();
          this.updateSyncStatus();
        }, 2000);

      } else {
        throw new Error(result.error || 'Sync failed');
      }

    } catch (error) {
      console.error('Safari Options: Sync failed:', error);
      this.showToast('Sync failed', 'error');
      this.updateSyncStatus();
    }
  }

  /**
   * Handle Safari bookmark import
   */
  async handleImportSafari() {
    if (this.isImporting) return;

    try {
      this.isImporting = true;
      this.showImportModal();

      const result = await this.sendMessage({ type: 'IMPORT_SAFARI_BOOKMARKS' });

      if (result.success) {
        this.hideImportModal();
        this.showToast(`Imported ${result.stats.processed} bookmarks from Safari`, 'success');

        // Refresh data stats
        this.loadDataStats();
      } else {
        this.hideImportModal();
        this.showToast(result.error || 'Import failed', 'error');
      }

    } catch (error) {
      console.error('Safari Options: Import failed:', error);
      this.hideImportModal();
      this.showToast('Import failed', 'error');
    } finally {
      this.isImporting = false;
    }
  }

  /**
   * Handle data export
   */
  async handleExportData() {
    try {
      // Get bookmarks data
      const result = await this.sendMessage({ type: 'GET_BOOKMARKS' });

      if (result.success) {
        const exportData = {
          timestamp: new Date().toISOString(),
          version: '1.0.0',
          bookmarks: result.data || [],
          preferences: this.preferences
        };

        const blob = new Blob([JSON.stringify(exportData, null, 2)], {
          type: 'application/json'
        });

        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `bookmark-sync-export-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showToast('Data exported successfully', 'success');
      } else {
        throw new Error(result.error || 'Export failed');
      }

    } catch (error) {
      console.error('Safari Options: Export failed:', error);
      this.showToast('Export failed', 'error');
    }
  }

  /**
   * Handle data import
   */
  handleImportData() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';

    input.onchange = async (e) => {
      const file = e.target.files[0];
      if (!file) return;

      try {
        const text = await file.text();
        const data = JSON.parse(text);

        // Validate data structure
        if (!data.bookmarks || !Array.isArray(data.bookmarks)) {
          throw new Error('Invalid data format');
        }

        // Import bookmarks (would need API endpoint)
        this.showToast('Data import not yet implemented', 'warning');

      } catch (error) {
        console.error('Safari Options: Import failed:', error);
        this.showToast('Import failed: Invalid file format', 'error');
      }
    };

    input.click();
  }

  /**
   * Handle clear cache
   */
  async handleClearCache() {
    if (!confirm('Are you sure you want to clear the cache? This will remove locally stored bookmark data.')) {
      return;
    }

    try {
      // Clear cache (would need API endpoint)
      await browser.storage.local.clear();

      this.showToast('Cache cleared successfully', 'success');

      // Refresh data stats
      this.loadDataStats();

    } catch (error) {
      console.error('Safari Options: Clear cache failed:', error);
      this.showToast('Failed to clear cache', 'error');
    }
  }

  /**
   * Handle reset settings
   */
  async handleResetSettings() {
    if (!confirm('Are you sure you want to reset all settings to default values?')) {
      return;
    }

    try {
      this.preferences = this.getDefaultPreferences();

      const result = await this.sendMessage({
        type: 'SET_PREFERENCES',
        data: this.preferences
      });

      if (result.success) {
        this.updatePreferencesUI();
        this.showToast('Settings reset successfully', 'success');
      } else {
        throw new Error(result.error || 'Reset failed');
      }

    } catch (error) {
      console.error('Safari Options: Reset failed:', error);
      this.showToast('Failed to reset settings', 'error');
    }
  }

  /**
   * Show logs modal
   */
  async showLogsModal() {
    try {
      const result = await this.sendMessage({ type: 'GET_ERROR_LOG' });

      const logsContent = document.getElementById('logs-content');

      if (result && result.length > 0) {
        logsContent.innerHTML = result.map(log => `
          <div class="log-entry">
            <span class="log-timestamp">${new Date(log.timestamp).toLocaleString()}</span>
            <span class="log-level ${log.type}">${log.type.toUpperCase()}</span>
            <span class="log-message">${log.message}</span>
          </div>
        `).join('');
      } else {
        logsContent.innerHTML = '<p class="no-logs">No error logs found</p>';
      }

      this.showModal('logs-modal');

    } catch (error) {
      console.error('Safari Options: Failed to load logs:', error);
      this.showToast('Failed to load logs', 'error');
    }
  }

  /**
   * Clear logs
   */
  async clearLogs() {
    try {
      await this.sendMessage({ type: 'CLEAR_ERROR_LOG' });

      const logsContent = document.getElementById('logs-content');
      logsContent.innerHTML = '<p class="no-logs">No error logs found</p>';

      this.showToast('Logs cleared successfully', 'success');

    } catch (error) {
      console.error('Safari Options: Failed to clear logs:', error);
      this.showToast('Failed to clear logs', 'error');
    }
  }

  /**
   * Export logs
   */
  async exportLogs() {
    try {
      const result = await this.sendMessage({ type: 'GET_ERROR_LOG' });

      if (result && result.length > 0) {
        const logText = result.map(log =>
          `[${new Date(log.timestamp).toISOString()}] ${log.type.toUpperCase()}: ${log.message}`
        ).join('\n');

        const blob = new Blob([logText], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `safari-extension-logs-${new Date().toISOString().split('T')[0]}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showToast('Logs exported successfully', 'success');
      } else {
        this.showToast('No logs to export', 'warning');
      }

    } catch (error) {
      console.error('Safari Options: Failed to export logs:', error);
      this.showToast('Failed to export logs', 'error');
    }
  }

  /**
   * Cancel import
   */
  cancelImport() {
    this.isImporting = false;
    this.hideImportModal();
  }

  /**
   * Open help page
   */
  openHelpPage() {
    browser.tabs.create({ url: 'https://github.com/your-repo/bookmark-sync-service/wiki' });
  }

  /**
   * Open privacy page
   */
  openPrivacyPage() {
    browser.tabs.create({ url: 'https://github.com/your-repo/bookmark-sync-service/blob/main/PRIVACY.md' });
  }

  /**
   * Open GitHub page
   */
  openGitHubPage() {
    browser.tabs.create({ url: 'https://github.com/your-repo/bookmark-sync-service' });
  }

  /**
   * Show modal
   */
  showModal(modalId) {
    document.getElementById(modalId).style.display = 'flex';
    document.getElementById('overlay').style.display = 'block';
  }

  /**
   * Hide modal
   */
  hideModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
    document.getElementById('overlay').style.display = 'none';
  }

  /**
   * Show import modal
   */
  showImportModal() {
    this.showModal('import-modal');
  }

  /**
   * Hide import modal
   */
  hideImportModal() {
    this.hideModal('import-modal');
  }

  /**
   * Hide logs modal
   */
  hideLogsModal() {
    this.hideModal('logs-modal');
  }

  /**
   * Hide all modals
   */
  hideAllModals() {
    document.querySelectorAll('.modal').forEach(modal => {
      modal.style.display = 'none';
    });
    document.getElementById('overlay').style.display = 'none';
  }

  /**
   * Show toast notification
   */
  showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;

    const container = document.getElementById('toast-container');
    container.appendChild(toast);

    // Auto remove after 3 seconds
    setTimeout(() => {
      if (toast.parentNode) {
        toast.parentNode.removeChild(toast);
      }
    }, 3000);
  }

  /**
   * Send message to background script
   */
  async sendMessage(message) {
    return new Promise((resolve) => {
      browser.runtime.sendMessage(message, resolve);
    });
  }
}

// Initialize options page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  const options = new SafariOptions();
  options.init();
});

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = SafariOptions;
}