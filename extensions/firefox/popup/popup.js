// Popup script for Chrome extension
import { formatTimestamp, extractPageMetadata, isValidUrl } from '../../shared/utils.js';
import { UI_CONFIG } from '../../shared/constants.js';

class PopupApp {
  constructor() {
    this.currentView = UI_CONFIG.VIEW_MODES.GRID;
    this.currentSort = 'created_at:desc';
    this.bookmarks = [];
    this.filteredBookmarks = [];
    this.currentUser = null;
    this.isAuthenticated = false;
    this.editingBookmark = null;

    this.init();
  }

  async init() {
    this.bindEvents();
    await this.checkAuthState();
    this.setupMessageListener();
  }

  bindEvents() {
    // Auth tabs
    document.getElementById('login-tab').addEventListener('click', () => this.showLoginForm());
    document.getElementById('register-tab').addEventListener('click', () => this.showRegisterForm());

    // Auth forms
    document.getElementById('login-form').addEventListener('submit', (e) => this.handleLogin(e));
    document.getElementById('register-form').addEventListener('submit', (e) => this.handleRegister(e));

    // Main actions
    document.getElementById('logout-btn').addEventListener('click', () => this.handleLogout());
    document.getElementById('bookmark-current-btn').addEventListener('click', () => this.bookmarkCurrentPage());
    document.getElementById('refresh-btn').addEventListener('click', () => this.refreshBookmarks());

    // Search
    document.getElementById('search-input').addEventListener('input', (e) => this.handleSearch(e.target.value));

    // View controls
    document.getElementById('grid-view-btn').addEventListener('click', () => this.setView(UI_CONFIG.VIEW_MODES.GRID));
    document.getElementById('list-view-btn').addEventListener('click', () => this.setView(UI_CONFIG.VIEW_MODES.LIST));
    document.getElementById('sort-select').addEventListener('change', (e) => this.setSort(e.target.value));

    // Modal
    document.getElementById('modal-close').addEventListener('click', () => this.closeModal());
    document.getElementById('modal-cancel').addEventListener('click', () => this.closeModal());
    document.getElementById('bookmark-form').addEventListener('submit', (e) => this.handleBookmarkSave(e));
    document.getElementById('modal-delete').addEventListener('click', () => this.handleBookmarkDelete());

    // Click outside modal to close
    document.getElementById('bookmark-modal').addEventListener('click', (e) => {
      if (e.target.id === 'bookmark-modal') {
        this.closeModal();
      }
    });
  }

  setupMessageListener() {
    const api = typeof browser !== 'undefined' ? browser : chrome;
    api.runtime.onMessage.addListener((message) => {
      switch (message.type) {
        case 'AUTH_STATE_CHANGED':
          this.handleAuthStateChange(message.authenticated, message.user);
          break;
        case 'SYNC_COMPLETED':
          this.handleSyncCompleted(message.eventsCount);
          break;
        case 'SYNC_EVENT_RECEIVED':
          this.handleSyncEvent(message.event);
          break;
      }
    });
  }

  async checkAuthState() {
    try {
      const response = await this.sendMessage({ type: 'GET_AUTH_STATE' });
      this.handleAuthStateChange(response.authenticated, response.user);
    } catch (error) {
      console.error('Failed to check auth state:', error);
      this.showScreen('auth-screen');
    }
  }

  handleAuthStateChange(authenticated, user) {
    this.isAuthenticated = authenticated;
    this.currentUser = user;

    if (authenticated) {
      document.getElementById('user-name').textContent = user.name || user.email;
      this.showScreen('main-screen');
      this.loadBookmarks();
      this.updateSyncStatus();
    } else {
      this.showScreen('auth-screen');
    }
  }

  showScreen(screenId) {
    document.querySelectorAll('.screen').forEach(screen => {
      screen.classList.add('hidden');
    });
    document.getElementById('loading').classList.add('hidden');
    document.getElementById(screenId).classList.remove('hidden');
  }

  showLoginForm() {
    document.getElementById('login-tab').classList.add('active');
    document.getElementById('register-tab').classList.remove('active');
    document.getElementById('login-form').classList.remove('hidden');
    document.getElementById('register-form').classList.add('hidden');
  }

  showRegisterForm() {
    document.getElementById('register-tab').classList.add('active');
    document.getElementById('login-tab').classList.remove('active');
    document.getElementById('register-form').classList.remove('hidden');
    document.getElementById('login-form').classList.add('hidden');
  }

  async handleLogin(e) {
    e.preventDefault();

    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error');

    try {
      const response = await this.sendMessage({
        type: 'LOGIN',
        email,
        password
      });

      if (response.success) {
        errorEl.classList.add('hidden');
      } else {
        errorEl.textContent = response.error;
        errorEl.classList.remove('hidden');
      }
    } catch (error) {
      errorEl.textContent = 'Login failed. Please try again.';
      errorEl.classList.remove('hidden');
    }
  }

  async handleRegister(e) {
    e.preventDefault();

    const name = document.getElementById('register-name').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    const errorEl = document.getElementById('register-error');

    try {
      const response = await this.sendMessage({
        type: 'REGISTER',
        name,
        email,
        password
      });

      if (response.success) {
        errorEl.classList.add('hidden');
      } else {
        errorEl.textContent = response.error;
        errorEl.classList.remove('hidden');
      }
    } catch (error) {
      errorEl.textContent = 'Registration failed. Please try again.';
      errorEl.classList.remove('hidden');
    }
  }

  async handleLogout() {
    try {
      await this.sendMessage({ type: 'LOGOUT' });
    } catch (error) {
      console.error('Logout failed:', error);
    }
  }

  async loadBookmarks() {
    try {
      const response = await this.sendMessage({ type: 'GET_BOOKMARKS' });

      if (response.bookmarks) {
        this.bookmarks = response.bookmarks;
        this.filteredBookmarks = [...this.bookmarks];
        this.renderBookmarks();
      }
    } catch (error) {
      console.error('Failed to load bookmarks:', error);
    }
  }

  async refreshBookmarks() {
    const refreshBtn = document.getElementById('refresh-btn');
    refreshBtn.disabled = true;

    try {
      // Force sync first
      await this.sendMessage({ type: 'FORCE_SYNC' });

      // Then reload bookmarks
      await this.loadBookmarks();
    } catch (error) {
      console.error('Failed to refresh bookmarks:', error);
    } finally {
      refreshBtn.disabled = false;
    }
  }

  async bookmarkCurrentPage() {
    try {
      const metadata = await extractPageMetadata();
      if (!metadata) {
        alert('Cannot bookmark this page');
        return;
      }

      const response = await this.sendMessage({
        type: 'CREATE_BOOKMARK',
        data: {
          url: metadata.url,
          title: metadata.title,
          description: '',
          tags: [],
          favicon: metadata.favicon
        }
      });

      if (response.success) {
        this.bookmarks.unshift(response.data);
        this.applyFiltersAndSort();
        this.renderBookmarks();

        // Show success feedback
        this.showNotification('Bookmark saved!', 'success');
      } else {
        this.showNotification('Failed to save bookmark', 'error');
      }
    } catch (error) {
      console.error('Failed to bookmark current page:', error);
      this.showNotification('Failed to save bookmark', 'error');
    }
  }

  handleSearch(query) {
    if (!query.trim()) {
      this.filteredBookmarks = [...this.bookmarks];
    } else {
      const searchTerm = query.toLowerCase();
      this.filteredBookmarks = this.bookmarks.filter(bookmark =>
        bookmark.title.toLowerCase().includes(searchTerm) ||
        bookmark.url.toLowerCase().includes(searchTerm) ||
        (bookmark.description && bookmark.description.toLowerCase().includes(searchTerm)) ||
        (bookmark.tags && bookmark.tags.some(tag => tag.toLowerCase().includes(searchTerm)))
      );
    }

    this.renderBookmarks();
  }

  setView(viewMode) {
    this.currentView = viewMode;

    // Update button states
    document.getElementById('grid-view-btn').classList.toggle('active', viewMode === UI_CONFIG.VIEW_MODES.GRID);
    document.getElementById('list-view-btn').classList.toggle('active', viewMode === UI_CONFIG.VIEW_MODES.LIST);

    // Update container visibility
    document.getElementById('bookmarks-grid').classList.toggle('hidden', viewMode !== UI_CONFIG.VIEW_MODES.GRID);
    document.getElementById('bookmarks-list').classList.toggle('hidden', viewMode !== UI_CONFIG.VIEW_MODES.LIST);

    this.renderBookmarks();
  }

  setSort(sortValue) {
    this.currentSort = sortValue;
    this.applyFiltersAndSort();
    this.renderBookmarks();
  }

  applyFiltersAndSort() {
    // Apply current search filter
    const searchQuery = document.getElementById('search-input').value;
    this.handleSearch(searchQuery);

    // Apply sort
    const [field, order] = this.currentSort.split(':');
    this.filteredBookmarks.sort((a, b) => {
      let aVal = a[field];
      let bVal = b[field];

      if (field === 'title' || field === 'url') {
        aVal = aVal.toLowerCase();
        bVal = bVal.toLowerCase();
      } else if (field === 'created_at' || field === 'updated_at') {
        aVal = new Date(aVal);
        bVal = new Date(bVal);
      }

      if (order === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });
  }

  renderBookmarks() {
    const gridContainer = document.getElementById('bookmarks-grid');
    const listContainer = document.getElementById('bookmarks-list');
    const emptyState = document.getElementById('empty-state');

    if (this.filteredBookmarks.length === 0) {
      gridContainer.innerHTML = '';
      listContainer.innerHTML = '';
      emptyState.classList.remove('hidden');
      return;
    }

    emptyState.classList.add('hidden');

    const container = this.currentView === UI_CONFIG.VIEW_MODES.GRID ? gridContainer : listContainer;
    const otherContainer = this.currentView === UI_CONFIG.VIEW_MODES.GRID ? listContainer : gridContainer;

    otherContainer.innerHTML = '';
    container.innerHTML = this.filteredBookmarks.map(bookmark => this.renderBookmarkItem(bookmark)).join('');

    // Add click listeners
    container.querySelectorAll('.bookmark-item').forEach((item, index) => {
      const bookmark = this.filteredBookmarks[index];

      item.addEventListener('click', (e) => {
        if (e.ctrlKey || e.metaKey) {
          // Open in new tab
          const api = typeof browser !== 'undefined' ? browser : chrome;
          api.tabs.create({ url: bookmark.url });
        } else {
          // Edit bookmark
          this.editBookmark(bookmark);
        }
      });

      item.addEventListener('contextmenu', (e) => {
        e.preventDefault();
        this.editBookmark(bookmark);
      });
    });
  }

  renderBookmarkItem(bookmark) {
    const viewClass = this.currentView === UI_CONFIG.VIEW_MODES.GRID ? 'grid' : 'list';
    const favicon = bookmark.favicon || 'data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><rect width="16" height="16" fill="%23ddd"/></svg>';

    const tags = bookmark.tags && bookmark.tags.length > 0 ?
      `<div class="bookmark-tags">${bookmark.tags.map(tag => `<span class="bookmark-tag">${tag}</span>`).join('')}</div>` : '';

    return `
      <div class="bookmark-item ${viewClass}">
        <img src="${favicon}" alt="" class="bookmark-favicon" onerror="this.style.display='none'">
        <div class="bookmark-title">${bookmark.title}</div>
        <div class="bookmark-url">${bookmark.url}</div>
        ${tags}
        <div class="bookmark-meta">${formatTimestamp(bookmark.created_at)}</div>
      </div>
    `;
  }

  editBookmark(bookmark) {
    this.editingBookmark = bookmark;

    document.getElementById('modal-title').textContent = 'Edit Bookmark';
    document.getElementById('bookmark-title').value = bookmark.title;
    document.getElementById('bookmark-url').value = bookmark.url;
    document.getElementById('bookmark-description').value = bookmark.description || '';
    document.getElementById('bookmark-tags').value = bookmark.tags ? bookmark.tags.join(', ') : '';
    document.getElementById('modal-delete').classList.remove('hidden');

    document.getElementById('bookmark-modal').classList.remove('hidden');
  }

  async handleBookmarkSave(e) {
    e.preventDefault();

    const title = document.getElementById('bookmark-title').value;
    const url = document.getElementById('bookmark-url').value;
    const description = document.getElementById('bookmark-description').value;
    const tagsInput = document.getElementById('bookmark-tags').value;
    const tags = tagsInput ? tagsInput.split(',').map(tag => tag.trim()).filter(tag => tag) : [];

    if (!isValidUrl(url)) {
      alert('Please enter a valid URL');
      return;
    }

    try {
      const bookmarkData = { title, url, description, tags };

      if (this.editingBookmark) {
        // Update existing bookmark
        const response = await this.sendMessage({
          type: 'UPDATE_BOOKMARK',
          id: this.editingBookmark.id,
          data: bookmarkData
        });

        if (response.success) {
          // Update local bookmark
          const index = this.bookmarks.findIndex(b => b.id === this.editingBookmark.id);
          if (index !== -1) {
            this.bookmarks[index] = { ...this.bookmarks[index], ...bookmarkData };
            this.applyFiltersAndSort();
            this.renderBookmarks();
          }
          this.closeModal();
          this.showNotification('Bookmark updated!', 'success');
        } else {
          this.showNotification('Failed to update bookmark', 'error');
        }
      } else {
        // Create new bookmark
        const response = await this.sendMessage({
          type: 'CREATE_BOOKMARK',
          data: bookmarkData
        });

        if (response.success) {
          this.bookmarks.unshift(response.data);
          this.applyFiltersAndSort();
          this.renderBookmarks();
          this.closeModal();
          this.showNotification('Bookmark created!', 'success');
        } else {
          this.showNotification('Failed to create bookmark', 'error');
        }
      }
    } catch (error) {
      console.error('Failed to save bookmark:', error);
      this.showNotification('Failed to save bookmark', 'error');
    }
  }

  async handleBookmarkDelete() {
    if (!this.editingBookmark) return;

    if (!confirm('Are you sure you want to delete this bookmark?')) {
      return;
    }

    try {
      const response = await this.sendMessage({
        type: 'DELETE_BOOKMARK',
        id: this.editingBookmark.id
      });

      if (response.success) {
        // Remove from local bookmarks
        this.bookmarks = this.bookmarks.filter(b => b.id !== this.editingBookmark.id);
        this.applyFiltersAndSort();
        this.renderBookmarks();
        this.closeModal();
        this.showNotification('Bookmark deleted!', 'success');
      } else {
        this.showNotification('Failed to delete bookmark', 'error');
      }
    } catch (error) {
      console.error('Failed to delete bookmark:', error);
      this.showNotification('Failed to delete bookmark', 'error');
    }
  }

  closeModal() {
    document.getElementById('bookmark-modal').classList.add('hidden');
    this.editingBookmark = null;
    document.getElementById('bookmark-form').reset();
    document.getElementById('modal-delete').classList.add('hidden');
  }

  async updateSyncStatus() {
    try {
      const status = await this.sendMessage({ type: 'GET_SYNC_STATUS' });
      const indicator = document.getElementById('sync-indicator');
      const text = document.getElementById('sync-text');

      if (status.connected) {
        indicator.className = 'sync-indicator online';
        text.textContent = 'Online';
      } else {
        indicator.className = 'sync-indicator offline';
        text.textContent = 'Offline';
      }
    } catch (error) {
      console.error('Failed to get sync status:', error);
    }
  }

  handleSyncCompleted(eventsCount) {
    if (eventsCount > 0) {
      this.loadBookmarks();
      this.showNotification(`Synced ${eventsCount} changes`, 'info');
    }
    this.updateSyncStatus();
  }

  handleSyncEvent(event) {
    // Refresh bookmarks when sync events are received
    this.loadBookmarks();
  }

  showNotification(message, type = 'info') {
    // Simple notification - could be enhanced with a proper notification system
    console.log(`${type.toUpperCase()}: ${message}`);

    // For now, just update the sync status text briefly
    const text = document.getElementById('sync-text');
    const originalText = text.textContent;
    text.textContent = message;

    setTimeout(() => {
      text.textContent = originalText;
    }, 2000);
  }

  sendMessage(message) {
    return new Promise((resolve, reject) => {
      // Use browser API for Firefox
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

// Initialize popup when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  new PopupApp();
});