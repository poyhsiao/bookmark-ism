// Safari Extension Popup Script
// Main popup interface for Safari Web Extension

class SafariPopup {
  constructor() {
    this.currentView = 'grid';
    this.currentUser = null;
    this.bookmarks = [];
    this.filteredBookmarks = [];
    this.searchQuery = '';
    this.isImporting = false;
    this.preferences = {};
  }

  /**
   * Initialize popup
   */
  async init() {
    try {
      // Load preferences
      await this.loadPreferences();

      // Apply theme
      this.applyTheme();

      // Check authentication state
      const authState = await this.sendMessage({ type: 'GET_AUTH_STATE' });

      if (authState.authenticated) {
        this.currentUser = authState.user;
        await this.showBookmarksSection();
      } else {
        this.showAuthSection();
      }

      // Setup event listeners
      this.setupEventListeners();

      // Hide loading
      this.hideSection('loading-section');

    } catch (error) {
      console.error('Safari Popup: Failed to initialize:', error);
      this.showError('Failed to initialize popup');
    }
  }

  /**
   * Setup event listeners
   */
  setupEventListeners() {
    // Auth tabs
    document.getElementById('login-tab').addEventListener('click', () => {
      this.showAuthForm('login');
    });

    document.getElementById('register-tab').addEventListener('click', () => {
      this.showAuthForm('register');
    });

    // Auth forms
    document.getElementById('login-form').addEventListener('submit', (e) => {
      this.handleLogin(e);
    });

    document.getElementById('register-form').addEventListener('submit', (e) => {
      this.handleRegister(e);
    });

    // Header actions
    document.getElementById('sync-button').addEventListener('click', () => {
      this.handleSync();
    });

    document.getElementById('view-toggle').addEventListener('click', () => {
      this.toggleView();
    });

    document.getElementById('settings-button').addEventListener('click', () => {
      this.showSettingsSection();
    });

    // Search
    document.getElementById('search-input').addEventListener('input', (e) => {
      this.handleSearch(e.target.value);
    });

    document.getElementById('search-clear').addEventListener('click', () => {
      this.clearSearch();
    });

    // Quick actions
    document.getElementById('bookmark-current').addEventListener('click', () => {
      this.bookmarkCurrentPage();
    });

    document.getElementById('import-safari').addEventListener('click', () => {
      this.importSafariBookmarks();
    });

    // View mode buttons
    document.getElementById('grid-view').addEventListener('click', () => {
      this.setViewMode('grid');
    });

    document.getElementById('list-view').addEventListener('click', () => {
      this.setViewMode('list');
    });

    // Settings
    document.getElementById('back-to-bookmarks').addEventListener('click', () => {
      this.showBookmarksSection();
    });

    document.getElementById('theme-select').addEventListener('change', (e) => {
      this.setTheme(e.target.value);
    });

    document.getElementById('grid-size-select').addEventListener('change', (e) => {
      this.setGridSize(e.target.value);
    });

    document.getElementById('auto-sync-checkbox').addEventListener('change', (e) => {
      this.setAutoSync(e.target.checked);
    });

    document.getElementById('notifications-checkbox').addEventListener('change', (e) => {
      this.setNotifications(e.target.checked);
    });

    document.getElementById('clear-cache-button').addEventListener('click', () => {
      this.clearCache();
    });

    document.getElementById('export-data-button').addEventListener('click', () => {
      this.exportData();
    });

    // User actions
    document.getElementById('logout-button').addEventListener('click', () => {
      this.handleLogout();
    });

    // Import modal
    document.getElementById('cancel-import').addEventListener('click', () => {
      this.cancelImport();
    });

    // Listen for storage changes (sync notifications)
    if (browser.storage && browser.storage.onChanged) {
      browser.storage.onChanged.addListener((changes, namespace) => {
        if (namespace === 'local' && changes.sync_notification) {
          this.handleSyncNotification(changes.sync_notification.newValue);
        }
      });
    }
  }

  /**
   * Show authentication section
   */
  showAuthSection() {
    this.hideAllSections();
    this.showSection('auth-section');
  }

  /**
   * Show bookmarks section
   */
  async showBookmarksSection() {
    this.hideAllSections();
    this.showSection('bookmarks-section');

    // Update user info
    if (this.currentUser) {
      document.getElementById('user-name').textContent = this.currentUser.name || this.currentUser.email;
    }

    // Load bookmarks
    await this.loadBookmarks();

    // Update sync status
    this.updateSyncStatus();
  }

  /**
   * Show settings section
   */
  showSettingsSection() {
    this.hideAllSections();
    this.showSection('settings-section');

    // Load current settings
    this.loadSettingsForm();
  }

  /**
   * Handle login form submission
   */
  async handleLogin(event) {
    event.preventDefault();

    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    if (!email || !password) {
      this.showAuthError('Please fill in all fields');
      return;
    }

    try {
      const result = await this.sendMessage({
        type: 'LOGIN',
        email: email,
        password: password
      });

      if (result.success) {
        this.currentUser = result.user;
        await this.showBookmarksSection();
      } else {
        this.showAuthError(result.error || 'Login failed');
      }
    } catch (error) {
      console.error('Safari Popup: Login failed:', error);
      this.showAuthError('Login failed. Please try again.');
    }
  }

  /**
   * Handle register form submission
   */
  async handleRegister(event) {
    event.preventDefault();

    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;

    if (!name || !email || !password) {
      this.showAuthError('Please fill in all fields');
      return;
    }

    try {
      const result = await this.sendMessage({
        type: 'REGISTER',
        name: name,
        email: email,
        password: password
      });

      if (result.success) {
        this.currentUser = result.user;
        await this.showBookmarksSection();
      } else {
        this.showAuthError(result.error || 'Registration failed');
      }
    } catch (error) {
      console.error('Safari Popup: Registration failed:', error);
      this.showAuthError('Registration failed. Please try again.');
    }
  }

  /**
   * Handle logout
   */
  async handleLogout() {
    try {
      await this.sendMessage({ type: 'LOGOUT' });
      this.currentUser = null;
      this.bookmarks = [];
      this.showAuthSection();
    } catch (error) {
      console.error('Safari Popup: Logout failed:', error);
    }
  }

  /**
   * Load bookmarks
   */
  async loadBookmarks() {
    try {
      this.showElement('bookmarks-loading');
      this.hideElement('bookmarks-grid');
      this.hideElement('bookmarks-list');
      this.hideElement('empty-state');

      const result = await this.sendMessage({ type: 'GET_BOOKMARKS' });

      if (result.success) {
        this.bookmarks = result.data || [];
        this.filteredBookmarks = [...this.bookmarks];
        this.renderBookmarks();
        this.updateBookmarkCount();
      } else {
        throw new Error(result.error || 'Failed to load bookmarks');
      }

    } catch (error) {
      console.error('Safari Popup: Failed to load bookmarks:', error);
      this.showError('Failed to load bookmarks');
    } finally {
      this.hideElement('bookmarks-loading');
    }
  }

  /**
   * Render bookmarks
   */
  renderBookmarks() {
    const gridContainer = document.getElementById('bookmarks-grid');
    const listContainer = document.getElementById('bookmarks-list');

    // Clear containers
    gridContainer.innerHTML = '';
    listContainer.innerHTML = '';

    if (this.filteredBookmarks.length === 0) {
      this.showElement('empty-state');
      this.hideElement('bookmarks-grid');
      this.hideElement('bookmarks-list');
      return;
    }

    this.hideElement('empty-state');

    // Render bookmarks
    this.filteredBookmarks.forEach(bookmark => {
      const bookmarkElement = this.createBookmarkElement(bookmark);

      if (this.currentView === 'grid') {
        gridContainer.appendChild(bookmarkElement);
      } else {
        listContainer.appendChild(bookmarkElement);
      }
    });

    // Show appropriate container
    if (this.currentView === 'grid') {
      this.showElement('bookmarks-grid');
      this.hideElement('bookmarks-list');
    } else {
      this.showElement('bookmarks-list');
      this.hideElement('bookmarks-grid');
    }
  }

  /**
   * Create bookmark element
   */
  createBookmarkElement(bookmark) {
    const div = document.createElement('div');
    div.className = 'bookmark-item';
    div.dataset.bookmarkId = bookmark.id;

    const favicon = bookmark.favicon || this.getDefaultFavicon(bookmark.url);
    const title = bookmark.title || 'Untitled';
    const url = bookmark.url;

    div.innerHTML = `
      <img src="${favicon}" alt="" class="bookmark-favicon" onerror="this.src='data:image/svg+xml,<svg xmlns=\\"http://www.w3.org/2000/svg\\" width=\\"16\\" height=\\"16\\"><rect width=\\"16\\" height=\\"16\\" fill=\\"%23ddd\\"/></svg>'">
      <div class="bookmark-content">
        <div class="bookmark-title">${this.escapeHtml(title)}</div>
        <div class="bookmark-url">${this.escapeHtml(url)}</div>
      </div>
      <div class="bookmark-actions">
        <button class="btn btn-icon edit-bookmark" title="Edit">
          <span class="icon">✏️</span>
        </button>
        <button class="btn btn-icon delete-bookmark" title="Delete">
          <span class="icon">🗑️</span>
        </button>
      </div>
    `;

    // Add click handler to open bookmark
    div.addEventListener('click', (e) => {
      if (!e.target.closest('.bookmark-actions')) {
        this.openBookmark(bookmark);
      }
    });

    // Add action handlers
    div.querySelector('.edit-bookmark').addEventListener('click', (e) => {
      e.stopPropagation();
      this.editBookmark(bookmark);
    });

    div.querySelector('.delete-bookmark').addEventListener('click', (e) => {
      e.stopPropagation();
      this.deleteBookmark(bookmark);
    });

    return div;
  }

  /**
   * Open bookmark in new tab
   */
  async openBookmark(bookmark) {
    try {
      await browser.tabs.create({ url: bookmark.url });
      window.close();
    } catch (error) {
      console.error('Safari Popup: Failed to open bookmark:', error);
    }
  }

  /**
   * Bookmark current page
   */
  async bookmarkCurrentPage() {
    try {
      const tabs = await browser.tabs.query({ active: true, currentWindow: true });
      const currentTab = tabs[0];

      if (!currentTab || !currentTab.url || currentTab.url.startsWith('safari://')) {
        this.showError('Cannot bookmark this page');
        return;
      }

      const bookmarkData = {
        url: currentTab.url,
        title: currentTab.title || 'Untitled',
        description: '',
        tags: [],
        favicon: currentTab.favIconUrl || this.getDefaultFavicon(currentTab.url)
      };

      const result = await this.sendMessage({
        type: 'CREATE_BOOKMARK',
        data: bookmarkData
      });

      if (result.success) {
        await this.loadBookmarks();
        this.showSuccess('Bookmark saved successfully');
      } else {
        this.showError(result.error || 'Failed to save bookmark');
      }

    } catch (error) {
      console.error('Safari Popup: Failed to bookmark current page:', error);
      this.showError('Failed to bookmark current page');
    }
  }

  /**
   * Import Safari bookmarks
   */
  async importSafariBookmarks() {
    if (this.isImporting) return;

    try {
      this.isImporting = true;
      this.showImportModal();

      const result = await this.sendMessage({ type: 'IMPORT_SAFARI_BOOKMARKS' });

      if (result.success) {
        this.hideImportModal();
        await this.loadBookmarks();
        this.showSuccess(`Imported ${result.stats.processed} bookmarks from Safari`);
      } else {
        this.hideImportModal();
        this.showError(result.error || 'Import failed');
      }

    } catch (error) {
      console.error('Safari Popup: Import failed:', error);
      this.hideImportModal();
      this.showError('Import failed. Please try again.');
    } finally {
      this.isImporting = false;
    }
  }

  /**
   * Handle search
   */
  handleSearch(query) {
    this.searchQuery = query.toLowerCase();

    if (this.searchQuery) {
      this.filteredBookmarks = this.bookmarks.filter(bookmark =>
        bookmark.title.toLowerCase().includes(this.searchQuery) ||
        bookmark.url.toLowerCase().includes(this.searchQuery) ||
        (bookmark.description && bookmark.description.toLowerCase().includes(this.searchQuery))
      );
      this.showElement('search-clear');
    } else {
      this.filteredBookmarks = [...this.bookmarks];
      this.hideElement('search-clear');
    }

    this.renderBookmarks();
    this.updateBookmarkCount();
  }

  /**
   * Clear search
   */
  clearSearch() {
    document.getElementById('search-input').value = '';
    this.handleSearch('');
  }

  /**
   * Set view mode
   */
  setViewMode(mode) {
    this.currentView = mode;

    // Update buttons
    document.querySelectorAll('.view-btn').forEach(btn => {
      btn.classList.remove('active');
    });
    document.getElementById(`${mode}-view`).classList.add('active');

    // Update view toggle icon
    const viewToggle = document.getElementById('view-toggle');
    viewToggle.querySelector('.icon').textContent = mode === 'grid' ? '☰' : '⊞';

    // Re-render bookmarks
    this.renderBookmarks();

    // Save preference
    this.preferences.viewMode = mode;
    this.savePreferences();
  }

  /**
   * Toggle view mode
   */
  toggleView() {
    const newMode = this.currentView === 'grid' ? 'list' : 'grid';
    this.setViewMode(newMode);
  }

  /**
   * Update bookmark count
   */
  updateBookmarkCount() {
    const count = this.filteredBookmarks.length;
    const total = this.bookmarks.length;

    let text = `${count} bookmark${count !== 1 ? 's' : ''}`;
    if (count !== total) {
      text += ` of ${total}`;
    }

    document.getElementById('bookmark-count-text').textContent = text;
  }

  /**
   * Update sync status
   */
  async updateSyncStatus() {
    try {
      const status = await this.sendMessage({ type: 'GET_SYNC_STATUS' });

      const indicator = document.querySelector('.status-indicator');
      const statusText = document.querySelector('.status-text');

      if (status.connected) {
        indicator.className = 'status-indicator online';
        statusText.textContent = 'Online';
      } else {
        indicator.className = 'status-indicator offline';
        statusText.textContent = 'Offline';
      }

    } catch (error) {
      console.error('Safari Popup: Failed to get sync status:', error);
    }
  }

  /**
   * Handle sync button click
   */
  async handleSync() {
    try {
      const result = await this.sendMessage({ type: 'FORCE_SYNC' });

      if (result.success) {
        // Show syncing status temporarily
        const indicator = document.querySelector('.status-indicator');
        const statusText = document.querySelector('.status-text');

        indicator.className = 'status-indicator syncing';
        statusText.textContent = 'Syncing...';

        // Refresh bookmarks
        setTimeout(async () => {
          await this.loadBookmarks();
          this.updateSyncStatus();
        }, 2000);

      } else {
        this.showError('Sync failed: ' + result.error);
      }

    } catch (error) {
      console.error('Safari Popup: Sync failed:', error);
      this.showError('Sync failed');
    }
  }

  /**
   * Load preferences
   */
  async loadPreferences() {
    try {
      const result = await this.sendMessage({ type: 'GET_PREFERENCES' });

      if (result.success) {
        this.preferences = result.data;
      } else {
        // Use defaults
        this.preferences = {
          theme: 'auto',
          viewMode: 'grid',
          gridSize: 'medium',
          notifications: true,
          autoSync: true
        };
      }

    } catch (error) {
      console.error('Safari Popup: Failed to load preferences:', error);
      this.preferences = {};
    }
  }

  /**
   * Save preferences
   */
  async savePreferences() {
    try {
      await this.sendMessage({
        type: 'SET_PREFERENCES',
        data: this.preferences
      });
    } catch (error) {
      console.error('Safari Popup: Failed to save preferences:', error);
    }
  }

  /**
   * Apply theme
   */
  applyTheme() {
    const theme = this.preferences.theme || 'auto';
    document.body.setAttribute('data-theme', theme);
  }

  /**
   * Utility functions
   */
  async sendMessage(message) {
    return new Promise((resolve) => {
      browser.runtime.sendMessage(message, resolve);
    });
  }

  showSection(sectionId) {
    document.getElementById(sectionId).style.display = 'flex';
  }

  hideSection(sectionId) {
    document.getElementById(sectionId).style.display = 'none';
  }

  hideAllSections() {
    document.querySelectorAll('.section').forEach(section => {
      section.style.display = 'none';
    });
  }

  showElement(elementId) {
    document.getElementById(elementId).style.display = 'block';
  }

  hideElement(elementId) {
    document.getElementById(elementId).style.display = 'none';
  }

  showAuthForm(type) {
    // Update tabs
    document.querySelectorAll('.tab-button').forEach(btn => {
      btn.classList.remove('active');
    });
    document.getElementById(`${type}-tab`).classList.add('active');

    // Show/hide forms
    document.getElementById('login-form').style.display = type === 'login' ? 'block' : 'none';
    document.getElementById('register-form').style.display = type === 'register' ? 'block' : 'none';

    // Clear error
    this.hideElement('auth-error');
  }

  showAuthError(message) {
    const errorElement = document.getElementById('auth-error');
    errorElement.textContent = message;
    this.showElement('auth-error');
  }

  showError(message) {
    // Simple error display - could be enhanced with toast notifications
    console.error('Safari Popup Error:', message);
  }

  showSuccess(message) {
    // Simple success display - could be enhanced with toast notifications
    console.log('Safari Popup Success:', message);
  }

  showImportModal() {
    this.showElement('import-modal');
    this.showElement('overlay');
  }

  hideImportModal() {
    this.hideElement('import-modal');
    this.hideElement('overlay');
  }

  getDefaultFavicon(url) {
    try {
      const domain = new URL(url).hostname;
      return `https://www.google.com/s2/favicons?domain=${domain}&sz=16`;
    } catch {
      return 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><rect width="16" height="16" fill="%23ddd"/></svg>';
    }
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  loadSettingsForm() {
    document.getElementById('theme-select').value = this.preferences.theme || 'auto';
    document.getElementById('grid-size-select').value = this.preferences.gridSize || 'medium';
    document.getElementById('auto-sync-checkbox').checked = this.preferences.autoSync !== false;
    document.getElementById('notifications-checkbox').checked = this.preferences.notifications !== false;
  }

  setTheme(theme) {
    this.preferences.theme = theme;
    this.applyTheme();
    this.savePreferences();
  }

  setGridSize(size) {
    this.preferences.gridSize = size;
    this.savePreferences();
  }

  setAutoSync(enabled) {
    this.preferences.autoSync = enabled;
    this.savePreferences();
  }

  setNotifications(enabled) {
    this.preferences.notifications = enabled;
    this.savePreferences();
  }

  async clearCache() {
    try {
      // Implementation would clear local cache
      this.showSuccess('Cache cleared successfully');
    } catch (error) {
      this.showError('Failed to clear cache');
    }
  }

  async exportData() {
    try {
      // Implementation would export bookmark data
      this.showSuccess('Data exported successfully');
    } catch (error) {
      this.showError('Failed to export data');
    }
  }

  editBookmark(bookmark) {
    // Implementation for editing bookmarks
    console.log('Edit bookmark:', bookmark);
  }

  async deleteBookmark(bookmark) {
    if (!confirm(`Delete bookmark "${bookmark.title}"?`)) {
      return;
    }

    try {
      const result = await this.sendMessage({
        type: 'DELETE_BOOKMARK',
        id: bookmark.id
      });

      if (result.success) {
        await this.loadBookmarks();
        this.showSuccess('Bookmark deleted successfully');
      } else {
        this.showError(result.error || 'Failed to delete bookmark');
      }

    } catch (error) {
      console.error('Safari Popup: Failed to delete bookmark:', error);
      this.showError('Failed to delete bookmark');
    }
  }

  cancelImport() {
    this.isImporting = false;
    this.hideImportModal();
  }

  handleSyncNotification(notificationData) {
    try {
      const notification = JSON.parse(notificationData);

      if (notification.type === 'SYNC_UPDATE') {
        // Refresh bookmarks when sync update received
        this.loadBookmarks();
      }

    } catch (error) {
      console.warn('Safari Popup: Failed to handle sync notification:', error);
    }
  }
}

// Initialize popup when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  const popup = new SafariPopup();
  popup.init();
});

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
  module.exports = SafariPopup;
}