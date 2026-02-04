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
        const projects = await api.getProjects();
        const sprints = await api.getSprints();
        const users = await api.getUsersForAssignment();
        const ticketsContainer = container.querySelector('#tickets-container');

        // Create a map of project_id to project for quick lookup
        const projectMap = {};
        projects.forEach(project => {
            projectMap[project.id] = project;
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
                                <th>Project</th>
                                <th>Sprint</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody id="tickets-tbody">
                            ${tickets.map(ticket => {
                                const assignee = ticket.assigned_to ? userMap[ticket.assigned_to] : null;
                                const project = ticket.project_id ? projectMap[ticket.project_id] : null;
                                const sprint = ticket.sprint_id ? sprintMap[ticket.sprint_id] : null;
                                return `
                                <tr class="draggable-row"
                                    data-ticket-id="${ticket.id}"
                                    data-priority="${ticket.priority}"
                                    draggable="true"
                                    style="cursor: move;">
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
                                    <td data-label="Project">
                                        ${project ? `<span title="${escapeHtml(project.name)}">${getProjectAcronym(project.name, projects, project.id)}</span>` : '-'}
                                    </td>
                                    <td data-label="Sprint">
                                        ${sprint ? `<a href="/sprints/${sprint.id}/board" style="color: var(--primary); text-decoration: none;">${escapeHtml(sprint.name)}</a>` : '-'}
                                    </td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <button class="btn btn-secondary action-btn promote-top-btn"
                                                    data-id="${ticket.id}"
                                                    title="Promote to top">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/Arrows/arrow-up-circle-fill.svg" alt="Promote to top" style="width: 20px; height: 20px; display: block;">
                                            </button>
                                            <a href="/tickets/${ticket.id}" class="btn btn-secondary action-btn" title="View details">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/System/eye-fill.svg" alt="View" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                            <a href="/tickets/${ticket.id}/edit" class="btn btn-secondary action-btn" title="Edit ticket">
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

        // Implement drag-and-drop with fractional indexing
        setupDragAndDrop(ticketsContainer, tickets);

        // Handle promote to top button
        const promoteTopButtons = ticketsContainer.querySelectorAll('.promote-top-btn');
        promoteTopButtons.forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const id = e.currentTarget.dataset.id;
                try {
                    await api.promoteTicket(id);
                    ticketsListView();
                } catch (error) {
                    console.error('Failed to promote ticket:', error);
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

        // Fetch project if ticket is assigned to one
        let project = null;
        if (ticket.project_id) {
            try {
                project = await api.getProject(ticket.project_id);
            } catch (error) {
                // Project might have been deleted, continue without it
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
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
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
                    <p style="white-space: pre-wrap; line-height: 1.6; color: var(--text-primary);">${escapeHtml(ticket.description) || '<span style="color: var(--text-secondary); font-style: italic;">No description provided</span>'}</p>
                </div>

                <!-- Acceptance Criteria Card -->
                <div class="card">
                    <h2 class="card-header">Acceptance Criteria</h2>
                    ${ticket.acceptance_criteria && ticket.acceptance_criteria.length > 0 ? `
                        <ul style="margin: 0; padding: 0; list-style: none; line-height: 1.8;">
                            ${ticket.acceptance_criteria.map(criterion => `
                                <li style="margin-bottom: 0.75rem; display: flex; align-items: flex-start; gap: 0.75rem;">
                                    <input
                                        type="checkbox"
                                        class="criteria-checkbox"
                                        data-criteria-id="${criterion.id}"
                                        data-ticket-id="${ticket.id}"
                                        ${criterion.completed ? 'checked' : ''}
                                        style="margin-top: 0.25rem; cursor: pointer; flex-shrink: 0;"
                                    >
                                    <span style="flex: 1; ${criterion.completed ? 'text-decoration: line-through; color: var(--text-secondary);' : 'color: var(--text-primary);'}">${escapeHtml(criterion.content)}</span>
                                </li>
                            `).join('')}
                        </ul>
                    ` : '<p style="color: var(--text-secondary); font-style: italic;">No acceptance criteria provided</p>'}
                </div>

                <!-- Project Details Card -->
                <div class="card">
                    <h2 class="card-header">Project Details</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Project</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem;">
                                ${project ? `<a href="/projects/${project.id}" style="color: var(--primary); text-decoration: none; font-weight: 500;">${escapeHtml(project.name)}</a>` : '<span style="color: var(--text-secondary);">None</span>'}
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

                ${auth.isAdmin() ? `
                <!-- Danger Zone Card -->
                <div class="card">
                    <h2 class="card-header">Danger Zone</h2>
                    <p style="color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.6;">
                        These actions are irreversible and will permanently delete data from the system.
                    </p>
                    <div>
                        <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Delete Ticket</h3>
                        <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                            Permanently delete this ticket. This action cannot be undone.
                        </p>
                        <button type="button" class="btn btn-danger" id="delete-btn">
                            Delete Ticket
                        </button>
                    </div>
                </div>
                ` : ''}
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', async () => {
                if (confirm('Are you sure you want to delete this ticket?')) {
                    try {
                        await api.deleteTicket(id);
                        router.navigate('/tickets');
                    } catch (error) {
                    }
                }
            });
        }

        // Handle acceptance criteria checkbox toggling
        const criteriaCheckboxes = container.querySelectorAll('.criteria-checkbox');
        criteriaCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', async (e) => {
                const criteriaId = e.target.dataset.criteriaId;
                const ticketId = e.target.dataset.ticketId;
                const completed = e.target.checked;

                try {
                    await api.updateAcceptanceCriteriaCompleted(ticketId, criteriaId, completed);

                    // Update the text styling
                    const span = e.target.nextElementSibling;
                    if (completed) {
                        span.style.textDecoration = 'line-through';
                        span.style.color = 'var(--text-secondary)';
                    } else {
                        span.style.textDecoration = 'none';
                        span.style.color = 'var(--text-primary)';
                    }
                } catch (error) {
                    // Revert checkbox on error
                    e.target.checked = !completed;
                    console.error('Failed to update acceptance criteria:', error);
                }
            });
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
    let projects = [];
    let sprints = [];
    let users = [];

    container.innerHTML = '<div class="loading">Loading...</div>';

    try {
        // Fetch users for assignee dropdown
        users = await api.getUsersForAssignment();

        if (isEdit && id) {
            ticket = await api.getTicket(id);
            projects = await api.getProjects();
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

                <!-- Acceptance Criteria Card -->
                <div class="card">
                    <h2 class="card-header">Acceptance Criteria *</h2>
                    <div id="acceptance-criteria-container">
                        <!-- Criteria fields will be added here dynamically -->
                    </div>
                    <button type="button" id="add-criteria-btn" class="btn btn-secondary" style="margin-top: 1rem;">
                        Add AC
                    </button>
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
                            <label class="form-label" for="project">Project</label>
                            <select id="project" class="form-select">
                                <option value="">None</option>
                                ${projects.map(project => `
                                    <option value="${project.id}" ${ticket && ticket.project_id === project.id ? 'selected' : ''}>
                                        ${escapeHtml(project.name)}
                                    </option>
                                `).join('')}
                            </select>
                        </div>
                        <div class="form-group" style="margin-bottom: 0;">
                            <label class="form-label" for="sprint">Sprint</label>
                            <select id="sprint" class="form-select" ${ticket && !ticket.size ? 'disabled' : ''}>
                                <option value="">None</option>
                                ${sprints.map(sprint => `
                                    <option value="${sprint.id}" ${ticket && ticket.sprint_id === sprint.id ? 'selected' : ''}>
                                        ${escapeHtml(sprint.name)}
                                    </option>
                                `).join('')}
                            </select>
                            <small style="color: var(--text-secondary); margin-top: 0.25rem; display: block;">Tickets must be sized before adding to a sprint</small>
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
                        <button type="button" class="btn btn-secondary" onclick="history.back()">Cancel</button>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#ticket-form');

    // Setup acceptance criteria
    const criteriaContainer = form.querySelector('#acceptance-criteria-container');
    const addCriteriaBtn = form.querySelector('#add-criteria-btn');
    let criteriaCount = 0;

    function createCriteriaField(criterionData = {}, showCompleted = false) {
        const content = criterionData.content || '';
        const completed = criterionData.completed || false;

        const fieldDiv = document.createElement('div');
        fieldDiv.className = 'form-group';
        fieldDiv.style.marginBottom = '0.5rem';
        fieldDiv.innerHTML = `
            <div style="display: flex; gap: 0.5rem; align-items: center;">
                ${showCompleted ? `
                <input
                    type="checkbox"
                    class="criteria-completed-checkbox"
                    ${completed ? 'checked' : ''}
                    style="cursor: pointer; flex-shrink: 0;"
                >
                ` : ''}
                <input
                    type="text"
                    class="form-input acceptance-criteria-input"
                    placeholder="Enter acceptance criterion..."
                    maxlength="256"
                    value="${escapeHtml(content)}"
                    style="flex: 1; ${completed ? 'text-decoration: line-through; color: var(--text-secondary);' : ''}"
                >
                <button type="button" class="btn btn-danger remove-criteria-btn" style="flex-shrink: 0; padding: 0.375rem 0.75rem; font-size: 0.875rem;">
                    Remove
                </button>
            </div>
        `;

        const input = fieldDiv.querySelector('.acceptance-criteria-input');
        const removeBtn = fieldDiv.querySelector('.remove-criteria-btn');
        const checkbox = fieldDiv.querySelector('.criteria-completed-checkbox');

        // Handle checkbox toggle for strikethrough
        if (checkbox) {
            checkbox.addEventListener('change', () => {
                if (checkbox.checked) {
                    input.style.textDecoration = 'line-through';
                    input.style.color = 'var(--text-secondary)';
                } else {
                    input.style.textDecoration = 'none';
                    input.style.color = '';
                }
            });
        }

        // Remove button handler
        removeBtn.addEventListener('click', () => {
            if (criteriaContainer.querySelectorAll('.form-group').length > 1) {
                fieldDiv.remove();
                criteriaCount--;
                updateAddButtonVisibility();
            }
        });

        criteriaCount++;
        return fieldDiv;
    }

    function updateAddButtonVisibility() {
        addCriteriaBtn.style.display = criteriaCount >= 6 ? 'none' : 'block';
    }

    // Initialize with existing criteria or one empty field
    if (isEdit && ticket && ticket.acceptance_criteria && ticket.acceptance_criteria.length > 0) {
        ticket.acceptance_criteria.forEach(criterion => {
            criteriaContainer.appendChild(createCriteriaField(criterion, isEdit));
        });
    } else {
        criteriaContainer.appendChild(createCriteriaField({}, isEdit));
    }
    updateAddButtonVisibility();

    // Add criteria button handler
    addCriteriaBtn.addEventListener('click', () => {
        if (criteriaCount < 6) {
            criteriaContainer.appendChild(createCriteriaField({}, isEdit));
            updateAddButtonVisibility();
        }
    });

    // Enable/disable sprint dropdown based on size selection
    if (isEdit) {
        const sizeSelect = form.querySelector('#size');
        const sprintSelect = form.querySelector('#sprint');

        sizeSelect.addEventListener('change', () => {
            const hasSize = sizeSelect.value !== '';
            sprintSelect.disabled = !hasSize;

            // Clear sprint selection if size is removed
            if (!hasSize) {
                sprintSelect.value = '';
            }
        });
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const title = form.title.value.trim();
        const description = form.description.value.trim();

        // Collect acceptance criteria
        const criteriaFields = form.querySelectorAll('.form-group');
        const acceptanceCriteria = Array.from(criteriaFields)
            .map(field => {
                const textarea = field.querySelector('.acceptance-criteria-input');
                const checkbox = field.querySelector('.criteria-completed-checkbox');
                const content = textarea ? textarea.value.trim() : '';
                const completed = checkbox ? checkbox.checked : false;
                return { content, completed };
            })
            .filter(criterion => criterion.content.length > 0);

        // Validate acceptance criteria
        if (acceptanceCriteria.length < 1) {
            alert('At least one acceptance criterion is required');
            return;
        }
        if (acceptanceCriteria.length > 6) {
            alert('Maximum 6 acceptance criteria allowed');
            return;
        }

        // Validate: if sprint is selected, size must be set
        if (isEdit) {
            const sprintValue = form.sprint.value;
            const sizeValue = form.size.value;
            if (sprintValue && !sizeValue) {
                alert('Tickets must be sized before adding to a sprint');
                return;
            }
        }

        const data = { title, description, acceptance_criteria: acceptanceCriteria };

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

        // Only include status, project, and sprint when editing
        if (isEdit) {
            data.status = form.status.value;
            const projectValue = form.project.value;
            if (projectValue) {
                data.project_id = projectValue;
            } else {
                data.project_id = null;
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

function getProjectAcronym(projectName, allProjects, currentProjectId) {
    // Remove all whitespace for acronym generation
    const baseName = projectName.replace(/\s+/g, '');
    let length = 6;
    let acronym;

    // Start with 6 characters and extend if there's a collision
    while (length <= baseName.length) {
        acronym = baseName.substring(0, length).toUpperCase();

        // Check if any other project has this same acronym
        const hasCollision = allProjects.some(p => {
            if (p.id === currentProjectId) return false; // Don't compare with self
            const otherBaseName = p.name.replace(/\s+/g, '');
            const otherAcronym = otherBaseName.substring(0, length).toUpperCase();
            return acronym === otherAcronym;
        });

        if (!hasCollision) {
            return acronym;
        }

        length++;
    }

    // If we've used all characters and still have collision, return full name uppercase
    return baseName.toUpperCase();
}

function setupDragAndDrop(container, tickets) {
    let draggedElement = null;
    let draggedTicketId = null;

    const tbody = container.querySelector('#tickets-tbody');
    const rows = tbody.querySelectorAll('.draggable-row');

    rows.forEach(row => {
        // Prevent links and buttons from being draggable
        row.querySelectorAll('a, button').forEach(el => {
            el.setAttribute('draggable', 'false');
        });

        // Dragstart event
        row.addEventListener('dragstart', (e) => {
            draggedElement = row;
            draggedTicketId = row.dataset.ticketId;
            row.style.opacity = '0.5';
            e.dataTransfer.effectAllowed = 'move';
            e.dataTransfer.setData('text/plain', row.dataset.ticketId);
        });

        // Dragend event
        row.addEventListener('dragend', (e) => {
            row.style.opacity = '1';
            // Remove all drag-over classes
            rows.forEach(r => r.classList.remove('drag-over'));
        });

        // Dragover event (allow drop)
        row.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';

            if (draggedElement !== row) {
                row.classList.add('drag-over');
            }
            return false;
        });

        // Dragleave event
        row.addEventListener('dragleave', (e) => {
            row.classList.remove('drag-over');
        });

        // Drop event
        row.addEventListener('drop', async (e) => {
            e.preventDefault();
            e.stopPropagation();

            if (draggedElement !== row) {
                // Calculate new priority using fractional indexing
                const allRows = Array.from(tbody.querySelectorAll('.draggable-row'));
                const targetIndex = allRows.indexOf(row);
                const draggedIndex = allRows.indexOf(draggedElement);

                let newPriority;

                if (draggedIndex < targetIndex) {
                    // Moving down - insert after the target row
                    const nextRow = allRows[targetIndex + 1];
                    if (nextRow) {
                        // Insert between target and next
                        const targetPriority = parseFloat(row.dataset.priority);
                        const nextPriority = parseFloat(nextRow.dataset.priority);
                        newPriority = (targetPriority + nextPriority) / 2.0;
                    } else {
                        // Insert at bottom
                        const targetPriority = parseFloat(row.dataset.priority);
                        newPriority = targetPriority - 1.0;
                    }
                } else {
                    // Moving up - insert before the target row
                    const prevRow = allRows[targetIndex - 1];
                    if (prevRow) {
                        // Insert between prev and target
                        const prevPriority = parseFloat(prevRow.dataset.priority);
                        const targetPriority = parseFloat(row.dataset.priority);
                        newPriority = (prevPriority + targetPriority) / 2.0;
                    } else {
                        // Insert at top
                        const targetPriority = parseFloat(row.dataset.priority);
                        newPriority = targetPriority + 1.0;
                    }
                }

                try {
                    // Update priority via API
                    await api.updateTicketPriority(draggedTicketId, newPriority);
                    // Refresh the view
                    ticketsListView();
                } catch (error) {
                    console.error('Failed to update ticket priority:', error);
                }
            }

            row.classList.remove('drag-over');
            return false;
        });
    });
}

