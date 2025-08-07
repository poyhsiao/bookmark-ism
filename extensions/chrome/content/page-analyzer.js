// Content script for page analysis
(function() {
  'use strict';

  /**
   * Extract page metadata for bookmarking
   */
  function extractPageMetadata() {
    const metadata = {
      url: window.location.href,
      title: document.title,
      description: '',
      keywords: [],
      favicon: '',
      language: document.documentElement.lang || 'en',
      timestamp: new Date().toISOString()
    };

    // Extract description from meta tags
    const descriptionMeta = document.querySelector('meta[name="description"]') ||
                           document.querySelector('meta[property="og:description"]') ||
                           document.querySelector('meta[name="twitter:description"]');

    if (descriptionMeta) {
      metadata.description = descriptionMeta.content.trim();
    }

    // Extract keywords from meta tags
    const keywordsMeta = document.querySelector('meta[name="keywords"]');
    if (keywordsMeta) {
      metadata.keywords = keywordsMeta.content
        .split(',')
        .map(keyword => keyword.trim())
        .filter(keyword => keyword.length > 0);
    }

    // Extract favicon
    const faviconLink = document.querySelector('link[rel="icon"]') ||
                       document.querySelector('link[rel="shortcut icon"]') ||
                       document.querySelector('link[rel="apple-touch-icon"]');

    if (faviconLink) {
      metadata.favicon = new URL(faviconLink.href, window.location.origin).href;
    } else {
      // Default favicon location
      metadata.favicon = new URL('/favicon.ico', window.location.origin).href;
    }

    // Extract additional Open Graph data
    const ogTitle = document.querySelector('meta[property="og:title"]');
    if (ogTitle && ogTitle.content.trim()) {
      metadata.ogTitle = ogTitle.content.trim();
    }

    const ogImage = document.querySelector('meta[property="og:image"]');
    if (ogImage && ogImage.content.trim()) {
      metadata.ogImage = new URL(ogImage.content, window.location.origin).href;
    }

    const ogType = document.querySelector('meta[property="og:type"]');
    if (ogType && ogType.content.trim()) {
      metadata.ogType = ogType.content.trim();
    }

    // Extract article metadata if available
    const articleAuthor = document.querySelector('meta[name="author"]') ||
                         document.querySelector('meta[property="article:author"]');
    if (articleAuthor) {
      metadata.author = articleAuthor.content.trim();
    }

    const articlePublished = document.querySelector('meta[property="article:published_time"]');
    if (articlePublished) {
      metadata.publishedTime = articlePublished.content.trim();
    }

    // Extract canonical URL
    const canonicalLink = document.querySelector('link[rel="canonical"]');
    if (canonicalLink) {
      metadata.canonicalUrl = canonicalLink.href;
    }

    return metadata;
  }

  /**
   * Detect if page is bookmarkable
   */
  function isBookmarkablePage() {
    const url = window.location.href;

    // Skip non-http(s) URLs
    if (!url.startsWith('http://') && !url.startsWith('https://')) {
      return false;
    }

    // Skip browser internal pages
    if (url.startsWith('chrome://') ||
        url.startsWith('chrome-extension://') ||
        url.startsWith('moz-extension://') ||
        url.startsWith('safari-extension://')) {
      return false;
    }

    // Skip empty or loading pages
    if (!document.title || document.title === 'Loading...') {
      return false;
    }

    return true;
  }

  /**
   * Send page metadata to background script
   */
  function sendPageMetadata() {
    if (!isBookmarkablePage()) {
      return;
    }

    const metadata = extractPageMetadata();

    chrome.runtime.sendMessage({
      type: 'PAGE_METADATA_EXTRACTED',
      metadata: metadata
    }).catch(error => {
      console.debug('Failed to send page metadata:', error);
    });
  }

  /**
   * Listen for messages from extension
   */
  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    switch (message.type) {
      case 'GET_PAGE_METADATA':
        if (isBookmarkablePage()) {
          sendResponse({
            success: true,
            metadata: extractPageMetadata()
          });
        } else {
          sendResponse({
            success: false,
            error: 'Page is not bookmarkable'
          });
        }
        break;

      case 'CHECK_BOOKMARKABLE':
        sendResponse({
          bookmarkable: isBookmarkablePage()
        });
        break;

      default:
        sendResponse({ success: false, error: 'Unknown message type' });
    }
  });

  /**
   * Detect page changes for single-page applications
   */
  let lastUrl = window.location.href;

  function detectUrlChange() {
    if (window.location.href !== lastUrl) {
      lastUrl = window.location.href;

      // Wait a bit for the page to update
      setTimeout(() => {
        sendPageMetadata();
      }, 1000);
    }
  }

  // Monitor for URL changes (for SPAs)
  const observer = new MutationObserver(detectUrlChange);
  observer.observe(document.body, {
    childList: true,
    subtree: true
  });

  // Also listen for popstate events
  window.addEventListener('popstate', detectUrlChange);

  // Send initial metadata when page loads
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', sendPageMetadata);
  } else {
    sendPageMetadata();
  }

  /**
   * Add visual indicator for bookmarked pages (optional feature)
   */
  function addBookmarkIndicator() {
    // This could be implemented to show a visual indicator
    // when the current page is already bookmarked
    // For now, we'll skip this to keep the MVP simple
  }

  /**
   * Handle keyboard shortcuts (optional feature)
   */
  document.addEventListener('keydown', (event) => {
    // Ctrl+D or Cmd+D to bookmark (override browser default)
    if ((event.ctrlKey || event.metaKey) && event.key === 'd') {
      if (isBookmarkablePage()) {
        event.preventDefault();

        chrome.runtime.sendMessage({
          type: 'BOOKMARK_CURRENT_PAGE_SHORTCUT'
        }).catch(error => {
          console.debug('Failed to send bookmark shortcut:', error);
        });
      }
    }
  });

  console.debug('Bookmark Sync content script loaded');
})();