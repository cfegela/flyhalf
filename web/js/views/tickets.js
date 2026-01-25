import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function ticketsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Tickets</h1>
                <a href="/tickets/new" class="btn btn-primary">Create Ticket</a>
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
        const users = await api.getUsersForAssignment();
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

        // Create a map of user_id to user for quick lookup
        const userMap = {};
        users.forEach(user => {
            userMap[user.id] = user;
        });

        if (tickets.length === 0) {
            ticketsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">🎫</div>
                    <h2>No tickets yet</h2>
                    <p>Create your first ticket to get started</p>
                    <a href="/tickets/new" class="btn btn-primary" style="margin-top: 1rem;">
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
                                <th>Size</th>
                                <th>Assignee</th>
                                <th>Epic</th>
                                <th>Sprint</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${tickets.map(ticket => {
                                const assignee = ticket.assigned_to ? userMap[ticket.assigned_to] : null;
                                const epic = ticket.epic_id ? epicMap[ticket.epic_id] : null;
                                const sprint = ticket.sprint_id ? sprintMap[ticket.sprint_id] : null;
                                return `
                                <tr>
                                    <td data-label="Title">
                                        <strong>${escapeHtml(ticket.title)}</strong>
                                    </td>
                                    <td data-label="Status">
                                        <span class="badge ${getStatusBadgeClass(ticket.status)}">
                                            ${escapeHtml(ticket.status)}
                                        </span>
                                    </td>
                                    <td data-label="Size">
                                        ${getSizeLabel(ticket.size)}
                                    </td>
                                    <td data-label="Assignee">
                                        ${assignee ? `${escapeHtml(assignee.first_name)} ${escapeHtml(assignee.last_name)}` : '-'}
                                    </td>
                                    <td data-label="Epic">
                                        ${epic ? `<a href="/epics/${epic.id}" style="color: var(--primary); text-decoration: none;" title="${escapeHtml(epic.name)}">${getEpicAcronym(epic.name)}</a>` : '-'}
                                    </td>
                                    <td data-label="Sprint">
                                        ${sprint ? `<a href="/sprints/${sprint.id}" style="color: var(--primary); text-decoration: none;">${escapeHtml(sprint.name)}</a>` : '-'}
                                    </td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <button class="btn btn-primary action-btn promote-top-btn"
                                                    data-id="${ticket.id}"
                                                    title="Promote to top">
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="icon">
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 18.75 7.5-7.5 7.5 7.5"/>
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 7.5-7.5 7.5 7.5"/>
                                                </svg>
                                            </button>
                                            <button class="btn btn-primary action-btn promote-up-btn"
                                                    data-id="${ticket.id}"
                                                    title="Promote up one">
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="icon">
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 15.75 7.5-7.5 7.5 7.5"/>
                                                </svg>
                                            </button>
                                            <button class="btn btn-primary action-btn promote-down-btn"
                                                    data-id="${ticket.id}"
                                                    title="Promote down one">
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="icon">
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5"/>
                                                </svg>
                                            </button>
                                            <a href="/tickets/${ticket.id}"
                                               class="btn btn-secondary action-btn"
                                               title="View ticket">
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="icon">
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z"/>
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/>
                                                </svg>
                                            </a>
                                            <a href="/tickets/${ticket.id}/edit"
                                               class="btn btn-secondary action-btn"
                                               title="Edit ticket">
                                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="icon">
                                                    <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125"/>
                                                </svg>
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

        const promoteTopButtons = ticketsContainer.querySelectorAll('.promote-top-btn');
        promoteTopButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.currentTarget.dataset.id;
                try {
                    await api.promoteTicket(id);
                    ticketsListView();
                } catch (error) {
                }
            });
        });

        const promoteUpButtons = ticketsContainer.querySelectorAll('.promote-up-btn');
        promoteUpButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.currentTarget.dataset.id;
                try {
                    await api.promoteTicketUp(id);
                    ticketsListView();
                } catch (error) {
                }
            });
        });

        const promoteDownButtons = ticketsContainer.querySelectorAll('.promote-down-btn');
        promoteDownButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.currentTarget.dataset.id;
                try {
                    await api.demoteTicketDown(id);
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

        // Fetch assignee if ticket is assigned to someone
        let assignee = null;
        if (ticket.assigned_to) {
            try {
                assignee = await api.getUser(ticket.assigned_to);
            } catch (error) {
                // User might have been deleted, continue without it
            }
        }

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
                        <a href="/tickets/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn" ${auth.isAdmin() || ticket.user_id === auth.getUser().id ? '' : 'disabled'}>Delete</button>
                    </div>
                </div>

                <!-- Key Information Card -->
                <div class="card">
                    <h2 class="card-header">Key Information</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Status</label>
                            <div style="margin-top: 0.25rem;">
                                <span class="badge ${getStatusBadgeClass(ticket.status)}" style="font-size: 0.875rem; padding: 0.375rem 0.875rem;">
                                    ${escapeHtml(ticket.status)}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="form-label">Size</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${ticket.size ? getSizeLabel(ticket.size) : 'Not Sized'}
                            </p>
                        </div>
                        <div>
                            <label class="form-label">Assigned To</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${assignee ? `${escapeHtml(assignee.first_name)} ${escapeHtml(assignee.last_name)}` : 'Unassigned'}
                            </p>
                            ${assignee ? `<p style="font-size: 0.875rem; color: var(--text-secondary); margin-top: 0.125rem;">${escapeHtml(assignee.email)}</p>` : ''}
                        </div>
                    </div>
                </div>

                <!-- Description Card -->
                <div class="card">
                    <h2 class="card-header">Description</h2>
                    <p style="white-space: pre-wrap; line-height: 1.6; color: var(--text-primary);">
                        ${escapeHtml(ticket.description) || '<span style="color: var(--text-secondary); font-style: italic;">No description provided</span>'}
                    </p>
                </div>

                <!-- Project Details Card -->
                <div class="card">
                    <h2 class="card-header">Project Details</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Epic</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem;">
                                ${epic ? `<a href="/epics/${epic.id}" style="color: var(--primary); text-decoration: none; font-weight: 500;">${escapeHtml(epic.name)}</a>` : '<span style="color: var(--text-secondary);">None</span>'}
                            </p>
                        </div>
                        <div>
                            <label class="form-label">Sprint</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem;">
                                ${sprint ? `<a href="/sprints/${sprint.id}" style="color: var(--primary); text-decoration: none; font-weight: 500;">${escapeHtml(sprint.name)}</a>` : '<span style="color: var(--text-secondary);">None</span>'}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Metadata Card -->
                <div class="card">
                    <h2 class="card-header">Metadata</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Created</label>
                            <p style="margin-top: 0.25rem; font-size: 0.875rem; color: var(--text-secondary);">
                                ${formatDate(ticket.created_at)}
                            </p>
                        </div>
                        <div>
                            <label class="form-label">Last Updated</label>
                            <p style="margin-top: 0.25rem; font-size: 0.875rem; color: var(--text-secondary);">
                                ${formatDate(ticket.updated_at)}
                            </p>
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
                <a href="/tickets" class="btn btn-secondary" style="margin-top: 1rem;">Back to Tickets</a>
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
    let users = [];

    container.innerHTML = '<div class="loading">Loading...</div>';

    try {
        // Fetch users for assignee dropdown
        users = await api.getUsersForAssignment();

        if (isEdit && id) {
            ticket = await api.getTicket(id);
            epics = await api.getEpics();
            sprints = await api.getSprints();
        }
    } catch (error) {
        router.navigate('/tickets');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Ticket</h1>
            </div>

            <form id="ticket-form">
                <!-- Basic Information Card -->
                <div class="card">
                    <h2 class="card-header">Basic Information</h2>
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
                        <label class="form-label" for="description">Description *</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                            required
                            placeholder="Provide a detailed description of the ticket..."
                        >${ticket ? escapeHtml(ticket.description || '') : ''}</textarea>
                    </div>
                </div>

                <!-- Assignment & Sizing Card -->
                <div class="card">
                    <h2 class="card-header">Assignment & Sizing</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem;">
                        <div class="form-group" style="margin-bottom: 0;">
                            <label class="form-label" for="assigned_to">Assign To</label>
                            <select id="assigned_to" class="form-select">
                                <option value="">Unassigned</option>
                                ${users.map(user => `
                                    <option value="${user.id}" ${ticket && ticket.assigned_to === user.id ? 'selected' : ''}>
                                        ${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)} (${escapeHtml(user.email)})
                                    </option>
                                `).join('')}
                            </select>
                        </div>
                        <div class="form-group" style="margin-bottom: 0;">
                            <label class="form-label" for="size">Size</label>
                            <select id="size" class="form-select">
                                <option value="">Not Sized</option>
                                <option value="1" ${ticket && ticket.size === 1 ? 'selected' : ''}>Small</option>
                                <option value="2" ${ticket && ticket.size === 2 ? 'selected' : ''}>Medium</option>
                                <option value="3" ${ticket && ticket.size === 3 ? 'selected' : ''}>Large</option>
                                <option value="5" ${ticket && ticket.size === 5 ? 'selected' : ''}>X-Large</option>
                                <option value="8" ${ticket && ticket.size === 8 ? 'selected' : ''}>Danger</option>
                            </select>
                        </div>
                    </div>
                </div>

                ${isEdit ? `
                <!-- Project Organization Card -->
                <div class="card">
                    <h2 class="card-header">Project Organization</h2>
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
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem;">
                        <div class="form-group" style="margin-bottom: 0;">
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
                        <div class="form-group" style="margin-bottom: 0;">
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
                    </div>
                </div>
                ` : ''}

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Ticket
                        </button>
                        <a href="${isEdit ? `/tickets/${id}` : '/tickets'}" class="btn btn-secondary">Cancel</a>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#ticket-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const title = form.title.value.trim();
        const description = form.description.value.trim();

        const data = { title, description };

        // Handle assignee (available for both create and edit)
        const assignedToValue = form.assigned_to.value;
        if (assignedToValue) {
            data.assigned_to = assignedToValue;
        } else {
            data.assigned_to = null;
        }

        // Handle size (available for both create and edit)
        const sizeValue = form.size.value;
        if (sizeValue) {
            data.size = parseInt(sizeValue, 10);
        } else {
            data.size = null;
        }

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
        case 'open': return 'badge-open';
        case 'in-progress': return 'badge-in-progress';
        case 'blocked': return 'badge-blocked';
        case 'needs-review': return 'badge-needs-review';
        case 'closed': return 'badge-closed';
        default: return 'badge-open';
    }
}

function getSizeLabel(size) {
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

function getEpicAcronym(epicName) {
    // Remove spaces and lowercase letters, keeping only uppercase letters
    return epicName.replace(/[a-z\s]/g, '');
}

