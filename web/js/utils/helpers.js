/**
 * Creates a map/dictionary from an array of objects by their ID
 * @param {Array} items - Array of objects with 'id' property
 * @returns {Object} - Map of id -> object
 */
export function createIdMap(items) {
    const map = {};
    items.forEach(item => {
        map[item.id] = item;
    });
    return map;
}

/**
 * Gets the CSS class for a ticket status badge
 * @param {string} status - Ticket status
 * @returns {string} - CSS class name
 */
export function getStatusBadgeClass(status) {
    switch (status) {
        case 'open': return 'badge-open';
        case 'in-progress': return 'badge-in-progress';
        case 'blocked': return 'badge-blocked';
        case 'needs-review': return 'badge-needs-review';
        case 'closed': return 'badge-closed';
        default: return 'badge-open';
    }
}

/**
 * Gets the display label for a ticket size
 * @param {number} size - Ticket size value
 * @returns {string} - Size label or '-'
 */
export function getSizeLabel(size) {
    if (!size) return '-';
    switch (size) {
        case 1: return 'Small';
        case 2: return 'Medium';
        case 3: return 'Large';
        case 5: return 'X-Large';
        case 8: return 'Danger';
        default: return '-';
    }
}

/**
 * Gets the project acronym or full name
 * @param {string} projectName - Full project name
 * @param {Array} allProjects - Array of all projects
 * @param {string} currentProjectId - Current project ID
 * @returns {string} - Project acronym or name
 */
export function getProjectAcronym(projectName, allProjects, currentProjectId) {
    if (!projectName) return '-';

    // If there's only one project, show full name
    if (allProjects.length === 1) {
        return projectName;
    }

    // Check if this is the active project filter
    const isActiveProject = currentProjectId && allProjects.find(p => p.id === currentProjectId);
    if (isActiveProject) {
        return projectName;
    }

    // Extract acronym from project name
    const words = projectName.split(' ');
    if (words.length === 1) {
        // Single word: take first 3 letters
        return projectName.substring(0, 3).toUpperCase();
    } else {
        // Multiple words: take first letter of each word (max 3)
        return words
            .slice(0, 3)
            .map(word => word[0])
            .join('')
            .toUpperCase();
    }
}
