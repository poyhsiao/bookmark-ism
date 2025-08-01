// Safari Extension Error Handler
// Handles Safari-specific errors and provides graceful degradation

class SafariErrorHandler {
  constructor() {
    this.errorLog = [];
    this.maxLogSize = 100;
  }

  /**
   * Handle Safari-specific errors
   */
  async handleSafariError(error, context = '') {
    const errorInfo = {
      timestamp: new Date().toISOString(),
      context: context,
      message: error?.message || 'Unknown error',
      stack: error?.stack,
      type: this.categorizeError(error)
    };

    // Log error
    this.logError(errorInfo);

    // Handle specific error types
    switch (errorInfo.type) {
      case 'PERMISSION_DENIED':
        return this.handlePermissionError(error, context);

      case 'STORAGE_QUOTA_EXCEEDED':
        return this.handleStorageError(error, context);

      case 'NETWORK_ERROR':
        return this.handleNetworkError(error, context);

      case 'API_UNAVAILABLE':
        return this.handleApiUnavailableError(error, context);

      default:
        return this.handleGenericError(error, context);
    }
  }

  /**
   * Categorize error type
   */
  categorizeError(error) {
    if (!error) return 'UNKNOWN';

    const message = error.message?.toLowerCase() || '';

    if (message.includes('permission') || message.includes('denied')) {
      return 'PERMISSION_DENIED';
    }

    if (message.includes('quota') || message.includes('storage')) {
      return 'STORAGE_QUOTA_EXCEEDED';
    }

    if (message.includes('network') || message.includes('fetch')) {
      return 'NETWORK_ERROR';
    }

    if (message.includes('api') || message.includes('undefined')) {
      return 'API_UNAVAILABLE';
    }

    return 'GENERIC_ERROR';
  }

  /**
   * Handle permission errors
   */
  async handlePermissionError(error, context) {
    console.warn('Safari Error Handler: Permission error in', context, error);

    // Try to gracefully degrade functionality
    const fallbackResult = {
      success: false,
      error: 'Permission denied',
      fallback: true,
      message: 'Some features may be limited due to Safari permissions'
    };

    // Notify user about permission issue
    if (this.isNotificationAvailable()) {
      try {
        await browser.notifications.create({
          type: 'basic',
          iconUrl: 'icons/icon48.png',
          title: 'Permission Required',
          message: 'Please check Safari extension permissions in Safari preferences'
        });
      } catch (notificationError) {
        console.warn('Safari Error Handler: Could not show notification:', notificationError);
      }
    }

    return fallbackResult;
  }

  /**
   * Handle storage quota exceeded errors
   */
  async handleStorageError(error, context) {
    console.warn('Safari Error Handler: Storage error in', context, error);

    try {
      // Try to clean up storage
      const storageManager = (await import('./storage-manager.js')).storageManager;
      await storageManager.cleanup();

      return {
        success: false,
        error: 'Storage quota exceeded',
        recovered: true,
        message: 'Storage cleaned up, please try again'
      };
    } catch (cleanupError) {
      console.error('Safari Error Handler: Storage cleanup failed:', cleanupError);

      return {
        success: false,
        error: 'Storage quota exceeded',
        recovered: false,
        message: 'Please clear extension data in Safari preferences'
      };
    }
  }

  /**
   * Handle network errors
   */
  async handleNetworkError(error, context) {
    console.warn('Safari Error Handler: Network error in', context, error);

    // Check if we're offline
    const isOnline = navigator.onLine;

    if (!isOnline) {
      return {
        success: false,
        error: 'Network unavailable',
        offline: true,
        message: 'Working offline - changes will sync when connection is restored'
      };
    }

    // Try to determine if it's a server issue
    const isServerError = error.message?.includes('500') || error.message?.includes('502');

    return {
      success: false,
      error: 'Network error',
      serverError: isServerError,
      message: isServerError
        ? 'Server temporarily unavailable, please try again later'
        : 'Connection failed, please check your internet connection'
    };
  }

  /**
   * Handle API unavailable errors
   */
  async handleApiUnavailableError(error, context) {
    console.warn('Safari Error Handler: API unavailable in', context, error);

    // Check which APIs are available
    const availableApis = this.checkAvailableApis();

    return {
      success: false,
      error: 'API unavailable',
      availableApis: availableApis,
      message: 'Some Safari APIs are not available, functionality may be limited'
    };
  }

  /**
   * Handle generic errors
   */
  async handleGenericError(error, context) {
    console.error('Safari Error Handler: Generic error in', context, error);

    return {
      success: false,
      error: error?.message || 'Unknown error occurred',
      context: context,
      message: 'An unexpected error occurred, please try again'
    };
  }

  /**
   * Check which Safari APIs are available
   */
  checkAvailableApis() {
    const apis = {
      storage: !!browser?.storage?.local,
      bookmarks: !!browser?.bookmarks,
      tabs: !!browser?.tabs,
      contextMenus: !!browser?.contextMenus,
      notifications: !!browser?.notifications,
      runtime: !!browser?.runtime
    };

    console.log('Safari Error Handler: Available APIs:', apis);
    return apis;
  }

  /**
   * Check if notifications API is available
   */
  isNotificationAvailable() {
    return !!(browser?.notifications?.create);
  }

  /**
   * Log error to internal log
   */
  logError(errorInfo) {
    this.errorLog.unshift(errorInfo);

    // Keep log size manageable
    if (this.errorLog.length > this.maxLogSize) {
      this.errorLog = this.errorLog.slice(0, this.maxLogSize);
    }

    // Also log to console for debugging
    console.error('Safari Extension Error:', errorInfo);
  }

  /**
   * Get error log
   */
  getErrorLog() {
    return [...this.errorLog];
  }

  /**
   * Clear error log
   */
  clearErrorLog() {
    this.errorLog = [];
  }

  /**
   * Get error statistics
   */
  getErrorStats() {
    const stats = {
      total: this.errorLog.length,
      byType: {},
      byContext: {},
      recent: this.errorLog.slice(0, 10)
    };

    // Count by type
    this.errorLog.forEach(error => {
      stats.byType[error.type] = (stats.byType[error.type] || 0) + 1;
      stats.byContext[error.context] = (stats.byContext[error.context] || 0) + 1;
    });

    return stats;
  }

  /**
   * Setup global error handlers
   */
  setupGlobalErrorHandlers() {
    // Handle unhandled promise rejections
    if (typeof window !== 'undefined') {
      window.addEventListener('unhandledrejection', (event) => {
        this.handleSafariError(event.reason, 'unhandled_promise_rejection');
      });

      // Handle general errors
      window.addEventListener('error', (event) => {
        this.handleSafariError(event.error, 'global_error');
      });
    }

    console.log('Safari Error Handler: Global error handlers setup');
  }

  /**
   * Create user-friendly error message
   */
  createUserMessage(error, context) {
    const errorType = this.categorizeError(error);

    const messages = {
      PERMISSION_DENIED: 'Please check Safari extension permissions in Safari preferences.',
      STORAGE_QUOTA_EXCEEDED: 'Storage is full. Please clear some data or increase storage limit.',
      NETWORK_ERROR: 'Connection failed. Please check your internet connection.',
      API_UNAVAILABLE: 'Some features are not available in this version of Safari.',
      GENERIC_ERROR: 'An unexpected error occurred. Please try again.'
    };

    return messages[errorType] || messages.GENERIC_ERROR;
  }

  /**
   * Report error to server (if available)
   */
  async reportError(error, context) {
    try {
      // Only report in development or if user opted in
      const shouldReport = false; // Set based on user preferences

      if (!shouldReport) return;

      const errorReport = {
        timestamp: new Date().toISOString(),
        context: context,
        message: error?.message,
        stack: error?.stack,
        userAgent: navigator.userAgent,
        url: window.location?.href,
        extensionVersion: browser.runtime.getManifest()?.version
      };

      // Send to error reporting service
      await fetch('/api/v1/errors/report', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(errorReport)
      });

    } catch (reportError) {
      console.warn('Safari Error Handler: Failed to report error:', reportError);
    }
  }
}

// Export singleton instance
export const errorHandler = new SafariErrorHandler();