import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function sprintsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Sprints</h1>
                <a href="#/sprints/new" class="btn btn-primary">Create Sprint</a>
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
                    <a href="#/sprints/new" class="btn btn-primary" style="margin-top: 1rem;">
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
                                <th>Start Date</th>
                                <th>End Date</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${sprints.map(sprint => `
                                <tr>
                                    <td>
                                        <strong>${escapeHtml(sprint.name)}</strong>
                                    </td>
                                    <td>
                                        ${formatDate(sprint.start_date)}
                                    </td>
                                    <td>
                                        ${formatDate(sprint.end_date)}
                                    </td>
                                    <td>
                                        <div class="actions">
                                            <a href="#/sprints/${sprint.id}/board" class="btn btn-primary action-btn">
                                                Board
                                            </a>
                                            <a href="#/sprints/${sprint.id}" class="btn btn-secondary action-btn">
                                                View
                                            </a>
                                        </div>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;

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

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(sprint.name)}</h1>
                    <div class="actions">
                        <a href="#/sprints/${id}/board" class="btn btn-primary">View Board</a>
                        <a href="#/sprints/${id}/edit" class="btn btn-secondary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn" ${auth.isAdmin() || sprint.user_id === auth.getUser().id ? '' : 'disabled'}>Delete</button>
                    </div>
                </div>
                <div class="card">
                    <div style="display: grid; gap: 1rem;">
                        <div>
                            <label class="form-label">Start Date</label>
                            <p>${formatDate(sprint.start_date)}</p>
                        </div>
                        <div>
                            <label class="form-label">End Date</label>
                            <p>${formatDate(sprint.end_date)}</p>
                        </div>
                    </div>
                </div>
                <div class="card" style="margin-top: 1.5rem;">
                    <h2 style="margin-bottom: 1rem;">Tickets</h2>
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
                                        <th>ID</th>
                                        <th>Title</th>
                                        <th>Status</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${sprintTickets.map(ticket => `
                                        <tr>
                                            <td>
                                                <strong>${ticket.id.substring(0, 6)}</strong>
                                            </td>
                                            <td>
                                                <strong>${escapeHtml(ticket.title)}</strong>
                                            </td>
                                            <td>
                                                <span class="badge ${getStatusBadgeClass(ticket.status)}">
                                                    ${escapeHtml(ticket.status)}
                                                </span>
                                            </td>
                                            <td>
                                                <div class="actions">
                                                    <a href="#/tickets/${ticket.id}" class="btn btn-secondary action-btn">
                                                        View
                                                    </a>
                                                </div>
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
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load sprint: ${error.message}</p>
                <a href="#/sprints" class="btn btn-secondary" style="margin-top: 1rem;">Back to Sprints</a>
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
            <div class="card">
                <form id="sprint-form">
                    <div class="form-group">
                        <label class="form-label" for="name">Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
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
                    <div class="form-group">
                        <label class="form-label">End Date</label>
                        <p style="color: var(--text-secondary);">Automatically set to 2 weeks after start date</p>
                    </div>
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Sprint
                        </button>
                        <a href="${isEdit ? `#/sprints/${id}` : '#/sprints'}" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
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
    const date = new Date(dateString);
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
