/**
 * Escapes HTML special characters to prevent XSS attacks
 * @param {string} text - Text to escape
 * @returns {string} - HTML-safe text
 */
export function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Formats a date string into localized date and time
 * @param {string} dateString - ISO date string
 * @returns {string} - Formatted date and time
 */
export function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

/**
 * Formats a date string into localized date only (no time)
 * @param {string} dateString - ISO date string
 * @returns {string} - Formatted date
 */
export function formatDateOnly(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString();
}

/**
 * Formats a date string as relative time (e.g., "5 minutes ago")
 * Falls back to formatted date for dates older than 7 days
 * @param {string} dateString - ISO date string
 * @returns {string} - Relative time or formatted date
 */
export function formatRelativeTime(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffSecs < 60) return 'just now';
    if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    return formatDate(dateString);
}
