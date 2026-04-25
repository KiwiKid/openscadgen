/**
 * HTMX Error Handler - Shows dismissable Bulma toast notifications for HTMX errors
 */

(function() {
    'use strict';

    /**
     * Creates and shows a dismissable Bulma notification toast
     * @param {string} message - The error message to display
     * @param {string} type - The notification type (is-danger, is-warning, etc.)
     */
    function showToast(message, type = 'is-danger') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 9999; max-width: 400px; min-width: 300px;';
        
        // Create message content
        const messageDiv = document.createElement('div');
        messageDiv.innerHTML = `<strong>Error:</strong> ${escapeHtml(message)}`;
        
        // Create dismiss button
        const deleteButton = document.createElement('button');
        deleteButton.className = 'delete';
        deleteButton.setAttribute('aria-label', 'delete');
        deleteButton.onclick = function() {
            notification.remove();
        };
        
        // Assemble notification
        notification.appendChild(messageDiv);
        notification.appendChild(deleteButton);
        
        // Add to page (try to find a toast container, otherwise append to body)
        let container = document.getElementById('toast-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'toast-container';
            container.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 9999; display: flex; flex-direction: column; gap: 10px;';
            document.body.appendChild(container);
        }
        
        container.appendChild(notification);
        
        // Auto-dismiss after 10 seconds
        setTimeout(function() {
            if (notification.parentNode) {
                notification.style.transition = 'opacity 0.3s';
                notification.style.opacity = '0';
                setTimeout(function() {
                    notification.remove();
                }, 300);
            }
        }, 10000);
    }

    /**
     * Escapes HTML to prevent XSS
     * @param {string} text - Text to escape
     * @returns {string} Escaped text
     */
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * Extracts error message from HTMX event
     * @param {Event} event - HTMX event
     * @returns {string} Error message
     */
    function extractErrorMessage(event) {
        // Try to get error from xhr response
        if (event.detail && event.detail.xhr) {
            const xhr = event.detail.xhr;
            if (xhr.responseText) {
                // Try to parse as JSON first
                try {
                    const json = JSON.parse(xhr.responseText);
                    if (json.error || json.message) {
                        return json.error || json.message;
                    }
                } catch (e) {
                    // Not JSON, use response text (limit length)
                    const text = xhr.responseText.trim();
                    return text.length > 200 ? text.substring(0, 200) + '...' : text;
                }
            }
            // Fallback to status text
            if (xhr.statusText) {
                return `${xhr.status}: ${xhr.statusText}`;
            }
        }
        
        // Fallback to event type
        return event.type.replace('htmx:', '');
    }

    // Wait for DOM and HTMX to be ready
    function init() {
        if (typeof htmx === 'undefined') {
            console.warn('HTMX not found, error handler not initialized');
            return;
        }

        // Listen for HTMX response errors (4xx, 5xx status codes)
        document.body.addEventListener('htmx:responseError', function(event) {
            const message = extractErrorMessage(event);
            showToast(message, 'is-danger');
        });

        // Listen for HTMX send errors (network errors, etc.)
        document.body.addEventListener('htmx:sendError', function(event) {
            let message = 'Network error: Failed to send request';
            if (event.detail && event.detail.error) {
                message = `Network error: ${event.detail.error.message || event.detail.error}`;
            }
            showToast(message, 'is-danger');
        });

        // Listen for timeout errors
        document.body.addEventListener('htmx:timeout', function(event) {
            showToast('Request timed out. Please try again.', 'is-warning');
        });
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();

/**
 * OpenSCAD nav widget (report / config UIs). Keeps `isExpanded=1` in the query string
 * so refresh preserves open/close; htmx reload swaps the fragment and Alpine re-inits.
 */
function openscadNavXData() {
    return {
        detailsOpen: (function () {
            try {
                return new URLSearchParams(window.location.search).get('isExpanded') === '1';
            } catch (e) {
                return false;
            }
        })(),
        syncURL() {
            try {
                const u = new URL(window.location.href);
                if (this.detailsOpen) {
                    u.searchParams.set('isExpanded', '1');
                } else {
                    u.searchParams.delete('isExpanded');
                }
                history.replaceState(null, '', u);
            } catch (e) { /* no-op */ }
        },
        toggle() {
            this.detailsOpen = !this.detailsOpen;
            this.syncURL();
        },
    };
}

function openScadNavAfterHtmxSwapInit(evt) {
    var el = evt.detail.elt;
    if (!el || el.id !== 'openscad-nav-widget') return;
    var attempts = 0;
    (function tryInit() {
        if (!el.isConnected) return;
        if (window.Alpine && typeof window.Alpine.initTree === 'function') {
            window.Alpine.initTree(el);
            return;
        }
        if (++attempts < 50) {
            setTimeout(tryInit, 20);
        }
    })();
}
if (document.body) {
    document.body.addEventListener('htmx:afterSwap', openScadNavAfterHtmxSwapInit);
} else {
    document.addEventListener('DOMContentLoaded', function () {
        document.body.addEventListener('htmx:afterSwap', openScadNavAfterHtmxSwapInit);
    });
}




