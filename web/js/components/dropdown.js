import { api } from '../api.js';
import { UI_CONSTANTS } from '../constants/tickets.js';

/**
 * Configuration for a dropdown menu item
 * @typedef {Object} DropdownItem
 * @property {string} id - Item ID
 * @property {string} label - Display label
 * @property {boolean} [isSelected] - Whether item is currently selected
 */

/**
 * Configuration for creating a dropdown menu
 * @typedef {Object} DropdownConfig
 * @property {string} menuClass - CSS class for the dropdown menu
 * @property {string} itemClass - CSS class for dropdown items
 * @property {Array<DropdownItem>} items - Items to display in dropdown
 * @property {string} ticketProperty - Property name to update on ticket (e.g., 'assigned_to', 'project_id')
 * @property {Function} onSuccess - Callback after successful update
 * @property {boolean} [allowClear] - Whether to show "Clear" option
 * @property {string} [clearLabel] - Label for clear option
 */

/**
 * Closes all open dropdown menus
 */
export function closeAllDropdowns() {
    document.querySelectorAll('.assign-dropdown-menu, .project-dropdown-menu, .sprint-dropdown-menu').forEach(d => d.remove());
}

/**
 * Creates and displays a dropdown menu
 * @param {HTMLElement} triggerElement - Element that triggered the dropdown
 * @param {DropdownConfig} config - Dropdown configuration
 */
export function createDropdown(triggerElement, config) {
    const {
        menuClass,
        itemClass,
        items,
        ticketProperty,
        onSuccess,
        allowClear = true,
        clearLabel = 'Clear'
    } = config;

    // Check if dropdown already exists
    const existingDropdown = document.querySelector(`.${menuClass}`);
    if (existingDropdown) {
        closeAllDropdowns();
        return;
    }

    // Close any other dropdowns
    closeAllDropdowns();

    const ticketId = triggerElement.dataset.ticketId;
    const rect = triggerElement.getBoundingClientRect();

    // Create dropdown element
    const dropdown = document.createElement('div');
    dropdown.className = menuClass;

    // Build dropdown items HTML
    const itemsHtml = items.map(item => `
        <div class="${itemClass}" data-id="${item.id}" style="cursor: pointer;">
            ${item.isSelected ? '✓ ' : ''}${item.label}
        </div>
    `).join('');

    // Add clear option if allowed
    const clearHtml = allowClear ? `
        <div class="${itemClass}" data-id="" style="cursor: pointer; border-top: 1px solid var(--border); margin-top: 0.25rem; padding-top: 0.25rem;">
            ${clearLabel}
        </div>
    ` : '';

    dropdown.innerHTML = itemsHtml + clearHtml;

    // Position dropdown
    dropdown.style.position = 'fixed';
    dropdown.style.top = `${rect.bottom + UI_CONSTANTS.DROPDOWN_OFFSET_PX}px`;
    dropdown.style.left = `${rect.left}px`;
    dropdown.style.zIndex = '1000';

    document.body.appendChild(dropdown);

    // Handle item clicks
    const itemElements = dropdown.querySelectorAll(`.${itemClass}`);
    itemElements.forEach(itemElement => {
        itemElement.addEventListener('click', async function(e) {
            e.stopPropagation();
            const selectedId = this.dataset.id;

            try {
                // Fetch current ticket data
                const ticket = await api.getTicket(ticketId);

                // Update the specified property
                ticket[ticketProperty] = selectedId || null;

                // Update ticket via API
                await api.updateTicket(ticketId, ticket);

                // Close dropdown
                dropdown.remove();

                // Call success callback
                if (onSuccess) {
                    onSuccess();
                }
            } catch (error) {
                console.error('Failed to update ticket:', error);
                alert(`Failed to update ticket: ${error.message}`);
            }
        });
    });

    // Close dropdown when clicking outside
    const closeDropdown = (e) => {
        if (!dropdown.contains(e.target) && e.target !== triggerElement) {
            dropdown.remove();
            document.removeEventListener('click', closeDropdown);
        }
    };

    // Use setTimeout to avoid immediate closure
    setTimeout(() => {
        document.addEventListener('click', closeDropdown);
    }, 0);
}

/**
 * Setup dropdown for assignment selection
 * @param {HTMLElement} link - Trigger element
 * @param {Array} users - Available users
 * @param {string} currentAssigneeId - Currently assigned user ID
 * @param {Function} onSuccess - Success callback
 */
export function createAssignDropdown(link, users, currentAssigneeId, onSuccess) {
    const items = users.map(user => ({
        id: user.id,
        label: user.name,
        isSelected: user.id === currentAssigneeId
    }));

    createDropdown(link, {
        menuClass: 'assign-dropdown-menu',
        itemClass: 'assign-dropdown-item',
        items,
        ticketProperty: 'assigned_to',
        onSuccess,
        allowClear: true,
        clearLabel: 'Unassigned'
    });
}

/**
 * Setup dropdown for project selection
 * @param {HTMLElement} link - Trigger element
 * @param {Array} projects - Available projects
 * @param {string} currentProjectId - Currently selected project ID
 * @param {Function} onSuccess - Success callback
 */
export function createProjectDropdown(link, projects, currentProjectId, onSuccess) {
    const items = projects.map(project => ({
        id: project.id,
        label: project.name,
        isSelected: project.id === currentProjectId
    }));

    createDropdown(link, {
        menuClass: 'project-dropdown-menu',
        itemClass: 'project-dropdown-item',
        items,
        ticketProperty: 'project_id',
        onSuccess,
        allowClear: true,
        clearLabel: 'No Project'
    });
}

/**
 * Setup dropdown for sprint selection
 * @param {HTMLElement} link - Trigger element
 * @param {Array} sprints - Available sprints
 * @param {string} currentSprintId - Currently selected sprint ID
 * @param {boolean} hasSize - Whether ticket has a size (required for sprint assignment)
 * @param {Function} onSuccess - Success callback
 */
export function createSprintDropdown(link, sprints, currentSprintId, hasSize, onSuccess) {
    // Filter to only active sprints and sort by start date
    const activeSprints = sprints
        .filter(s => s.status === 'active')
        .sort((a, b) => new Date(b.start_date) - new Date(a.start_date));

    if (!hasSize) {
        alert('Tickets must be sized before adding to a sprint');
        return;
    }

    const items = activeSprints.map(sprint => ({
        id: sprint.id,
        label: sprint.name,
        isSelected: sprint.id === currentSprintId
    }));

    createDropdown(link, {
        menuClass: 'sprint-dropdown-menu',
        itemClass: 'sprint-dropdown-item',
        items,
        ticketProperty: 'sprint_id',
        onSuccess,
        allowClear: true,
        clearLabel: 'No Sprint'
    });
}
