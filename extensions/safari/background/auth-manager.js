// Safari Extension Auth Manager
// Handles authentication for Safari Web Extension

import { apiClient } from '../../shared/api-client.js';
import { STORAGE_KEYS } from '../../shared/constants.js';

class SafariAuthManager {
  constructor() {
    this.currentUser = null;
    this.isAuthenticated = false;
  }

  /**
   * Initialize auth manager
   */
  async init() {
    try {
      // Load stored authentication data
      const stored = await browser.storage.local.get([
        STORAGE_KEYS.AUTH_TOKEN,
        STORAGE_KEYS.USER_DATA
      ]);

      if (stored[STORAGE_KEYS.AUTH_TOKEN]) {
        apiClient.setToken(stored[STORAGE_KEYS.AUTH_TOKEN]);
        this.currentUser = stored[STORAGE_KEYS.USER_DATA];
        this.isAuthenticated = true;

        console.log('Safari Auth Manager: User authenticated from storage');
      }
    } catch (error) {
      console.error('Safari Auth Manager: Failed to initialize:', error);
    }
  }

  /**
   * Check if user is authenticated
   */
  isUserAuthenticated() {
    return this.isAuthenticated && this.currentUser;
  }

  /**
   * Get current user data
   */
  getCurrentUser() {
    return this.currentUser;
  }

  /**
   * Login user
   */
  async login(email, password) {
    try {
      const response = await apiClient.login(email, password);

      if (response.success || response.token) {
        this.currentUser = response.user;
        this.isAuthenticated = true;

        console.log('Safari Auth Manager: User logged in successfully');

        return {
          success: true,
          user: response.user,
          message: 'Login successful'
        };
      } else {
        return {
          success: false,
          error: response.error || 'Login failed'
        };
      }
    } catch (error) {
      console.error('Safari Auth Manager: Login failed:', error);
      return {
        success: false,
        error: error.message || 'Login failed'
      };
    }
  }

  /**
   * Register new user
   */
  async register(email, password, name) {
    try {
      const response = await apiClient.register(email, password, name);

      if (response.success || response.token) {
        this.currentUser = response.user;
        this.isAuthenticated = true;

        console.log('Safari Auth Manager: User registered successfully');

        return {
          success: true,
          user: response.user,
          message: 'Registration successful'
        };
      } else {
        return {
          success: false,
          error: response.error || 'Registration failed'
        };
      }
    } catch (error) {
      console.error('Safari Auth Manager: Registration failed:', error);
      return {
        success: false,
        error: error.message || 'Registration failed'
      };
    }
  }

  /**
   * Logout user
   */
  async logout() {
    try {
      await apiClient.logout();

      this.currentUser = null;
      this.isAuthenticated = false;

      console.log('Safari Auth Manager: User logged out successfully');

      return {
        success: true,
        message: 'Logout successful'
      };
    } catch (error) {
      console.error('Safari Auth Manager: Logout failed:', error);

      // Clear local state even if API call fails
      this.currentUser = null;
      this.isAuthenticated = false;

      return {
        success: true,
        message: 'Logout completed (with errors)'
      };
    }
  }

  /**
   * Refresh authentication token
   */
  async refreshToken() {
    try {
      const stored = await browser.storage.local.get([STORAGE_KEYS.REFRESH_TOKEN]);
      const refreshToken = stored[STORAGE_KEYS.REFRESH_TOKEN];

      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      // Note: This would need to be implemented in the API
      const response = await apiClient.request('/api/v1/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken })
      });

      if (response.token) {
        apiClient.setToken(response.token);
        await browser.storage.local.set({
          [STORAGE_KEYS.AUTH_TOKEN]: response.token
        });

        console.log('Safari Auth Manager: Token refreshed successfully');
        return { success: true };
      }

      throw new Error('Token refresh failed');
    } catch (error) {
      console.error('Safari Auth Manager: Token refresh failed:', error);

      // If refresh fails, logout user
      await this.logout();

      return { success: false, error: error.message };
    }
  }

  /**
   * Handle authentication errors
   */
  handleAuthError(error) {
    if (error.code === 'AUTH_EXPIRED') {
      // Try to refresh token
      this.refreshToken();
    } else if (error.code === 'AUTH_INVALID') {
      // Force logout
      this.logout();
    }
  }
}

// Export singleton instance
export const authManager = new SafariAuthManager();