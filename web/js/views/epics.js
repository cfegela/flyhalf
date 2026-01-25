import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function epicsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Epics</h1>
                <a href="/epics/new" class="btn btn-primary">Create Epic</a>
            </div>
            <div id="epics-container">
                <div class="loading">Loading epics...</div>
            </div>
        </div>
    `;

    try {
        const epics = await api.getEpics();
        const epicsContainer = container.querySelector('#epics-container');

        if (epics.length === 0) {
            epicsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📚</div>
                    <h2>No epics yet</h2>
                    <p>Create your first epic to get started</p>
                    <a href="/epics/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create Epic
                    </a>
                </div>
            `;
            return;
        }

        epicsContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${epics.map(epic => `
                                <tr class="clickable-row" data-epic-id="${epic.id}" style="cursor: pointer;">
                                    <td data-label="Name">
                                        <strong>${escapeHtml(epic.name)} (${getEpicAcronym(epic.name)})</strong>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;

        // Make rows clickable to navigate to epic details
        const clickableRows = epicsContainer.querySelectorAll('.clickable-row');
        clickableRows.forEach(row => {
            row.addEventListener('click', (e) => {
                const epicId = row.dataset.epicId;
                router.navigate(`/epics/${epicId}`);
            });
        });

    } catch (error) {
        const epicsContainer = container.querySelector('#epics-container');
        epicsContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load epics: ${error.message}</p>
            </div>
        `;
    }
}

export async function epicDetailView(params) {
    const container = document.getElementById('view-container');
    const [id] = params;

    if (!id) {
        router.navigate('/epics');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading epic...</div>
        </div>
    `;

    try {
        const epic = await api.getEpic(id);
        const allTickets = await api.getTickets();

        // Filter tickets for this epic
        const epicTickets = allTickets.filter(ticket => ticket.epic_id === id);

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(epic.name)}</h1>
                    <div class="actions">
                        <a href="/epics/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn" ${auth.isAdmin() || epic.user_id === auth.getUser().id ? '' : 'disabled'}>Delete</button>
                    </div>
                </div>

                <!-- Epic Information Card -->
                <div class="card">
                    <h2 class="card-header">Epic Details</h2>
                    <div>
                        <label class="form-label">Description</label>
                        <p style="white-space: pre-wrap; line-height: 1.6; color: var(--text-primary); margin-top: 0.25rem;">
                            ${escapeHtml(epic.description) || '<span style="color: var(--text-secondary); font-style: italic;">No description provided</span>'}
                        </p>
                    </div>
                </div>

                <!-- Tickets Card -->
                <div class="card">
                    <h2 class="card-header">Tickets (${epicTickets.length})</h2>
                    ${epicTickets.length === 0 ? `
                        <div class="empty-state">
                            <div class="empty-state-icon">🎫</div>
                            <p>No tickets assigned to this epic</p>
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
                                    ${epicTickets.map(ticket => `
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
            if (confirm('Are you sure you want to delete this epic?')) {
                try {
                    await api.deleteEpic(id);
                    router.navigate('/epics');
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
                <p style="color: var(--danger);">Failed to load epic: ${error.message}</p>
                <a href="/epics" class="btn btn-secondary" style="margin-top: 1rem;">Back to Epics</a>
            </div>
        `;
    }
}

export async function epicFormView(params) {
    const container = document.getElementById('view-container');
    const [id, action] = params;
    const isEdit = action === 'edit';

    let epic = null;
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading epic...</div>';
        try {
            epic = await api.getEpic(id);
        } catch (error) {
            router.navigate('/epics');
            return;
        }
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Epic</h1>
            </div>

            <form id="epic-form">
                <!-- Epic Information Card -->
                <div class="card">
                    <h2 class="card-header">Epic Information</h2>
                    <div class="form-group">
                        <label class="form-label" for="name">Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
                            placeholder="e.g., User Authentication System"
                            value="${epic ? escapeHtml(epic.name) : ''}"
                        >
                        <small style="color: var(--text-secondary);">Use title case for clarity. Uppercase letters will form the epic acronym.</small>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="description">Description *</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                            required
                            placeholder="Provide a detailed description of the epic's goals and scope..."
                        >${epic ? escapeHtml(epic.description || '') : ''}</textarea>
                    </div>
                </div>

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Epic
                        </button>
                        <a href="${isEdit ? `#/epics/${id}` : '#/epics'}" class="btn btn-secondary">Cancel</a>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#epic-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = form.name.value.trim();
        const description = form.description.value.trim();

        const data = { name, description };

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateEpic(id, data);
                router.navigate('/epics');
            } else {
                await api.createEpic(data);
                router.navigate('/epics');
            }
        } catch (error) {
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} Epic`;
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

function getEpicAcronym(epicName) {
    // Remove spaces and lowercase letters, keeping only uppercase letters
    return epicName.replace(/[a-z\s]/g, '');
}
