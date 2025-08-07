// Shared utility functions for browser extensions
import { STORAGE_KEYS } from './constants.js';

/**
 * Generate a unique device ID for this browser instance
 */
export function generateDeviceId() {
  return 'ext_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
}

/**
 * Get or create device ID
 */
export async function getDeviceId() {
  const stored = await chrome.storage.local.get([STORAGE_KEYS.DEVICE_ID]);
  if (stored[STORAGE_KEYS.DEVICE_ID]) {
    return stored[STORAGE_KEYS.DEVICE_ID];
  }

  const deviceId = generateDeviceId();
  await chrome.storage.local.set({ [STORAGE_KEYS.DEVICE_ID]: deviceId });
  return deviceId;
}

/**
 * Validate URL format
 */
export function isValidUrl(string) {
  try {
    new URL(string);
    return true;
  } catch (_) {
    return false;
  }
}

/**
 * Extract page metadata from current tab
 */
export async function extractPageMetadata() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

  if (!tab || !tab.url || tab.url.startsWith('chrome://')) {
    return null;
  }

  return {
    url: tab.url,
    title: tab.title || '',
    favicon: tab.favIconUrl || '',
    timestamp: new Date().toISOString()
  };
}

/**
 * Debounce function to limit API calls
 */
export function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

/**
 * Format timestamp for display
 */
export function formatTimestamp(timestamp) {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
}

/**
 * Sanitize HTML content
 */
export function sanitizeHtml(html) {
  const div = document.createElement('div');
  div.textContent = html;
  return div.innerHTML;
}

/**
 * Generate random color for collections
 */
export function generateRandomColor() {
  const colors = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7',
    '#DDA0DD', '#98D8C8', '#F7DC6F', '#BB8FCE', '#85C1E9'
  ];
  return colors[Math.floor(Math.random() * colors.length)];
}

/**
 * Handle API errors consistently
 */
export function handleApiError(error) {
  console.error('API Error:', error);

  if (error.status === 401) {
    return { code: 'AUTH_REQUIRED', message: 'Authentication required' };
  } else if (error.status === 403) {
    return { code: 'AUTH_FORBIDDEN', message: 'Access forbidden' };
  } else if (error.status === 404) {
    return { code: 'NOT_FOUND', message: 'Resource not found' };
  } else if (error.status >= 500) {
    return { code: 'SERVER_ERROR', message: 'Server error occurred' };
  } else {
    return { code: 'UNKNOWN_ERROR', message: error.message || 'Unknown error' };
  }
}