import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function sprintsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Sprints</h1>
                <a href="/sprints/new" class="btn btn-primary">Create Sprint</a>
            </div>
            <div id="sprints-container">
                <div class="loading">Loading sprints...</div>
            </div>
        </div>
    `;

    try {
        const sprints = await api.getSprints();
        const sprintsContainer = container.querySelector('#sprints-container');

        if (sprints.length === 0) {
            sprintsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">🏃</div>
                    <h2>No sprints yet</h2>
                    <p>Create your first sprint to get started</p>
                    <a href="/sprints/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create Sprint
                    </a>
                </div>
            `;
            return;
        }

        sprintsContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Status</th>
                                <th>Start Date</th>
                                <th>End Date</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${sprints.map(sprint => {
                                const statusBadgeMap = {
                                    'active': 'badge-in-progress',
                                    'completed': 'badge-closed',
                                    'upcoming': 'badge-open'
                                };
                                const statusLabelMap = {
                                    'active': 'Active',
                                    'completed': 'Completed',
                                    'upcoming': 'Upcoming'
                                };

                                return `
                                <tr data-sprint-id="${sprint.id}">
                                    <td data-label="Name">
                                        <strong>${escapeHtml(sprint.name)}</strong>
                                    </td>
                                    <td data-label="Status">
                                        <span class="badge ${statusBadgeMap[sprint.status] || 'badge-open'}">
                                            ${statusLabelMap[sprint.status] || sprint.status}
                                        </span>
                                    </td>
                                    <td data-label="Start Date">
                                        ${formatDate(sprint.start_date)}
                                    </td>
                                    <td data-label="End Date">
                                        ${formatDate(sprint.end_date)}
                                    </td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <a href="/sprints/${sprint.id}/board" class="btn btn-primary action-btn">
                                                board
                                            </a>
                                            <a href="/sprints/${sprint.id}/report" class="btn btn-primary action-btn">
                                                report
                                            </a>
                                            <a href="/sprints/${sprint.id}/retro" class="btn btn-primary action-btn">
                                                retro
                                            </a>
                                            <a href="/sprints/${sprint.id}" class="btn btn-secondary action-btn" title="View details">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/System/eye-fill.svg" alt="View" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                            <a href="/sprints/${sprint.id}/edit" class="btn btn-secondary action-btn" title="Edit sprint">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/Design/pencil-ai-fill.svg" alt="Edit" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                        </div>
                                    </td>
                                </tr>
                                `;
                            }).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;

        // Removed clickable row behavior - users now click buttons instead
        /*
        const clickableRows = sprintsContainer.querySelectorAll('.clickable-row');
        clickableRows.forEach(row => {
            row.addEventListener('click', (e) => {
                const sprintId = row.dataset.sprintId;
                router.navigate(`/sprints/${sprintId}`);
            });
        });

        // Prevent Board and Report button clicks from triggering row navigation
        const boardButtons = sprintsContainer.querySelectorAll('.board-btn');
        boardButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
            });
        });

        const reportButtons = sprintsContainer.querySelectorAll('.report-btn');
        reportButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
            });
        });
        */

    } catch (error) {
        const sprintsContainer = container.querySelector('#sprints-container');
        sprintsContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load sprints: ${error.message}</p>
            </div>
        `;
    }
}

export async function sprintDetailView(params) {
    const container = document.getElementById('view-container');
    const [id] = params;

    if (!id) {
        router.navigate('/sprints');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading sprint...</div>
        </div>
    `;

    try {
        const sprint = await api.getSprint(id);
        const allTickets = await api.getTickets();

        // Filter tickets for this sprint
        const sprintTickets = allTickets.filter(ticket => ticket.sprint_id === id);

        // Calculate sprint duration and progress
        // Parse dates as local dates to avoid timezone issues
        const parseDate = (dateStr) => {
            const [year, month, day] = dateStr.split('T')[0].split('-').map(Number);
            return new Date(year, month - 1, day);
        };
        const startDate = parseDate(sprint.start_date);
        const endDate = parseDate(sprint.end_date);
        const today = new Date();
        const totalDays = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24));
        const daysElapsed = Math.max(0, Math.ceil((today - startDate) / (1000 * 60 * 60 * 24)));
        const daysRemaining = Math.max(0, Math.ceil((endDate - today) / (1000 * 60 * 60 * 24)));
        const isActive = sprint.status === 'active';
        const isCompleted = sprint.status === 'completed';
        const isUpcoming = sprint.status === 'upcoming';

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(sprint.name)}</h1>
                    <div class="actions">
                        <a href="/sprints/${id}/board" class="btn btn-primary">View Board</a>
                        <a href="/sprints/${id}/report" class="btn btn-primary">View Report</a>
                        <a href="/sprints/${id}/retro" class="btn btn-primary">View Retro</a>
                        <a href="/sprints/${id}/edit" class="btn btn-secondary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn" ${auth.isAdmin() || sprint.user_id === auth.getUser().id ? '' : 'disabled'}>Delete</button>
                    </div>
                </div>

                <!-- Sprint Information Card -->
                <div class="card">
                    <h2 class="card-header">Sprint Details</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Status</label>
                            <div style="margin-top: 0.25rem;">
                                <span class="badge ${isActive ? 'badge-in-progress' : isCompleted ? 'badge-closed' : 'badge-open'}" style="font-size: 0.875rem; padding: 0.375rem 0.875rem;">
                                    ${isActive ? 'Active' : isCompleted ? 'Completed' : 'Upcoming'}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="form-label">Duration</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${totalDays} days
                            </p>
                        </div>
                        <div>
                            <label class="form-label">${isCompleted ? 'Completed' : isUpcoming ? 'Starts In' : 'Days Remaining'}</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${isCompleted ? formatDate(sprint.end_date) : isUpcoming ? `${Math.abs(daysElapsed)} days` : `${daysRemaining} days`}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Date Information Card -->
                <div class="card">
                    <h2 class="card-header">Timeline</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Start Date</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${formatDate(sprint.start_date)}
                            </p>
                        </div>
                        <div>
                            <label class="form-label">End Date</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${formatDate(sprint.end_date)}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Tickets Card -->
                <div class="card">
                    <h2 class="card-header">Tickets (${sprintTickets.length})</h2>
                    ${sprintTickets.length === 0 ? `
                        <div class="empty-state">
                            <div class="empty-state-icon">🎫</div>
                            <p>No tickets assigned to this sprint</p>
                        </div>
                    ` : `
                        <div class="table-container">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Title</th>
                                        <th>Status</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${sprintTickets.map(ticket => `
                                        <tr class="clickable-row" data-ticket-id="${ticket.id}" style="cursor: pointer;">
                                            <td data-label="Title">
                                                <strong>${escapeHtml(ticket.title)}</strong>
                                            </td>
                                            <td data-label="Status">
                                                <span class="badge ${getStatusBadgeClass(ticket.status)}">
                                                    ${escapeHtml(ticket.status)}
                                                </span>
                                            </td>
                                        </tr>
                                    `).join('')}
                                </tbody>
                            </table>
                        </div>
                    `}
                </div>
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        deleteBtn.addEventListener('click', async () => {
            if (deleteBtn.disabled) return;
            if (confirm('Are you sure you want to delete this sprint?')) {
                try {
                    await api.deleteSprint(id);
                    router.navigate('/sprints');
                } catch (error) {
                }
            }
        });

        // Make ticket rows clickable to navigate to ticket details
        const clickableRows = container.querySelectorAll('.clickable-row');
        clickableRows.forEach(row => {
            row.addEventListener('click', (e) => {
                const ticketId = row.dataset.ticketId;
                router.navigate(`/tickets/${ticketId}`);
            });
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load sprint: ${error.message}</p>
                <a href="/sprints" class="btn btn-secondary" style="margin-top: 1rem;">Back to Sprints</a>
            </div>
        `;
    }
}

export async function sprintFormView(params) {
    const container = document.getElementById('view-container');
    const [id, action] = params;
    const isEdit = action === 'edit';

    let sprint = null;
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading sprint...</div>';
        try {
            sprint = await api.getSprint(id);
        } catch (error) {
            router.navigate('/sprints');
            return;
        }
    }

    // Format date for input field (YYYY-MM-DD)
    const formatDateForInput = (dateString) => {
        const date = new Date(dateString);
        return date.toISOString().split('T')[0];
    };

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Sprint</h1>
            </div>

            <form id="sprint-form">
                <!-- Sprint Information Card -->
                <div class="card">
                    <h2 class="card-header">Sprint Information</h2>
                    <div class="form-group">
                        <label class="form-label" for="name">Sprint Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
                            placeholder="e.g., Sprint 1, Q1 2026 Sprint 3"
                            value="${sprint ? escapeHtml(sprint.name) : ''}"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="start_date">Start Date *</label>
                        <input
                            type="date"
                            id="start_date"
                            class="form-input"
                            required
                            value="${sprint ? formatDateForInput(sprint.start_date) : ''}"
                        >
                    </div>
                    <div class="form-group" style="margin-bottom: 0;">
                        <label class="form-label">End Date</label>
                        <p style="color: var(--text-secondary); margin-top: 0.25rem; line-height: 1.6;">
                            The end date will be automatically set to 2 weeks (14 days) after the start date.
                        </p>
                    </div>
                </div>

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Sprint
                        </button>
                        <a href="${isEdit ? `#/sprints/${id}` : '#/sprints'}" class="btn btn-secondary">Cancel</a>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#sprint-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = form.name.value.trim();
        const start_date = form.start_date.value;

        const data = { name, start_date };

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateSprint(id, data);
                router.navigate('/sprints');
            } else {
                await api.createSprint(data);
                router.navigate('/sprints');
            }
        } catch (error) {
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} Sprint`;
        }
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateString) {
    // Parse date string as local date to avoid timezone conversion issues
    const [year, month, day] = dateString.split('T')[0].split('-').map(Number);
    const date = new Date(year, month - 1, day);
    return date.toLocaleDateString();
}

function getStatusBadgeClass(status) {
    switch (status) {
        case 'open': return 'badge-open';
        case 'in-progress': return 'badge-in-progress';
        case 'blocked': return 'badge-blocked';
        case 'needs-review': return 'badge-needs-review';
        case 'closed': return 'badge-closed';
        default: return 'badge-open';
    }
}
