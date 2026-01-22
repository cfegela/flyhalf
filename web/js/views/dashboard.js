import { auth } from '../auth.js';
import { api } from '../api.js';
import { toast } from '../components/toast.js';

export async function dashboardView() {
    const container = document.getElementById('view-container');
    const user = auth.getUser();

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Dashboard</h1>
            </div>
            <div class="card">
                <h2 class="card-header">Welcome back, ${user.first_name}!</h2>
                <div style="color: var(--text-secondary);">
                    <p>Email: ${user.email}</p>
                    <p>Role: <span class="badge badge-primary">${user.role}</span></p>
                    <p>Status: <span class="badge ${user.is_active ? 'badge-success' : 'badge-danger'}">
                        ${user.is_active ? 'Active' : 'Inactive'}
                    </span></p>
                </div>
            </div>
            <div id="stats-container">
                <div class="loading">Loading statistics...</div>
            </div>
        </div>
    `;

    try {
        const resources = await api.getResources();
        const statsContainer = container.querySelector('#stats-container');

        const activeResources = resources.filter(r => r.status === 'active').length;
        const totalResources = resources.length;

        statsContainer.innerHTML = `
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem; margin-top: 1.5rem;">
                <div class="card">
                    <h3 style="font-size: 0.875rem; color: var(--text-secondary); margin-bottom: 0.5rem;">
                        Total Resources
                    </h3>
                    <p style="font-size: 2rem; font-weight: 700; color: var(--primary);">
                        ${totalResources}
                    </p>
                </div>
                <div class="card">
                    <h3 style="font-size: 0.875rem; color: var(--text-secondary); margin-bottom: 0.5rem;">
                        Active Resources
                    </h3>
                    <p style="font-size: 2rem; font-weight: 700; color: var(--success);">
                        ${activeResources}
                    </p>
                </div>
            </div>
            <div class="card" style="margin-top: 1.5rem;">
                <h3 class="card-header">Quick Actions</h3>
                <div style="display: flex; gap: 1rem; flex-wrap: wrap;">
                    <a href="#/resources" class="btn btn-primary">View Resources</a>
                    <a href="#/resources/new" class="btn btn-secondary">Create Resource</a>
                    ${auth.isAdmin() ? '<a href="#/admin/users" class="btn btn-secondary">Manage Users</a>' : ''}
                </div>
            </div>
        `;
    } catch (error) {
        const statsContainer = container.querySelector('#stats-container');
        statsContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load statistics: ${error.message}</p>
            </div>
        `;
    }
}
