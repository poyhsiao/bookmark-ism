// Safari Extension Content Script - Page Analyzer
// Analyzes web pages for bookmark metadata extraction

(function() {
  'use strict';

  /**
   * Page Analyzer for Safari Extension
   */
  class SafariPageAnalyzer {
    constructor() {
      this.metadata = null;
      this.isBookmarkable = true;
    }

    /**
     * Extract page metadata
     */
    extractPageMetadata() {
      try {
        // Check if page is bookmarkable
        if (!this.isBookmarkable()) {
          return null;
        }

        const metadata = {
          url: window.location.href,
          title: this.extractTitle(),
          description: this.extractDescription(),
          keywords: this.extractKeywords(),
          favicon: this.extractFavicon(),
          language: this.extractLanguage(),
          author: this.extractAuthor(),
          publishedDate: this.extractPublishedDate(),
          contentType: this.extractContentType(),
          readingTime: this.estimateReadingTime(),
          wordCount: this.getWordCount(),
          timestamp: new Date().toISOString()
        };

        this.metadata = metadata;
        return metadata;

      } catch (error) {
        console.error('Safari Page Analyzer: Failed to extract metadata:', error);
        return null;
      }
    }

    /**
     * Check if current page is bookmarkable
     */
    isBookmarkable() {
      const url = window.location.href;

      // Exclude Safari-specific URLs
      if (url.startsWith('safari://') ||
          url.startsWith('safari-extension://') ||
          url.startsWith('about:') ||
          url.startsWith('chrome://') ||
          url.startsWith('moz-extension://') ||
          url.startsWith('data:') ||
          url.startsWith('javascript:')) {
        return false;
      }

      // Only allow HTTP/HTTPS
      if (!url.startsWith('http://') && !url.startsWith('https://')) {
        return false;
      }

      return true;
    }

    /**
     * Extract page title
     */
    extractTitle() {
      // Try multiple sources for title
      const sources = [
        () => document.querySelector('meta[property="og:title"]')?.content,
        () => document.querySelector('meta[name="twitter:title"]')?.content,
        () => document.querySelector('title')?.textContent,
        () => document.querySelector('h1')?.textContent
      ];

      for (const source of sources) {
        try {
          const title = source();
          if (title && title.trim()) {
            return title.trim();
          }
        } catch (error) {
          continue;
        }
      }

      return document.title || 'Untitled';
    }

    /**
     * Extract page description
     */
    extractDescription() {
      // Try multiple sources for description
      const sources = [
        () => document.querySelector('meta[property="og:description"]')?.content,
        () => document.querySelector('meta[name="twitter:description"]')?.content,
        () => document.querySelector('meta[name="description"]')?.content,
        () => document.querySelector('meta[name="Description"]')?.content,
        () => this.extractFirstParagraph()
      ];

      for (const source of sources) {
        try {
          const description = source();
          if (description && description.trim()) {
            return description.trim().substring(0, 300);
          }
        } catch (error) {
          continue;
        }
      }

      return '';
    }

    /**
     * Extract first meaningful paragraph
     */
    extractFirstParagraph() {
      const paragraphs = document.querySelectorAll('p');

      for (const p of paragraphs) {
        const text = p.textContent?.trim();
        if (text && text.length > 50) {
          return text;
        }
      }

      return '';
    }

    /**
     * Extract keywords
     */
    extractKeywords() {
      const keywords = [];

      // From meta keywords
      const metaKeywords = document.querySelector('meta[name="keywords"]')?.content;
      if (metaKeywords) {
        keywords.push(...metaKeywords.split(',').map(k => k.trim()));
      }

      // From headings
      const headings = document.querySelectorAll('h1, h2, h3');
      headings.forEach(heading => {
        const text = heading.textContent?.trim();
        if (text && text.length < 100) {
          keywords.push(text);
        }
      });

      // Remove duplicates and limit
      return [...new Set(keywords)].slice(0, 10);
    }

    /**
     * Extract favicon URL
     */
    extractFavicon() {
      // Try multiple favicon sources
      const sources = [
        () => document.querySelector('link[rel="icon"]')?.href,
        () => document.querySelector('link[rel="shortcut icon"]')?.href,
        () => document.querySelector('link[rel="apple-touch-icon"]')?.href,
        () => document.querySelector('meta[property="og:image"]')?.content,
        () => this.getDefaultFavicon()
      ];

      for (const source of sources) {
        try {
          const favicon = source();
          if (favicon) {
            return this.resolveUrl(favicon);
          }
        } catch (error) {
          continue;
        }
      }

      return this.getDefaultFavicon();
    }

    /**
     * Get default favicon URL
     */
    getDefaultFavicon() {
      try {
        const domain = new URL(window.location.href).hostname;
        return `https://www.google.com/s2/favicons?domain=${domain}&sz=32`;
      } catch (error) {
        return null;
      }
    }

    /**
     * Extract page language
     */
    extractLanguage() {
      return document.documentElement.lang ||
             document.querySelector('meta[http-equiv="content-language"]')?.content ||
             'en';
    }

    /**
     * Extract author information
     */
    extractAuthor() {
      const sources = [
        () => document.querySelector('meta[name="author"]')?.content,
        () => document.querySelector('meta[property="article:author"]')?.content,
        () => document.querySelector('[rel="author"]')?.textContent,
        () => document.querySelector('.author')?.textContent,
        () => document.querySelector('.byline')?.textContent
      ];

      for (const source of sources) {
        try {
          const author = source();
          if (author && author.trim()) {
            return author.trim();
          }
        } catch (error) {
          continue;
        }
      }

      return null;
    }

    /**
     * Extract published date
     */
    extractPublishedDate() {
      const sources = [
        () => document.querySelector('meta[property="article:published_time"]')?.content,
        () => document.querySelector('meta[name="date"]')?.content,
        () => document.querySelector('time[datetime]')?.getAttribute('datetime'),
        () => document.querySelector('time')?.textContent,
        () => document.querySelector('.date')?.textContent,
        () => document.querySelector('.published')?.textContent
      ];

      for (const source of sources) {
        try {
          const date = source();
          if (date && date.trim()) {
            const parsedDate = new Date(date.trim());
            if (!isNaN(parsedDate.getTime())) {
              return parsedDate.toISOString();
            }
          }
        } catch (error) {
          continue;
        }
      }

      return null;
    }

    /**
     * Extract content type
     */
    extractContentType() {
      // Try to determine content type from various sources
      const ogType = document.querySelector('meta[property="og:type"]')?.content;
      if (ogType) {
        return ogType;
      }

      // Analyze URL and content
      const url = window.location.href.toLowerCase();
      const title = document.title.toLowerCase();

      if (url.includes('/blog/') || url.includes('/post/') || title.includes('blog')) {
        return 'article';
      }

      if (url.includes('/video/') || document.querySelector('video')) {
        return 'video';
      }

      if (url.includes('/product/') || url.includes('/shop/')) {
        return 'product';
      }

      if (document.querySelector('article')) {
        return 'article';
      }

      return 'website';
    }

    /**
     * Estimate reading time
     */
    estimateReadingTime() {
      try {
        const text = this.getPageText();
        const words = text.split(/\s+/).length;
        const wordsPerMinute = 200; // Average reading speed
        const minutes = Math.ceil(words / wordsPerMinute);
        return minutes;
      } catch (error) {
        return null;
      }
    }

    /**
     * Get word count
     */
    getWordCount() {
      try {
        const text = this.getPageText();
        return text.split(/\s+/).length;
      } catch (error) {
        return 0;
      }
    }

    /**
     * Get main text content from page
     */
    getPageText() {
      // Try to get main content
      const contentSelectors = [
        'main',
        'article',
        '.content',
        '.post-content',
        '.entry-content',
        '#content',
        '.main-content'
      ];

      for (const selector of contentSelectors) {
        const element = document.querySelector(selector);
        if (element) {
          return element.textContent || '';
        }
      }

      // Fallback to body text
      return document.body.textContent || '';
    }

    /**
     * Resolve relative URLs to absolute
     */
    resolveUrl(url) {
      try {
        return new URL(url, window.location.href).href;
      } catch (error) {
        return url;
      }
    }

    /**
     * Check if page has changed since last analysis
     */
    hasPageChanged() {
      const currentUrl = window.location.href;
      const currentTitle = document.title;

      if (!this.metadata) {
        return true;
      }

      return this.metadata.url !== currentUrl ||
             this.metadata.title !== currentTitle;
    }

    /**
     * Get structured data from page
     */
    extractStructuredData() {
      const structuredData = [];

      // JSON-LD
      const jsonLdScripts = document.querySelectorAll('script[type="application/ld+json"]');
      jsonLdScripts.forEach(script => {
        try {
          const data = JSON.parse(script.textContent);
          structuredData.push(data);
        } catch (error) {
          // Ignore invalid JSON-LD
        }
      });

      // Microdata
      const microdataElements = document.querySelectorAll('[itemscope]');
      microdataElements.forEach(element => {
        try {
          const data = this.extractMicrodata(element);
          if (data) {
            structuredData.push(data);
          }
        } catch (error) {
          // Ignore microdata extraction errors
        }
      });

      return structuredData;
    }

    /**
     * Extract microdata from element
     */
    extractMicrodata(element) {
      const data = {};
      const itemType = element.getAttribute('itemtype');

      if (itemType) {
        data['@type'] = itemType;
      }

      const properties = element.querySelectorAll('[itemprop]');
      properties.forEach(prop => {
        const name = prop.getAttribute('itemprop');
        const value = prop.getAttribute('content') ||
                     prop.textContent ||
                     prop.getAttribute('href') ||
                     prop.getAttribute('src');

        if (name && value) {
          data[name] = value;
        }
      });

      return Object.keys(data).length > 0 ? data : null;
    }

    /**
     * Detect if page is a single-page application
     */
    isSPA() {
      // Check for common SPA frameworks
      return !!(window.React ||
                window.Vue ||
                window.angular ||
                window.Angular ||
                document.querySelector('[ng-app]') ||
                document.querySelector('[data-reactroot]'));
    }

    /**
     * Monitor page changes for SPAs
     */
    startPageMonitoring() {
      if (!this.isSPA()) {
        return;
      }

      // Monitor URL changes
      let lastUrl = window.location.href;

      const checkForChanges = () => {
        if (window.location.href !== lastUrl) {
          lastUrl = window.location.href;

          // Wait for content to load
          setTimeout(() => {
            this.extractPageMetadata();
            this.notifyPageChange();
          }, 1000);
        }
      };

      // Check periodically
      setInterval(checkForChanges, 2000);

      // Listen for history changes
      window.addEventListener('popstate', checkForChanges);

      // Override pushState and replaceState
      const originalPushState = history.pushState;
      const originalReplaceState = history.replaceState;

      history.pushState = function(...args) {
        originalPushState.apply(history, args);
        setTimeout(checkForChanges, 100);
      };

      history.replaceState = function(...args) {
        originalReplaceState.apply(history, args);
        setTimeout(checkForChanges, 100);
      };
    }

    /**
     * Notify background script of page changes
     */
    notifyPageChange() {
      try {
        browser.runtime.sendMessage({
          type: 'PAGE_CHANGED',
          metadata: this.metadata
        });
      } catch (error) {
        console.warn('Safari Page Analyzer: Failed to notify page change:', error);
      }
    }

    /**
     * Initialize page analyzer
     */
    init() {
      // Extract initial metadata
      this.extractPageMetadata();

      // Start monitoring for SPAs
      this.startPageMonitoring();

      console.log('Safari Page Analyzer: Initialized for', window.location.href);
    }
  }

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      const analyzer = new SafariPageAnalyzer();
      analyzer.init();
    });
  } else {
    const analyzer = new SafariPageAnalyzer();
    analyzer.init();
  }

  // Export for testing
  if (typeof module !== 'undefined' && module.exports) {
    module.exports = SafariPageAnalyzer;
  }

})();