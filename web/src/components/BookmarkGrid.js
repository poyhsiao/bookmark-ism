/**
 * BookmarkGrid Component
 * 視覺化書籤網格介面組件
 */

class BookmarkGrid {
    constructor(container, options = {}) {
        this.container = container;
        this.options = {
            gridSize: options.gridSize || 'medium',
            showThumbnails: options.showThumbnails !== false,
            enableDragDrop: options.enableDragDrop !== false,
            itemsPerRow: options.itemsPerRow || 'auto',
            ...options
        };

        this.bookmarks = [];
        this.selectedItems = new Set();
        this.draggedItem = null;

        this.init();
    }

    init() {
        this.createGridContainer();
        this.setupEventListeners();
        this.applyGridStyles();
    }

    createGridContainer() {
        this.container.innerHTML = `
            <div class="bookmark-grid-container">
                <div class="grid-controls">
                    <div class="view-controls">
                        <button class="grid-size-btn" data-size="small" title="Small Grid">
                            <svg width="16" height="16" viewBox="0 0 16 16">
                                <rect x="1" y="1" width="6" height="6" fill="currentColor"/>
                                <rect x="9" y="1" width="6" height="6" fill="currentColor"/>
                                <rect x="1" y="9" width="6" height="6" fill="currentColor"/>
                                <rect x="9" y="9" width="6" height="6" fill="currentColor"/>
                            </svg>
                        </button>
                        <button class="grid-size-btn" data-size="medium" title="Medium Grid">
                            <svg width="16" height="16" viewBox="0 0 16 16">
                                <rect x="1" y="1" width="7" height="7" fill="currentColor"/>
                                <rect x="9" y="1" width="6" height="7" fill="currentColor"/>
                                <rect x="1" y="9" width="7" height="6" fill="currentColor"/>
                                <rect x="9" y="9" width="6" height="6" fill="currentColor"/>
                            </svg>
                        </button>
                        <button class="grid-size-btn" data-size="large" title="Large Grid">
                            <svg width="16" height="16" viewBox="0 0 16 16">
                                <rect x="1" y="1" width="14" height="6" fill="currentColor"/>
                                <rect x="1" y="9" width="14" height="6" fill="currentColor"/>
                            </svg>
                        </button>
                    </div>
                    <div class="sort-controls">
                        <select class="sort-select">
                            <option value="created_at">Recently Added</option>
                            <option value="updated_at">Recently Updated</option>
                            <option value="title">Title</option>
                            <option value="url">URL</option>
                        </select>
                    </div>
                </div>
                <div class="bookmark-grid" id="bookmark-grid"></div>
                <div class="grid-loading" style="display: none;">
                    <div class="loading-spinner"></div>
                    <p>Loading bookmarks...</p>
                </div>
                <div class="grid-empty" style="display: none;">
                    <div class="empty-icon">📚</div>
                    <h3>No bookmarks yet</h3>
                    <p>Start by adding your first bookmark!</p>
                </div>
            </div>
        `;

        this.gridElement = this.container.querySelector('.bookmark-grid');
        this.loadingElement = this.container.querySelector('.grid-loading');
        this.emptyElement = this.container.querySelector('.grid-empty');
    }

    setupEventListeners() {
        // Grid size controls
        this.container.querySelectorAll('.grid-size-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const size = e.currentTarget.dataset.size;
                this.setGridSize(size);
            });
        });

        // Sort controls
        const sortSelect = this.container.querySelector('.sort-select');
        sortSelect.addEventListener('change', (e) => {
            this.sortBookmarks(e.target.value);
        });

        // Drag and drop
        if (this.options.enableDragDrop) {
            this.setupDragAndDrop();
        }
    }

    setupDragAndDrop() {
        this.gridElement.addEventListener('dragstart', (e) => {
            if (e.target.classList.contains('bookmark-item')) {
                this.draggedItem = e.target;
                e.target.classList.add('dragging');
                e.dataTransfer.effectAllowed = 'move';
                e.dataTransfer.setData('text/html', e.target.outerHTML);
            }
        });

        this.gridElement.addEventListener('dragend', (e) => {
            if (e.target.classList.contains('bookmark-item')) {
                e.target.classList.remove('dragging');
                this.draggedItem = null;
            }
        });

        this.gridElement.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
        });

        this.gridElement.addEventListener('drop', (e) => {
            e.preventDefault();
            if (this.draggedItem) {
                const dropTarget = e.target.closest('.bookmark-item');
                if (dropTarget && dropTarget !== this.draggedItem) {
                    this.reorderBookmarks(this.draggedItem, dropTarget);
                }
            }
        });
    }

    applyGridStyles() {
        const styles = `
            .bookmark-grid-container {
                width: 100%;
                height: 100%;
                display: flex;
                flex-direction: column;
            }

            .grid-controls {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 16px;
                border-bottom: 1px solid #e1e5e9;
                background: #f8f9fa;
            }

            .view-controls {
                display: flex;
                gap: 8px;
            }

            .grid-size-btn {
                padding: 8px;
                border: 1px solid #d0d7de;
                background: white;
                border-radius: 6px;
                cursor: pointer;
                transition: all 0.2s;
            }

            .grid-size-btn:hover {
                background: #f3f4f6;
                border-color: #8b949e;
            }

            .grid-size-btn.active {
                background: #0969da;
                border-color: #0969da;
                color: white;
            }

            .sort-select {
                padding: 6px 12px;
                border: 1px solid #d0d7de;
                border-radius: 6px;
                background: white;
                font-size: 14px;
            }

            .bookmark-grid {
                flex: 1;
                padding: 20px;
                display: grid;
                gap: 16px;
                overflow-y: auto;
            }

            .bookmark-grid.size-small {
                grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            }

            .bookmark-grid.size-medium {
                grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
            }

            .bookmark-grid.size-large {
                grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
            }

            .bookmark-item {
                background: white;
                border: 1px solid #d0d7de;
                border-radius: 12px;
                overflow: hidden;
                transition: all 0.2s;
                cursor: pointer;
                position: relative;
            }

            .bookmark-item:hover {
                border-color: #0969da;
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
                transform: translateY(-2px);
            }

            .bookmark-item.selected {
                border-color: #0969da;
                box-shadow: 0 0 0 2px rgba(9, 105, 218, 0.3);
            }

            .bookmark-item.dragging {
                opacity: 0.5;
                transform: rotate(5deg);
            }

            .bookmark-thumbnail {
                width: 100%;
                height: 160px;
                background: #f6f8fa;
                display: flex;
                align-items: center;
                justify-content: center;
                overflow: hidden;
                position: relative;
            }

            .bookmark-thumbnail img {
                width: 100%;
                height: 100%;
                object-fit: cover;
            }

            .bookmark-thumbnail .favicon {
                width: 32px;
                height: 32px;
                border-radius: 4px;
            }

            .bookmark-thumbnail .placeholder {
                color: #656d76;
                font-size: 48px;
            }

            .bookmark-content {
                padding: 16px;
            }

            .bookmark-title {
                font-size: 16px;
                font-weight: 600;
                color: #24292f;
                margin: 0 0 8px 0;
                line-height: 1.3;
                display: -webkit-box;
                -webkit-line-clamp: 2;
                -webkit-box-orient: vertical;
                overflow: hidden;
            }

            .bookmark-description {
                font-size: 14px;
                color: #656d76;
                margin: 0 0 12px 0;
                line-height: 1.4;
                display: -webkit-box;
                -webkit-line-clamp: 3;
                -webkit-box-orient: vertical;
                overflow: hidden;
            }

            .bookmark-url {
                font-size: 12px;
                color: #8b949e;
                text-decoration: none;
                display: block;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .bookmark-meta {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-top: 12px;
                padding-top: 12px;
                border-top: 1px solid #f1f3f4;
            }

            .bookmark-tags {
                display: flex;
                flex-wrap: wrap;
                gap: 4px;
            }

            .bookmark-tag {
                background: #ddf4ff;
                color: #0969da;
                padding: 2px 8px;
                border-radius: 12px;
                font-size: 12px;
                font-weight: 500;
            }

            .bookmark-date {
                font-size: 12px;
                color: #8b949e;
            }

            .bookmark-actions {
                position: absolute;
                top: 8px;
                right: 8px;
                opacity: 0;
                transition: opacity 0.2s;
            }

            .bookmark-item:hover .bookmark-actions {
                opacity: 1;
            }

            .action-btn {
                background: rgba(255, 255, 255, 0.9);
                border: 1px solid #d0d7de;
                border-radius: 6px;
                padding: 4px;
                margin-left: 4px;
                cursor: pointer;
                transition: all 0.2s;
            }

            .action-btn:hover {
                background: white;
                border-color: #8b949e;
            }

            .grid-loading, .grid-empty {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                padding: 60px 20px;
                color: #656d76;
            }

            .loading-spinner {
                width: 32px;
                height: 32px;
                border: 3px solid #f1f3f4;
                border-top: 3px solid #0969da;
                border-radius: 50%;
                animation: spin 1s linear infinite;
                margin-bottom: 16px;
            }

            .empty-icon {
                font-size: 48px;
                margin-bottom: 16px;
            }

            @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
            }

            @media (max-width: 768px) {
                .bookmark-grid.size-small,
                .bookmark-grid.size-medium,
                .bookmark-grid.size-large {
                    grid-template-columns: 1fr;
                }

                .grid-controls {
                    flex-direction: column;
                    gap: 12px;
                }
            }
        `;

        // Add styles to document if not already added
        if (!document.getElementById('bookmark-grid-styles')) {
            const styleSheet = document.createElement('style');
            styleSheet.id = 'bookmark-grid-styles';
            styleSheet.textContent = styles;
            document.head.appendChild(styleSheet);
        }
    }

    setGridSize(size) {
        this.options.gridSize = size;
        this.gridElement.className = `bookmark-grid size-${size}`;

        // Update active button
        this.container.querySelectorAll('.grid-size-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.size === size);
        });

        // Save preference
        localStorage.setItem('bookmarkGridSize', size);
    }

    loadBookmarks(bookmarks) {
        this.bookmarks = bookmarks;
        this.renderBookmarks();
    }

    renderBookmarks() {
        if (this.bookmarks.length === 0) {
            this.showEmpty();
            return;
        }

        this.hideLoading();
        this.hideEmpty();

        this.gridElement.innerHTML = this.bookmarks.map(bookmark =>
            this.renderBookmarkItem(bookmark)
        ).join('');

        // Set initial grid size
        this.setGridSize(this.options.gridSize);

        // Setup item event listeners
        this.setupItemEventListeners();
    }

    renderBookmarkItem(bookmark) {
        const thumbnailContent = this.renderThumbnail(bookmark);
        const tags = bookmark.tags || [];
        const formattedDate = this.formatDate(bookmark.created_at);

        return `
            <div class="bookmark-item" data-id="${bookmark.id}" draggable="${this.options.enableDragDrop}">
                <div class="bookmark-actions">
                    <button class="action-btn" data-action="edit" title="Edit">
                        <svg width="14" height="14" viewBox="0 0 16 16">
                            <path fill="currentColor" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"/>
                        </svg>
                    </button>
                    <button class="action-btn" data-action="delete" title="Delete">
                        <svg width="14" height="14" viewBox="0 0 16 16">
                            <path fill="currentColor" d="M6.5 1.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V3h-3V1.75zm4.5 0V3h2.25a.75.75 0 010 1.5H2.75a.75.75 0 010-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75zM4.496 6.675a.75.75 0 10-1.492.15l.66 6.6A1.75 1.75 0 005.405 15h5.19c.9 0 1.652-.681 1.741-1.575l.66-6.6a.75.75 0 00-1.492-.15l-.66 6.6a.25.25 0 01-.249.225H5.405a.25.25 0 01-.249-.225l-.66-6.6z"/>
                        </svg>
                    </button>
                </div>
                <div class="bookmark-thumbnail">
                    ${thumbnailContent}
                </div>
                <div class="bookmark-content">
                    <h3 class="bookmark-title">${this.escapeHtml(bookmark.title)}</h3>
                    ${bookmark.description ? `<p class="bookmark-description">${this.escapeHtml(bookmark.description)}</p>` : ''}
                    <a href="${bookmark.url}" class="bookmark-url" target="_blank" rel="noopener">${bookmark.url}</a>
                    <div class="bookmark-meta">
                        <div class="bookmark-tags">
                            ${tags.slice(0, 3).map(tag => `<span class="bookmark-tag">${this.escapeHtml(tag)}</span>`).join('')}
                            ${tags.length > 3 ? `<span class="bookmark-tag">+${tags.length - 3}</span>` : ''}
                        </div>
                        <span class="bookmark-date">${formattedDate}</span>
                    </div>
                </div>
            </div>
        `;
    }

    renderThumbnail(bookmark) {
        if (this.options.showThumbnails && bookmark.screenshot) {
            return `<img src="${bookmark.screenshot}" alt="Screenshot" loading="lazy">`;
        } else if (bookmark.favicon) {
            return `<img src="${bookmark.favicon}" alt="Favicon" class="favicon">`;
        } else {
            return `<div class="placeholder">🔖</div>`;
        }
    }

    setupItemEventListeners() {
        this.gridElement.querySelectorAll('.bookmark-item').forEach(item => {
            // Click to open bookmark
            item.addEventListener('click', (e) => {
                if (!e.target.closest('.bookmark-actions')) {
                    const url = item.querySelector('.bookmark-url').href;
                    window.open(url, '_blank', 'noopener');
                }
            });

            // Action buttons
            item.querySelectorAll('.action-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    const action = btn.dataset.action;
                    const bookmarkId = item.dataset.id;
                    this.handleAction(action, bookmarkId);
                });
            });

            // Selection
            item.addEventListener('contextmenu', (e) => {
                e.preventDefault();
                this.toggleSelection(item);
            });
        });
    }

    handleAction(action, bookmarkId) {
        const bookmark = this.bookmarks.find(b => b.id === bookmarkId);
        if (!bookmark) return;

        switch (action) {
            case 'edit':
                this.onEdit?.(bookmark);
                break;
            case 'delete':
                this.onDelete?.(bookmark);
                break;
        }
    }

    toggleSelection(item) {
        const bookmarkId = item.dataset.id;
        if (this.selectedItems.has(bookmarkId)) {
            this.selectedItems.delete(bookmarkId);
            item.classList.remove('selected');
        } else {
            this.selectedItems.add(bookmarkId);
            item.classList.add('selected');
        }
    }

    sortBookmarks(sortBy) {
        this.bookmarks.sort((a, b) => {
            switch (sortBy) {
                case 'title':
                    return a.title.localeCompare(b.title);
                case 'url':
                    return a.url.localeCompare(b.url);
                case 'updated_at':
                    return new Date(b.updated_at) - new Date(a.updated_at);
                case 'created_at':
                default:
                    return new Date(b.created_at) - new Date(a.created_at);
            }
        });
        this.renderBookmarks();
    }

    reorderBookmarks(draggedItem, dropTarget) {
        const draggedId = draggedItem.dataset.id;
        const dropTargetId = dropTarget.dataset.id;

        const draggedIndex = this.bookmarks.findIndex(b => b.id === draggedId);
        const dropTargetIndex = this.bookmarks.findIndex(b => b.id === dropTargetId);

        if (draggedIndex !== -1 && dropTargetIndex !== -1) {
            const [draggedBookmark] = this.bookmarks.splice(draggedIndex, 1);
            this.bookmarks.splice(dropTargetIndex, 0, draggedBookmark);
            this.renderBookmarks();

            // Notify parent about reorder
            this.onReorder?.(draggedId, dropTargetId);
        }
    }

    showLoading() {
        this.gridElement.style.display = 'none';
        this.emptyElement.style.display = 'none';
        this.loadingElement.style.display = 'flex';
    }

    hideLoading() {
        this.loadingElement.style.display = 'none';
        this.gridElement.style.display = 'grid';
    }

    showEmpty() {
        this.gridElement.style.display = 'none';
        this.loadingElement.style.display = 'none';
        this.emptyElement.style.display = 'flex';
    }

    hideEmpty() {
        this.emptyElement.style.display = 'none';
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffTime = Math.abs(now - date);
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

        if (diffDays === 1) return 'Today';
        if (diffDays === 2) return 'Yesterday';
        if (diffDays <= 7) return `${diffDays} days ago`;
        if (diffDays <= 30) return `${Math.ceil(diffDays / 7)} weeks ago`;
        if (diffDays <= 365) return `${Math.ceil(diffDays / 30)} months ago`;
        return `${Math.ceil(diffDays / 365)} years ago`;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Public API methods
    addBookmark(bookmark) {
        this.bookmarks.unshift(bookmark);
        this.renderBookmarks();
    }

    updateBookmark(bookmarkId, updates) {
        const index = this.bookmarks.findIndex(b => b.id === bookmarkId);
        if (index !== -1) {
            this.bookmarks[index] = { ...this.bookmarks[index], ...updates };
            this.renderBookmarks();
        }
    }

    removeBookmark(bookmarkId) {
        this.bookmarks = this.bookmarks.filter(b => b.id !== bookmarkId);
        this.renderBookmarks();
    }

    getSelectedBookmarks() {
        return Array.from(this.selectedItems);
    }

    clearSelection() {
        this.selectedItems.clear();
        this.gridElement.querySelectorAll('.bookmark-item.selected').forEach(item => {
            item.classList.remove('selected');
        });
    }

    // Event handlers (to be set by parent)
    onEdit = null;
    onDelete = null;
    onReorder = null;
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = BookmarkGrid;
} else if (typeof window !== 'undefined') {
    window.BookmarkGrid = BookmarkGrid;
}