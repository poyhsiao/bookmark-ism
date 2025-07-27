// Authentication manager for Firefox extension
// Adapted from Chrome version for Firefox compatibility

class AuthManager {
  constructor() {
    this.isAuthenticated = false;
    this.user = null;
  }

  /**
   * Initialize authentication state
   */
  async init() {
    await apiClient.init();

    const stored = await browser.storage.local.get([
      'auth_access_token',
      'user_data'
    ]);

    if (stored.auth_access_token && stored.user_data) {
      this.isAuthenticated = true;
      this.user = stored.user_data;
      apiClient.setToken(stored.auth_access_token);
    }

    return this.isAuthenticated;
  }

  /**
   * Login with email and password
   */
  async login(email, password) {
    try {
      const response = await apiClient.login(email, password);

      if (response.success && response.user) {
        this.isAuthenticated = true;
        this.user = response.user;

        // Notify other parts of the extension
        browser.runtime.sendMessage({
          type: 'AUTH_STATE_CHANGED',
          authenticated: true,
          user: this.user
        }).catch(() => {}); // Ignore if no listeners

        return { success: true, user: this.user };
      }

      return { success: false, error: response.error || 'Login failed' };
    } catch (error) {
      console.error('Login error:', error);
      return { success: false, error: error.message };
    }
  }

  /**
   * Register new user
   */
  async register(email, password, name) {
    try {
      const response = await apiClient.register(email, password, name);

      if (response.success && response.user) {
        this.isAuthenticated = true;
        this.user = response.user;

        // Notify other parts of the extension
        browser.runtime.sendMessage({
          type: 'AUTH_STATE_CHANGED',
          authenticated: true,
          user: this.user
        }).catch(() => {}); // Ignore if no listeners

        return { success: true, user: this.user };
      }

      return { success: false, error: response.error || 'Registration failed' };
    } catch (error) {
      console.error('Registration error:', error);
      return { success: false, error: error.message };
    }
  }

  /**
   * Logout user
   */
  async logout() {
    try {
      await apiClient.logout();

      this.isAuthenticated = false;
      this.user = null;

      // Notify other parts of the extension
      browser.runtime.sendMessage({
        type: 'AUTH_STATE_CHANGED',
        authenticated: false,
        user: null
      }).catch(() => {}); // Ignore if no listeners

      return { success: true };
    } catch (error) {
      console.error('Logout error:', error);
      return { success: false, error: error.message };
    }
  }

  /**
   * Check if user is authenticated
   */
  isUserAuthenticated() {
    return this.isAuthenticated;
  }

  /**
   * Get current user data
   */
  getCurrentUser() {
    return this.user;
  }

  /**
   * Validate current token
   */
  async validateToken() {
    if (!this.isAuthenticated) {
      return false;
    }

    try {
      // Try to fetch user bookmarks to validate token
      await apiClient.getBookmarks({ limit: 1 });
      return true;
    } catch (error) {
      if (error.code === 'AUTH_REQUIRED' || error.code === 'AUTH_INVALID') {
        // Token is invalid, logout user
        await this.logout();
        return false;
      }
      // Other errors don't necessarily mean invalid token
      return true;
    }
  }
}

// Create global instance for Firefox
const authManager = new AuthManager();