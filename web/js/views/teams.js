import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function teamsListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Teams</h1>
                <a href="/admin/teams/new" class="btn btn-primary">Create Team</a>
            </div>
            <div id="teams-container">
                <div class="loading">Loading teams...</div>
            </div>
        </div>
    `;

    try {
        const teams = await api.getTeams();
        const leagues = await api.getLeagues();
        const teamsContainer = container.querySelector('#teams-container');

        // Create a map of league_id to league for quick lookup
        const leagueMap = {};
        leagues.forEach(league => {
            leagueMap[league.id] = league;
        });

        if (teams.length === 0) {
            teamsContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">👥</div>
                    <h2>No teams yet</h2>
                    <p>Create the first team to get started</p>
                    <a href="/admin/teams/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create Team
                    </a>
                </div>
            `;
            return;
        }

        teamsContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>League</th>
                                <th>Description</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${teams.map(team => {
                                const league = team.league_id ? leagueMap[team.league_id] : null;
                                return `
                                <tr data-team-id="${team.id}">
                                    <td data-label="Name"><strong>${escapeHtml(team.name)}</strong></td>
                                    <td data-label="League">${league ? escapeHtml(league.name) : '-'}</td>
                                    <td data-label="Description">${team.description ? escapeHtml(team.description.substring(0, 100)) + (team.description.length > 100 ? '...' : '') : '-'}</td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <a href="/admin/teams/${team.id}" class="btn btn-secondary action-btn" title="View details">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/System/eye-fill.svg" alt="View" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                            <a href="/admin/teams/${team.id}/edit" class="btn btn-secondary action-btn" title="Edit team">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/Design/pencil-ai-fill.svg" alt="Edit" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                        </div>
                                    </td>
                                </tr>
                            `}).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;

    } catch (error) {
        const teamsContainer = container.querySelector('#teams-container');
        teamsContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load teams: ${error.message}</p>
            </div>
        `;
    }
}

export async function teamDetailView(params) {
    const container = document.getElementById('view-container');
    const [, id] = params;

    if (!id) {
        router.navigate('/admin/teams');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading team...</div>
        </div>
    `;

    try {
        const team = await api.getTeam(id);
        const users = await api.getUsers();
        const leagues = await api.getLeagues();

        // Filter users who belong to this team
        const teamMembers = users.filter(user => user.team_id === id);

        // Find the team's league
        const league = team.league_id ? leagues.find(l => l.id === team.league_id) : null;

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(team.name)}</h1>
                    <div class="actions">
                        <a href="/admin/teams/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
                    </div>
                </div>

                <!-- Team Information Card -->
                <div class="card">
                    <h2 class="card-header">Team Information</h2>
                    ${league ? `
                    <div style="margin-bottom: 1.5rem;">
                        <label class="form-label">League</label>
                        <p style="line-height: 1.6; color: var(--text-primary); margin-top: 0.25rem;">
                            <a href="/admin/leagues/${league.id}" style="color: var(--primary); text-decoration: none;">
                                ${escapeHtml(league.name)}
                            </a>
                        </p>
                    </div>
                    ` : ''}
                    <div>
                        <label class="form-label">Description</label>
                        <p style="white-space: pre-wrap; line-height: 1.6; color: var(--text-primary); margin-top: 0.25rem;">${team.description ? escapeHtml(team.description) : '<span style="color: var(--text-secondary); font-style: italic;">No description provided</span>'}</p>
                    </div>
                </div>

                <!-- Team Members Card -->
                <div class="card">
                    <h2 class="card-header">Team Members (${teamMembers.length})</h2>
                    ${teamMembers.length === 0 ? `
                        <div class="empty-state">
                            <div class="empty-state-icon">👥</div>
                            <p>No members assigned to this team</p>
                        </div>
                    ` : `
                        <div class="table-container">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Email</th>
                                        <th>Role</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${teamMembers.map(user => `
                                        <tr class="clickable-row" data-user-id="${user.id}" style="cursor: pointer;">
                                            <td data-label="Name">
                                                <strong>${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</strong>
                                            </td>
                                            <td data-label="Email">
                                                ${escapeHtml(user.email)}
                                            </td>
                                            <td data-label="Role">
                                                <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}">
                                                    ${escapeHtml(user.role)}
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
                        <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Delete Team</h3>
                        <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                            Permanently delete this team. Team members will not be deleted but will be unassigned from this team.
                        </p>
                        <button type="button" class="btn btn-danger" id="delete-btn">
                            Delete Team
                        </button>
                    </div>
                </div>
                ` : ''}
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', async () => {
                if (confirm('Are you sure you want to delete this team? Team members will not be deleted but will be unassigned from this team.')) {
                    try {
                        await api.deleteTeam(id);
                        router.navigate('/admin/teams');
                    } catch (error) {
                        alert('Failed to delete team: ' + error.message);
                    }
                }
            });
        }

        // Make user rows clickable to navigate to user details
        const clickableRows = container.querySelectorAll('.clickable-row');
        clickableRows.forEach(row => {
            row.addEventListener('click', (e) => {
                const userId = row.dataset.userId;
                router.navigate(`/admin/users/${userId}`);
            });
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load team: ${error.message}</p>
                <a href="/admin/teams" class="btn btn-secondary" style="margin-top: 1rem;">Back to Teams</a>
            </div>
        `;
    }
}

export async function teamFormView(params) {
    const container = document.getElementById('view-container');
    const [, id, action] = params;
    const isEdit = action === 'edit';

    let team = null;
    let leagues = [];

    container.innerHTML = '<div class="loading">Loading...</div>';

    try {
        leagues = await api.getLeagues();
        if (isEdit && id) {
            team = await api.getTeam(id);
        }
    } catch (error) {
        if (isEdit) {
            router.navigate('/admin/teams');
            return;
        }
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} Team</h1>
            </div>

            <form id="team-form">
                <!-- Team Information Card -->
                <div class="card">
                    <h2 class="card-header">Team Information</h2>
                    <div class="form-group">
                        <label class="form-label" for="name">Team Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
                            placeholder="e.g., Engineering, Sales, Marketing"
                            value="${team ? escapeHtml(team.name) : ''}"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="league_id">League</label>
                        <select id="league_id" class="form-input">
                            <option value="">No League</option>
                            ${leagues.map(league => `
                                <option value="${league.id}" ${team && team.league_id === league.id ? 'selected' : ''}>
                                    ${escapeHtml(league.name)}
                                </option>
                            `).join('')}
                        </select>
                    </div>
                    <div class="form-group" style="margin-bottom: 0;">
                        <label class="form-label" for="description">Description</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                            placeholder="Describe the team's purpose and responsibilities..."
                        >${team ? escapeHtml(team.description || '') : ''}</textarea>
                    </div>
                </div>

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} Team
                        </button>
                        <button type="button" class="btn btn-secondary" onclick="history.back()">Cancel</button>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#team-form');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = form.name.value.trim();
        const description = form.description.value.trim();
        const leagueIdValue = form.league_id.value;

        const data = { name, description };

        // Only include league_id if a league is selected
        if (leagueIdValue) {
            data.league_id = leagueIdValue;
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = isEdit ? 'Updating...' : 'Creating...';

        try {
            if (isEdit) {
                await api.updateTeam(id, data);
                router.navigate('/admin/teams');
            } else {
                await api.createTeam(data);
                router.navigate('/admin/teams');
            }
        } catch (error) {
            alert('Failed to save team: ' + error.message);
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} Team`;
        }
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
