import { api } from '../api.js';
import { router } from '../router.js';
import { toast } from '../components/toast.js';
import { auth } from '../auth.js';

export async function resourcesListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Resources</h1>
                <a href="#/resources/new" class="btn btn-primary">Create Resource</a>
            </div>
            <div id="resources-container">
                <div class="loading">Loading resources...</div>
            </div>
        </div>
    `;

    try {
        const resources = await api.getResources();
        const resourcesContainer = container.querySelector('#resources-container');

        if (resources.length === 0) {
            resourcesContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📦</div>
                    <h2>No resources yet</h2>
                    <p>Create your first resource to get started</p>
                    <a href="#/resources/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create Resource
                    </a>
                </div>
            `;
            return;
        }

        resourcesContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>Title</th>
                                <th>Status</th>
                                <th>Created</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${resources.map(resource => `
                                <tr>
                                    <td>
                                        <strong>${escapeHtml(resource.title)}</strong>
                                        ${resource.description ? `<br><small style="color: var(--text-secondary);">${escapeHtml(resource.description.substring(0, 60))}${resource.description.length > 60 ? '...' : ''}</small>` : ''}
                                    </td>
                                    <td>
                                        <span class="badge ${getStatusBadgeClass(resource.status)}">
                                            ${escapeHtml(resource.status)}
                                        </span>
                                    </td>
                                    <td>${formatDate(resource.created_at)}</td>
                                    <td>
                                        <div class="actions">
                                            <a href="#/resources/${resource.id}" class="btn btn-secondary action-btn">
                                                View
                                            </a>
                                            <a href="#/resources/${resource.id}/edit" class="btn btn-secondary action-btn">
                                                Edit
                                            </a>
                                            <button class="btn btn-danger action-btn delete-btn" data-id="${resource.id}">
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

        const deleteButtons = resourcesContainer.querySelectorAll('.delete-btn');
        deleteButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.target.dataset.id;
                if (confirm('Are you sure you want to delete this resource?')) {
                    try {
                        await api.deleteResource(id);
                        toast.success('Resource deleted successfully');
                        resourcesListView();
                    } catch (error) {
                        toast.error('Failed to delete resource: ' + error.message);
                    }
                }
            });
        });
    } catch (error) {
        const resourcesContainer = container.querySelector('#resources-container');
        resourcesContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load resources: ${error.message}</p>
            </div>
        `;
    }
}

export async function resourceDetailView(params) {
    const container = document.getElementById('view-container');
    const [id] = params;

    if (!id) {
        router.navigate('/resources');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading resource...</div>
        </div>
    `;

    try {
        const resource = await api.getResource(id);

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(resource.title)}</h1>
                    <div class="actions">
                        <a href="#/resources/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn">Delete</button>
                    </div>
                </div>
                <div class="card">
                    <div style="display: grid; gap: 1rem;">
                        <div>
                            <label class="form-label">Status</label>
                            <div>
                                <span class="badge ${getStatusBadgeClass(resource.status)}">
                                    ${escapeHtml(resource.status)}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="form-label">Description</label>
                            <p>${escapeHtml(resource.description) || 'No description'}</p>
                        </div>
                        ${resource.metadata ? `
                            <div>
                                <label class="form-label">Metadata</label>
                                <pre style="background: var(--bg-gray); padding: 1rem; border-radius: 0.375rem; overflow-x: auto;">${JSON.stringify(resource.metadata, null, 2)}</pre>
                            </div>
                        ` : ''}
                        <div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border);">
                            <div>
                                <label class="form-label">Created</label>
                                <p>${formatDate(resource.created_at)}</p>
                            </div>
                            <div>
                                <label class="form-label">Last Updated</label>
                                <p>${formatDate(resource.updated_at)}</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        deleteBtn.addEventListener('click', async () => {
            if (confirm('Are you sure you want to delete this resource?')) {
                try {
                    await api.deleteResource(id);
                    toast.success('Resource deleted successfully');
                    router.navigate('/resources');
                } catch (error) {
                    toast.error('Failed to delete resource: ' + error.message);
                }
            }
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load resource: ${error.message}</p>
                <a href="#/resources" class="btn btn-secondary" style="margin-top: 1rem;">Back to Resources</a>
            </div>
        `;
    }
}

export async function resourceFormView(params) {
    const container = document.getElementById('view-container');
    const [id, action] = params;
    const isEdit = action === 'edit';

    let resource = null;
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading resource...</div>';
        try {
            resource = await api.getResource(id);
        } catch (error) {
            toast.error('Failed to load resource: ' + error.message);
            router.navigate('/resources');
            return;
        }
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Resource</h1>
            </div>
            <div class="card">
                <form id="resource-form">
                    <div class="form-group">
                        <label class="form-label" for="title">Title *</label>
                        <input
                            type="text"
                            id="title"
                            class="form-input"
                            required
                            value="${resource ? escapeHtml(resource.title) : ''}"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="description">Description</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                        >${resource ? escapeHtml(resource.description || '') : ''}</textarea>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="status">Status *</label>
                        <select id="status" class="form-select" required>
                            <option value="active" ${resource && resource.status === 'active' ? 'selected' : ''}>Active</option>
                            <option value="inactive" ${resource && resource.status === 'inactive' ? 'selected' : ''}>Inactive</option>
                            <option value="archived" ${resource && resource.status === 'archived' ? 'selected' : ''}>Archived</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="metadata">Metadata (JSON)</label>
                        <textarea
                            id="metadata"
                            class="form-textarea"
                            placeholder='{"key": "value"}'
                        >${resource && resource.metadata ? JSON.stringify(resource.metadata, null, 2) : ''}</textarea>
                    </div>
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Resource
                        </button>
                        <a href="#/resources" class="btn btn-secondary">Cancel</a>
                    </div>
                </form>
            </div>
        </div>
    `;

    const form = container.querySelector('#resource-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const title = form.title.value.trim();
        const description = form.description.value.trim();
        const status = form.status.value;
        const metadataStr = form.metadata.value.trim();

        let metadata = null;
        if (metadataStr) {
            try {
                metadata = JSON.parse(metadataStr);
            } catch (error) {
                toast.error('Invalid JSON in metadata field');
                return;
            }
        }

        const data = { title, description, status, metadata };

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateResource(id, data);
                toast.success('Resource updated successfully');
                router.navigate(`/resources/${id}`);
            } else {
                const newResource = await api.createResource(data);
                toast.success('Resource created successfully');
                router.navigate(`/resources/${newResource.id}`);
            }
        } catch (error) {
            toast.error(`Failed to ${isEdit ? 'update' : 'create'} resource: ` + error.message);
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} Resource`;
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
        case 'active': return 'badge-success';
        case 'inactive': return 'badge-warning';
        case 'archived': return 'badge-danger';
        default: return 'badge-primary';
    }
}
