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
        const usersContainer = container.querySelector('#users-container');

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
                                <th>Role</th>
                                <th>Status</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${users.map(user => `
                                <tr>
                                    <td><strong>${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</strong></td>
                                    <td>${escapeHtml(user.email)}</td>
                                    <td>
                                        <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}">
                                            ${escapeHtml(user.role)}
                                        </span>
                                    </td>
                                    <td>
                                        <span class="badge ${user.is_active ? 'badge-success' : 'badge-danger'}">
                                            ${user.is_active ? 'Active' : 'Inactive'}
                                        </span>
                                    </td>
                                    <td>
                                        <div class="actions">
                                            <a href="/admin/users/${user.id}" class="btn btn-secondary action-btn">
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

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</h1>
                    <div class="actions">
                        <a href="/admin/users/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-danger" id="delete-btn">Delete</button>
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
                    </div>
                </div>
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

    let user = null;
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading user...</div>';
        try {
            user = await api.getUser(id);
        } catch (error) {
            router.navigate('/admin/users');
            return;
        }
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
                    <div class="form-group" ${!isEdit ? 'style="margin-bottom: 0;"' : ''}>
                        <label class="form-label" for="role">Role *</label>
                        <select id="role" class="form-select" required>
                            <option value="user" ${user && user.role === 'user' ? 'selected' : ''}>User</option>
                            <option value="admin" ${user && user.role === 'admin' ? 'selected' : ''}>Admin</option>
                        </select>
                        <small style="color: var(--text-secondary);">Admins can manage users, epics, and sprints. Users can manage their own tickets.</small>
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

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} User
                        </button>
                        <a href="${isEdit ? `#/admin/users/${id}` : '#/admin/users'}" class="btn btn-secondary">Cancel</a>
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

        const data = {
            first_name: firstName,
            last_name: lastName,
            email,
            role,
        };

        if (isEdit) {
            data.is_active = form.is_active.checked;
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
