import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';
import { escapeHtml, formatDate } from '../utils/formatting.js';
import { getStatusBadgeClass, getProjectAcronym } from '../utils/helpers.js';

export async function projectsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Projects</h1>
                <a href="/projects/new" class="btn btn-primary">Create Project</a>
            </div>
            <div id="projects-container">
                <div class="loading">Loading projects...</div>
            </div>
        </div>
    `;

    try {
        const projects = await api.getProjects();
        // Sort projects alphabetically by name
        projects.sort((a, b) => a.name.localeCompare(b.name));
        const projectsContainer = container.querySelector('#projects-container');

        if (projects.length === 0) {
            projectsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📚</div>
                    <h2>No projects yet</h2>
                    <p>Create your first project to get started</p>
                    <a href="/projects/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create Project
                    </a>
                </div>
            `;
            return;
        }

        projectsContainer.innerHTML = `
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
                            ${projects.map(project => `
                                <tr data-project-id="${project.id}">
                                    <td data-label="Name">
                                        <strong>${escapeHtml(project.name)} (${getProjectAcronym(project.name, projects, project.id)})</strong>
                                    </td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <a href="/projects/${project.id}" class="btn btn-secondary action-btn" title="View details">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/System/eye-fill.svg" alt="View" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                            <a href="/projects/${project.id}/edit" class="btn btn-secondary action-btn" title="Edit project">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/Design/pencil-ai-fill.svg" alt="Edit" style="width: 20px; height: 20px; display: block;">
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
        const projectsContainer = container.querySelector('#projects-container');
        projectsContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load projects: ${error.message}</p>
            </div>
        `;
    }
}

export async function projectDetailView(params) {
    const container = document.getElementById('view-container');
    const [id] = params;

    if (!id) {
        router.navigate('/projects');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading project...</div>
        </div>
    `;

    try {
        const project = await api.getProject(id);
        const allTickets = await api.getTickets();

        // Filter tickets for this project
        const projectTickets = allTickets.filter(ticket => ticket.project_id === id);

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(project.name)}</h1>
                    <div class="actions">
                        <a href="/projects/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
                    </div>
                </div>

                <!-- Project Information Card -->
                <div class="card">
                    <h2 class="card-header">Project Details</h2>
                    <div>
                        <label class="form-label">Description</label>
                        <p style="white-space: pre-wrap; line-height: 1.6; color: var(--text-primary); margin-top: 0.25rem;">${escapeHtml(project.description) || '<span style="color: var(--text-secondary); font-style: italic;">No description provided</span>'}</p>
                    </div>
                </div>

                <!-- Tickets Card -->
                <div class="card">
                    <h2 class="card-header">Tickets (${projectTickets.length})</h2>
                    ${projectTickets.length === 0 ? `
                        <div class="empty-state">
                            <div class="empty-state-icon">🎫</div>
                            <p>No tickets assigned to this project</p>
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
                                    ${projectTickets.map(ticket => `
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

                ${auth.isAdmin() ? `
                <!-- Danger Zone Card -->
                <div class="card">
                    <h2 class="card-header">Danger Zone</h2>
                    <p style="color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.6;">
                        These actions are irreversible and will permanently delete data from the system.
                    </p>
                    <div>
                        <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Delete Project</h3>
                        <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                            Permanently delete this project. This action cannot be undone.
                        </p>
                        <button type="button" class="btn btn-danger" id="delete-btn">
                            Delete Project
                        </button>
                    </div>
                </div>
                ` : ''}
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', async () => {
                if (confirm('Are you sure you want to delete this project?')) {
                    try {
                        await api.deleteProject(id);
                        router.navigate('/projects');
                    } catch (error) {
                    }
                }
            });
        }

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
                <p style="color: var(--danger);">Failed to load project: ${error.message}</p>
                <a href="/projects" class="btn btn-secondary" style="margin-top: 1rem;">Back to Projects</a>
            </div>
        `;
    }
}

export async function projectFormView(params) {
    const container = document.getElementById('view-container');
    const [id, action] = params;
    const isEdit = action === 'edit';

    let project = null;
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading project...</div>';
        try {
            project = await api.getProject(id);
        } catch (error) {
            router.navigate('/projects');
            return;
        }
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Project</h1>
            </div>

            <form id="project-form">
                <!-- Project Information Card -->
                <div class="card">
                    <h2 class="card-header">Project Information</h2>
                    <div class="form-group">
                        <label class="form-label" for="name">Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
                            placeholder="e.g., User Authentication System"
                            value="${project ? escapeHtml(project.name) : ''}"
                        >
                        <small style="color: var(--text-secondary);">The first 6 characters (excluding spaces) will be used as the project acronym (shown in uppercase).</small>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="description">Description *</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                            required
                            placeholder="Provide a detailed description of the project's goals and scope..."
                        >${project ? escapeHtml(project.description || '') : ''}</textarea>
                    </div>
                </div>

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Project
                        </button>
                        <button type="button" class="btn btn-secondary" onclick="history.back()">Cancel</button>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#project-form');
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
                await api.updateProject(id, data);
                router.navigate('/projects');
            } else {
                await api.createProject(data);
                router.navigate('/projects');
            }
        } catch (error) {
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} Project`;
        }
    });
}
