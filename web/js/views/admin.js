import { api } from '../api.js';
import { router } from '../router.js';

export async function usersListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">User Management</h1>
                <a href="/admin/users/new" class="btn btn-primary">Create User</a>
            </div>
            <div id="users-container">
                <div class="loading">Loading users...</div>
            </div>
        </div>
    `;

    try {
        const users = await api.getUsers();
        const teams = await api.getTeams();
        const usersContainer = container.querySelector('#users-container');

        // Create a map of team IDs to team names for quick lookup
        const teamMap = {};
        teams.forEach(team => {
            teamMap[team.id] = team.name;
        });

        if (users.length === 0) {
            usersContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">👥</div>
                    <h2>No users yet</h2>
                    <p>Create the first user to get started</p>
                </div>
            `;
            return;
        }

        usersContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Email</th>
                                <th>Team</th>
                                <th>Role</th>
                                <th>Status</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${users.map(user => `
                                <tr data-user-id="${user.id}">
                                    <td data-label="Name"><strong>${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</strong></td>
                                    <td data-label="Email">${escapeHtml(user.email)}</td>
                                    <td data-label="Team">${user.team_id && teamMap[user.team_id] ? escapeHtml(teamMap[user.team_id]) : '<span style="color: var(--text-secondary);">-</span>'}</td>
                                    <td data-label="Role">
                                        <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}">
                                            ${escapeHtml(user.role)}
                                        </span>
                                    </td>
                                    <td data-label="Status">
                                        <span class="badge ${user.is_active ? 'badge-success' : 'badge-danger'}">
                                            ${user.is_active ? 'Active' : 'Inactive'}
                                        </span>
                                    </td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <a href="/admin/users/${user.id}" class="btn btn-secondary action-btn" title="View details">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/System/eye-fill.svg" alt="View" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                            <a href="/admin/users/${user.id}/edit" class="btn btn-secondary action-btn" title="Edit user">
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
        const usersContainer = container.querySelector('#users-container');
        usersContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load users: ${error.message}</p>
            </div>
        `;
    }
}

export async function userDetailView(params) {
    const container = document.getElementById('view-container');
    const [, id] = params;

    if (!id) {
        router.navigate('/admin/users');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading user...</div>
        </div>
    `;

    try {
        const user = await api.getUser(id);
        let team = null;

        if (user.team_id) {
            try {
                team = await api.getTeam(user.team_id);
            } catch (error) {
                // Team not found, continue without it
            }
        }

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</h1>
                    <div class="actions">
                        <a href="/admin/users/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
                    </div>
                </div>

                <!-- User Information Card -->
                <div class="card">
                    <h2 class="card-header">User Information</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Full Name</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary); font-weight: 500;">
                                ${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}
                            </p>
                        </div>
                        <div>
                            <label class="form-label">Email Address</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${escapeHtml(user.email)}
                            </p>
                        </div>
                    </div>
                </div>

                <!-- Access & Permissions Card -->
                <div class="card">
                    <h2 class="card-header">Access & Permissions</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1.5rem;">
                        <div>
                            <label class="form-label">Role</label>
                            <div style="margin-top: 0.25rem;">
                                <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}" style="font-size: 0.875rem; padding: 0.375rem 0.875rem;">
                                    ${escapeHtml(user.role)}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="form-label">Account Status</label>
                            <div style="margin-top: 0.25rem;">
                                <span class="badge ${user.is_active ? 'badge-success' : 'badge-danger'}" style="font-size: 0.875rem; padding: 0.375rem 0.875rem;">
                                    ${user.is_active ? 'Active' : 'Inactive'}
                                </span>
                            </div>
                        </div>
                        <div>
                            <label class="form-label">Team</label>
                            <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                                ${team ? `<a href="/admin/teams/${team.id}" style="color: var(--primary); text-decoration: none;">${escapeHtml(team.name)}</a>` : '<span style="color: var(--text-secondary);">No team assigned</span>'}
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
                        <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Delete User</h3>
                        <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                            Permanently delete this user. This action cannot be undone.
                        </p>
                        <button type="button" class="btn btn-danger" id="delete-btn">
                            Delete User
                        </button>
                    </div>
                </div>
                ` : ''}
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        deleteBtn.addEventListener('click', async () => {
            if (confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
                try {
                    await api.deleteUser(id);
                    router.navigate('/admin/users');
                } catch (error) {
                }
            }
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load user: ${error.message}</p>
                <a href="/admin/users" class="btn btn-secondary" style="margin-top: 1rem;">Back to Users</a>
            </div>
        `;
    }
}

export async function userFormView(params) {
    const container = document.getElementById('view-container');
    const [, id, action] = params;
    const isEdit = action === 'edit';

    container.innerHTML = '<div class="loading">Loading...</div>';

    let user = null;
    let teams = [];

    try {
        teams = await api.getTeams();
        if (isEdit && id) {
            user = await api.getUser(id);
        }
    } catch (error) {
        router.navigate('/admin/users');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} User</h1>
            </div>

            <form id="user-form">
                <!-- Personal Information Card -->
                <div class="card">
                    <h2 class="card-header">Personal Information</h2>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem;">
                        <div class="form-group" style="margin-bottom: 0;">
                            <label class="form-label" for="first_name">First Name *</label>
                            <input
                                type="text"
                                id="first_name"
                                class="form-input"
                                required
                                placeholder="John"
                                value="${user ? escapeHtml(user.first_name) : ''}"
                            >
                        </div>
                        <div class="form-group" style="margin-bottom: 0;">
                            <label class="form-label" for="last_name">Last Name *</label>
                            <input
                                type="text"
                                id="last_name"
                                class="form-input"
                                required
                                placeholder="Doe"
                                value="${user ? escapeHtml(user.last_name) : ''}"
                            >
                        </div>
                    </div>
                    <div class="form-group" style="margin-bottom: 0; margin-top: 1.5rem;">
                        <label class="form-label" for="email">Email Address *</label>
                        <input
                            type="email"
                            id="email"
                            class="form-input"
                            required
                            placeholder="john.doe@example.com"
                            value="${user ? escapeHtml(user.email) : ''}"
                        >
                    </div>
                </div>

                ${!isEdit ? `
                <!-- Security Card -->
                <div class="card">
                    <h2 class="card-header">Security</h2>
                    <div class="form-group" style="margin-bottom: 0;">
                        <label class="form-label" for="password">Password *</label>
                        <input
                            type="password"
                            id="password"
                            class="form-input"
                            required
                            minlength="8"
                            placeholder="Minimum 8 characters"
                        >
                        <small style="color: var(--text-secondary);">User will be required to change this password on first login.</small>
                    </div>
                </div>
                ` : ''}

                <!-- Access & Permissions Card -->
                <div class="card">
                    <h2 class="card-header">Access & Permissions</h2>
                    <div class="form-group">
                        <label class="form-label" for="role">Role *</label>
                        <select id="role" class="form-select" required>
                            <option value="user" ${user && user.role === 'user' ? 'selected' : ''}>User</option>
                            <option value="admin" ${user && user.role === 'admin' ? 'selected' : ''}>Admin</option>
                        </select>
                        <small style="color: var(--text-secondary);">Admins can manage users, projects, and sprints. Users can manage their own tickets.</small>
                    </div>
                    <div class="form-group" ${!isEdit ? 'style="margin-bottom: 0;"' : ''}>
                        <label class="form-label" for="team_id">Team</label>
                        <select id="team_id" class="form-select">
                            <option value="">No Team</option>
                            ${teams.map(team => `
                                <option value="${team.id}" ${user && user.team_id === team.id ? 'selected' : ''}>
                                    ${escapeHtml(team.name)}
                                </option>
                            `).join('')}
                        </select>
                        <small style="color: var(--text-secondary);">Assign this user to a team (optional).</small>
                    </div>
                    ${isEdit ? `
                        <div class="form-group" style="margin-bottom: 0;">
                            <label class="form-label">Account Status</label>
                            <label style="display: flex; align-items: center; gap: 0.5rem; cursor: pointer; margin-top: 0.5rem;">
                                <input
                                    type="checkbox"
                                    id="is_active"
                                    ${user && user.is_active ? 'checked' : ''}
                                >
                                <span>Active</span>
                            </label>
                            <small style="color: var(--text-secondary);">Inactive users cannot log in to the system.</small>
                        </div>
                    ` : ''}
                </div>

                ${isEdit ? `
                <!-- Password Reset Card -->
                <div class="card">
                    <h2 class="card-header">Password Reset</h2>
                    <div class="form-group" style="margin-bottom: 0;">
                        <label class="form-label" for="new_password">New Password</label>
                        <input
                            type="password"
                            id="new_password"
                            class="form-input"
                            minlength="8"
                            placeholder="Leave blank to keep current password"
                        >
                        <small style="color: var(--text-secondary);">If you enter a new password, the user will be required to change it on their next login.</small>
                    </div>
                </div>
                ` : ''}

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} User
                        </button>
                        <button type="button" class="btn btn-secondary" onclick="history.back()">Cancel</button>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#user-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const firstName = form.first_name.value.trim();
        const lastName = form.last_name.value.trim();
        const email = form.email.value.trim();
        const role = form.role.value;
        const teamId = form.team_id.value;

        const data = {
            first_name: firstName,
            last_name: lastName,
            email,
            role,
            team_id: teamId || null,
        };

        if (isEdit) {
            data.is_active = form.is_active.checked;
            // Include new password if provided
            const newPassword = form.new_password.value;
            if (newPassword) {
                if (newPassword.length < 8) {
                    alert('Password must be at least 8 characters long');
                    return;
                }
                data.password = newPassword;
            }
        } else {
            const password = form.password.value;
            if (password.length < 8) {
                return;
            }
            data.password = password;
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateUser(id, data);
                router.navigate(`/admin/users/${id}`);
            } else {
                await api.createUser(data);
                router.navigate('/admin/users');
            }
        } catch (error) {
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} User`;
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
