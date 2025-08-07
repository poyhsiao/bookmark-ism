// Shared constants for browser extensions
export const API_CONFIG = {
  BASE_URL: 'http://localhost:8080',
  ENDPOINTS: {
    AUTH: '/api/v1/auth',
    BOOKMARKS: '/api/v1/bookmarks',
    COLLECTIONS: '/api/v1/collections',
    SYNC: '/api/v1/sync',
    WEBSOCKET: '/api/v1/sync/ws'
  }
};

export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_access_token',
  REFRESH_TOKEN: 'auth_refresh_token',
  USER_DATA: 'user_data',
  BOOKMARKS_CACHE: 'bookmarks_cache',
  SYNC_STATE: 'sync_state',
  DEVICE_ID: 'device_id'
};

export const SYNC_EVENTS = {
  BOOKMARK_CREATED: 'bookmark_created',
  BOOKMARK_UPDATED: 'bookmark_updated',
  BOOKMARK_DELETED: 'bookmark_deleted',
  COLLECTION_CREATED: 'collection_created',
  COLLECTION_UPDATED: 'collection_updated',
  COLLECTION_DELETED: 'collection_deleted'
};

export const MESSAGE_TYPES = {
  PING: 'ping',
  PONG: 'pong',
  SYNC_REQUEST: 'sync_request',
  SYNC_RESPONSE: 'sync_response',
  SYNC_EVENT: 'sync_event'
};

export const UI_CONFIG = {
  GRID_SIZES: {
    SMALL: 'small',
    MEDIUM: 'medium',
    LARGE: 'large'
  },
  VIEW_MODES: {
    GRID: 'grid',
    LIST: 'list'
  },
  THEMES: {
    LIGHT: 'light',
    DARK: 'dark',
    AUTO: 'auto'
  }
};

export const ERROR_CODES = {
  AUTH_REQUIRED: 'AUTH_REQUIRED',
  AUTH_INVALID: 'AUTH_INVALID',
  AUTH_EXPIRED: 'AUTH_EXPIRED',
  NETWORK_ERROR: 'NETWORK_ERROR',
  SYNC_CONFLICT: 'SYNC_CONFLICT',
  VALIDATION_FAILED: 'VALIDATION_FAILED'
};