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
        const ticketsContainer = container.querySelector('#tickets-container');

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
                                <th>Title</th>
                                <th>Status</th>
                                <th>Priority</th>
                                <th>Created</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${tickets.map(ticket => `
                                <tr>
                                    <td>
                                        <strong>${escapeHtml(ticket.title)}</strong>
                                        ${ticket.description ? `<br><small style="color: var(--text-secondary);">${escapeHtml(ticket.description.substring(0, 60))}${ticket.description.length > 60 ? '...' : ''}</small>` : ''}
                                    </td>
                                    <td>
                                        <span class="badge ${getStatusBadgeClass(ticket.status)}">
                                            ${escapeHtml(ticket.status)}
                                        </span>
                                    </td>
                                    <td>
                                        <span class="badge ${getPriorityBadgeClass(ticket.priority)}">
                                            ${escapeHtml(ticket.priority)}
                                        </span>
                                    </td>
                                    <td>${formatDate(ticket.created_at)}</td>
                                    <td>
                                        <div class="actions">
                                            <a href="#/tickets/${ticket.id}" class="btn btn-secondary action-btn">
                                                View
                                            </a>
                                            <a href="#/tickets/${ticket.id}/edit" class="btn btn-secondary action-btn">
                                                Edit
                                            </a>
                                            <button class="btn btn-danger action-btn delete-btn" data-id="${ticket.id}">
                                                Delete
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;

        const deleteButtons = ticketsContainer.querySelectorAll('.delete-btn');
        deleteButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
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

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(ticket.title)}</h1>
                    <div class="actions">
                        <a href="#/tickets/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn">Delete</button>
                    </div>
                </div>
                <div class="card">
                    <div style="display: grid; gap: 1rem;">
                        <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem;">
                            <div>
                                <label class="form-label">Status</label>
                                <div>
                                    <span class="badge ${getStatusBadgeClass(ticket.status)}">
                                        ${escapeHtml(ticket.status)}
                                    </span>
                                </div>
                            </div>
                            <div>
                                <label class="form-label">Priority</label>
                                <div>
                                    <span class="badge ${getPriorityBadgeClass(ticket.priority)}">
                                        ${escapeHtml(ticket.priority)}
                                    </span>
                                </div>
                            </div>
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
                        ${ticket.metadata ? `
                            <div>
                                <label class="form-label">Metadata</label>
                                <pre style="background: var(--bg-gray); padding: 1rem; border-radius: 0.375rem; overflow-x: auto;">${JSON.stringify(ticket.metadata, null, 2)}</pre>
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
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading ticket...</div>';
        try {
            ticket = await api.getTicket(id);
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
                    <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem;">
                        <div class="form-group">
                            <label class="form-label" for="status">Status *</label>
                            <select id="status" class="form-select" required>
                                <option value="open" ${ticket && ticket.status === 'open' ? 'selected' : ''}>Open</option>
                                <option value="in_progress" ${ticket && ticket.status === 'in_progress' ? 'selected' : ''}>In Progress</option>
                                <option value="resolved" ${ticket && ticket.status === 'resolved' ? 'selected' : ''}>Resolved</option>
                                <option value="closed" ${ticket && ticket.status === 'closed' ? 'selected' : ''}>Closed</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label class="form-label" for="priority">Priority *</label>
                            <select id="priority" class="form-select" required>
                                <option value="low" ${ticket && ticket.priority === 'low' ? 'selected' : ''}>Low</option>
                                <option value="medium" ${ticket && ticket.priority === 'medium' ? 'selected' : ''}>Medium</option>
                                <option value="high" ${ticket && ticket.priority === 'high' ? 'selected' : ''}>High</option>
                                <option value="urgent" ${ticket && ticket.priority === 'urgent' ? 'selected' : ''}>Urgent</option>
                            </select>
                        </div>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="metadata">Metadata (JSON)</label>
                        <textarea
                            id="metadata"
                            class="form-textarea"
                            placeholder='{"key": "value"}'
                        >${ticket && ticket.metadata ? JSON.stringify(ticket.metadata, null, 2) : ''}</textarea>
                    </div>
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
        const status = form.status.value;
        const priority = form.priority.value;
        const metadataStr = form.metadata.value.trim();

        let metadata = null;
        if (metadataStr) {
            try {
                metadata = JSON.parse(metadataStr);
            } catch (error) {
                return;
            }
        }

        const data = { title, description, status, priority, metadata };

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateTicket(id, data);
                router.navigate(`/tickets/${id}`);
            } else {
                const newTicket = await api.createTicket(data);
                router.navigate(`/tickets/${newTicket.id}`);
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
        case 'open': return 'badge-primary';
        case 'in_progress': return 'badge-warning';
        case 'resolved': return 'badge-success';
        case 'closed': return 'badge-danger';
        default: return 'badge-primary';
    }
}

function getPriorityBadgeClass(priority) {
    switch (priority) {
        case 'low': return 'badge-success';
        case 'medium': return 'badge-primary';
        case 'high': return 'badge-warning';
        case 'urgent': return 'badge-danger';
        default: return 'badge-primary';
    }
}
