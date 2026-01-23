import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function epicsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Epics</h1>
                <a href="#/epics/new" class="btn btn-primary">Create Epic</a>
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
                    <a href="#/epics/new" class="btn btn-primary" style="margin-top: 1rem;">
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
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${epics.map(epic => `
                                <tr>
                                    <td>
                                        <strong>${escapeHtml(epic.name)}</strong>
                                    </td>
                                    <td>
                                        <div class="actions">
                                            <a href="#/epics/${epic.id}" class="btn btn-secondary action-btn">
                                                View
                                            </a>
                                            <a href="#/epics/${epic.id}/edit" class="btn btn-secondary action-btn">
                                                Edit
                                            </a>
                                            <button class="btn btn-danger action-btn delete-btn"
                                                    data-id="${epic.id}"
                                                    ${auth.isAdmin() || epic.user_id === auth.getUser().id ? '' : 'disabled'}>
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

        const deleteButtons = epicsContainer.querySelectorAll('.delete-btn');
        deleteButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                if (e.target.disabled) return;
                const id = e.target.dataset.id;
                if (confirm('Are you sure you want to delete this epic?')) {
                    try {
                        await api.deleteEpic(id);
                        epicsListView();
                    } catch (error) {
                    }
                }
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
                        <a href="#/epics/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn" ${auth.isAdmin() || epic.user_id === auth.getUser().id ? '' : 'disabled'}>Delete</button>
                    </div>
                </div>
                <div class="card">
                    <div style="display: grid; gap: 1rem;">
                        <div>
                            <label class="form-label">Description</label>
                            <p>${escapeHtml(epic.description) || 'No description'}</p>
                        </div>
                    </div>
                </div>
                <div class="card" style="margin-top: 1.5rem;">
                    <h2 style="margin-bottom: 1rem;">Tickets</h2>
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
                                        <th>ID</th>
                                        <th>Title</th>
                                        <th>Status</th>
                                        <th>Actions</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${epicTickets.map(ticket => `
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
            if (confirm('Are you sure you want to delete this epic?')) {
                try {
                    await api.deleteEpic(id);
                    router.navigate('/epics');
                } catch (error) {
                }
            }
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load epic: ${error.message}</p>
                <a href="#/epics" class="btn btn-secondary" style="margin-top: 1rem;">Back to Epics</a>
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
            <div class="card">
                <form id="epic-form">
                    <div class="form-group">
                        <label class="form-label" for="name">Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
                            value="${epic ? escapeHtml(epic.name) : ''}"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="description">Description</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                        >${epic ? escapeHtml(epic.description || '') : ''}</textarea>
                    </div>
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Epic
                        </button>
                        <a href="#/epics" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
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
        case 'new': return 'badge-primary';
        case 'open': return 'badge-primary';
        case 'in-progress': return 'badge-warning';
        case 'blocked': return 'badge-danger';
        case 'needs-review': return 'badge-warning';
        case 'closed': return 'badge-success';
        default: return 'badge-primary';
    }
}
