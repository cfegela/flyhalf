import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function ticketsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Tickets</h1>
                <a href="#/tickets/new" class="btn btn-primary">Create Ticket</a>
            </div>
            <div id="tickets-container">
                <div class="loading">Loading tickets...</div>
            </div>
        </div>
    `;

    try {
        const tickets = await api.getTickets();
        const epics = await api.getEpics();
        const sprints = await api.getSprints();
        const ticketsContainer = container.querySelector('#tickets-container');

        // Create a map of epic_id to epic for quick lookup
        const epicMap = {};
        epics.forEach(epic => {
            epicMap[epic.id] = epic;
        });

        // Create a map of sprint_id to sprint for quick lookup
        const sprintMap = {};
        sprints.forEach(sprint => {
            sprintMap[sprint.id] = sprint;
        });

        if (tickets.length === 0) {
            ticketsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">🎫</div>
                    <h2>No tickets yet</h2>
                    <p>Create your first ticket to get started</p>
                    <a href="#/tickets/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create Ticket
                    </a>
                </div>
            `;
            return;
        }

        ticketsContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>Title</th>
                                <th>Status</th>
                                <th>Epic</th>
                                <th>Sprint</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${tickets.map(ticket => {
                                const epic = ticket.epic_id ? epicMap[ticket.epic_id] : null;
                                const sprint = ticket.sprint_id ? sprintMap[ticket.sprint_id] : null;
                                return `
                                <tr ${ticket.status === 'new' ? 'style="background-color: var(--primary-light, #e3f2fd); font-weight: 500;"' : ''}>
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
                                        ${epic ? `<a href="#/epics/${epic.id}" style="color: var(--primary); text-decoration: none;">${escapeHtml(epic.name)}</a>` : 'None'}
                                    </td>
                                    <td>
                                        ${sprint ? `<a href="#/sprints/${sprint.id}" style="color: var(--primary); text-decoration: none;">${escapeHtml(sprint.name)}</a>` : 'None'}
                                    </td>
                                    <td>
                                        <div class="actions">
                                            <button class="btn btn-primary action-btn promote-btn"
                                                    data-id="${ticket.id}"
                                                    title="Promote to top">
                                                ↑
                                            </button>
                                            <a href="#/tickets/${ticket.id}" class="btn btn-secondary action-btn">
                                                View
                                            </a>
                                            <a href="#/tickets/${ticket.id}/edit" class="btn btn-secondary action-btn">
                                                Edit
                                            </a>
                                            <button class="btn btn-danger action-btn delete-btn"
                                                    data-id="${ticket.id}"
                                                    ${auth.isAdmin() || ticket.user_id === auth.getUser().id ? '' : 'disabled'}>
                                                Delete
                                            </button>
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

        const deleteButtons = ticketsContainer.querySelectorAll('.delete-btn');
        deleteButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                if (e.target.disabled) return;
                const id = e.target.dataset.id;
                if (confirm('Are you sure you want to delete this ticket?')) {
                    try {
                        await api.deleteTicket(id);
                        ticketsListView();
                    } catch (error) {
                    }
                }
            });
        });

        const promoteButtons = ticketsContainer.querySelectorAll('.promote-btn');
        promoteButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.target.dataset.id;
                try {
                    await api.promoteTicket(id);
                    ticketsListView();
                } catch (error) {
                }
            });
        });
    } catch (error) {
        const ticketsContainer = container.querySelector('#tickets-container');
        ticketsContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load tickets: ${error.message}</p>
            </div>
        `;
    }
}

export async function ticketDetailView(params) {
    const container = document.getElementById('view-container');
    const [id] = params;

    if (!id) {
        router.navigate('/tickets');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading ticket...</div>
        </div>
    `;

    try {
        const ticket = await api.getTicket(id);

        // Fetch epic if ticket is assigned to one
        let epic = null;
        if (ticket.epic_id) {
            try {
                epic = await api.getEpic(ticket.epic_id);
            } catch (error) {
                // Epic might have been deleted, continue without it
            }
        }

        // Fetch sprint if ticket is assigned to one
        let sprint = null;
        if (ticket.sprint_id) {
            try {
                sprint = await api.getSprint(ticket.sprint_id);
            } catch (error) {
                // Sprint might have been deleted, continue without it
            }
        }

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(ticket.title)}</h1>
                    <div class="actions">
                        <a href="#/tickets/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn" ${auth.isAdmin() || ticket.user_id === auth.getUser().id ? '' : 'disabled'}>Delete</button>
                    </div>
                </div>
                <div class="card">
                    <div style="display: grid; gap: 1rem;">
                        <div>
                            <label class="form-label">Status</label>
                            <div>
                                <span class="badge ${getStatusBadgeClass(ticket.status)}">
                                    ${escapeHtml(ticket.status)}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="form-label">Epic</label>
                            <p>${epic ? `<a href="#/epics/${epic.id}" style="color: var(--primary); text-decoration: none;">${escapeHtml(epic.name)}</a>` : 'None'}</p>
                        </div>
                        <div>
                            <label class="form-label">Sprint</label>
                            <p>${sprint ? `<a href="#/sprints/${sprint.id}" style="color: var(--primary); text-decoration: none;">${escapeHtml(sprint.name)}</a>` : 'None'}</p>
                        </div>
                        <div>
                            <label class="form-label">Description</label>
                            <p>${escapeHtml(ticket.description) || 'No description'}</p>
                        </div>
                        ${ticket.assigned_to ? `
                            <div>
                                <label class="form-label">Assigned To</label>
                                <p>${ticket.assigned_to}</p>
                            </div>
                        ` : ''}
                        <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border);">
                            <div>
                                <label class="form-label">Created</label>
                                <p>${formatDate(ticket.created_at)}</p>
                            </div>
                            <div>
                                <label class="form-label">Last Updated</label>
                                <p>${formatDate(ticket.updated_at)}</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        deleteBtn.addEventListener('click', async () => {
            if (deleteBtn.disabled) return;
            if (confirm('Are you sure you want to delete this ticket?')) {
                try {
                    await api.deleteTicket(id);
                    router.navigate('/tickets');
                } catch (error) {
                }
            }
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load ticket: ${error.message}</p>
                <a href="#/tickets" class="btn btn-secondary" style="margin-top: 1rem;">Back to Tickets</a>
            </div>
        `;
    }
}

export async function ticketFormView(params) {
    const container = document.getElementById('view-container');
    const [id, action] = params;
    const isEdit = action === 'edit';

    let ticket = null;
    let epics = [];
    let sprints = [];
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading ticket...</div>';
        try {
            ticket = await api.getTicket(id);
            epics = await api.getEpics();
            sprints = await api.getSprints();
        } catch (error) {
            router.navigate('/tickets');
            return;
        }
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Ticket</h1>
            </div>
            <div class="card">
                <form id="ticket-form">
                    <div class="form-group">
                        <label class="form-label" for="title">Title *</label>
                        <input
                            type="text"
                            id="title"
                            class="form-input"
                            required
                            value="${ticket ? escapeHtml(ticket.title) : ''}"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="description">Description</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                        >${ticket ? escapeHtml(ticket.description || '') : ''}</textarea>
                    </div>
                    ${isEdit ? `
                    <div class="form-group">
                        <label class="form-label" for="status">Status *</label>
                        <select id="status" class="form-select" required>
                            <option value="open" ${ticket && ticket.status === 'open' ? 'selected' : ''}>Open</option>
                            <option value="in-progress" ${ticket && ticket.status === 'in-progress' ? 'selected' : ''}>In Progress</option>
                            <option value="blocked" ${ticket && ticket.status === 'blocked' ? 'selected' : ''}>Blocked</option>
                            <option value="needs-review" ${ticket && ticket.status === 'needs-review' ? 'selected' : ''}>Needs Review</option>
                            <option value="closed" ${ticket && ticket.status === 'closed' ? 'selected' : ''}>Closed</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="epic">Epic</label>
                        <select id="epic" class="form-select">
                            <option value="">None</option>
                            ${epics.map(epic => `
                                <option value="${epic.id}" ${ticket && ticket.epic_id === epic.id ? 'selected' : ''}>
                                    ${escapeHtml(epic.name)}
                                </option>
                            `).join('')}
                        </select>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="sprint">Sprint</label>
                        <select id="sprint" class="form-select">
                            <option value="">None</option>
                            ${sprints.map(sprint => `
                                <option value="${sprint.id}" ${ticket && ticket.sprint_id === sprint.id ? 'selected' : ''}>
                                    ${escapeHtml(sprint.name)}
                                </option>
                            `).join('')}
                        </select>
                    </div>
                    ` : ''}
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Ticket
                        </button>
                        <a href="#/tickets" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </div>
    `;

    const form = container.querySelector('#ticket-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const title = form.title.value.trim();
        const description = form.description.value.trim();

        const data = { title, description };

        // Only include status, epic, and sprint when editing
        if (isEdit) {
            data.status = form.status.value;
            const epicValue = form.epic.value;
            if (epicValue) {
                data.epic_id = epicValue;
            } else {
                data.epic_id = null;
            }
            const sprintValue = form.sprint.value;
            if (sprintValue) {
                data.sprint_id = sprintValue;
            } else {
                data.sprint_id = null;
            }
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateTicket(id, data);
                router.navigate('/tickets');
            } else {
                await api.createTicket(data);
                router.navigate('/tickets');
            }
        } catch (error) {
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} Ticket`;
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
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
}

function getStatusBadgeClass(status) {
    switch (status) {
        case 'new': return 'badge-primary';
        case 'open': return 'badge-primary';
        case 'in-progress': return 'badge-warning';
        case 'blocked': return 'badge-danger';
        case 'needs-review': return 'badge-warning';
        case 'closed': return 'badge-success';
        default: return 'badge-primary';
    }
}

