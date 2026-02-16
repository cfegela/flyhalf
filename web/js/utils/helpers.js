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

    // Remove all whitespace for acronym generation
    const baseName = projectName.replace(/\s+/g, '');
    let length = 6;
    let acronym;

    // Start with 6 characters and extend if there's a collision
    while (length <= baseName.length) {
        acronym = baseName.substring(0, length).toUpperCase();

        // Check if any other project has this same acronym
        const hasCollision = allProjects.some(p => {
            if (p.id === currentProjectId) return false; // Don't compare with self
            const otherBaseName = p.name.replace(/\s+/g, '');
            const otherAcronym = otherBaseName.substring(0, length).toUpperCase();
            return acronym === otherAcronym;
        });

        if (!hasCollision) {
            return acronym;
        }

        length++;
    }

    // If we've used all characters and still have collision, return full name uppercase
    return baseName.toUpperCase();
}
