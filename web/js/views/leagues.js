import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';

export async function leaguesListView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Leagues</h1>
                <a href="/admin/leagues/new" class="btn btn-primary">Create League</a>
            </div>
            <div id="leagues-container">
                <div class="loading">Loading leagues...</div>
            </div>
        </div>
    `;

    try {
        const leagues = await api.getLeagues();
        const leaguesContainer = container.querySelector('#leagues-container');

        if (leagues.length === 0) {
            leaguesContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">🏆</div>
                    <h2>No leagues yet</h2>
                    <p>Create the first league to get started</p>
                    <a href="/admin/leagues/new" class="btn btn-primary" style="margin-top: 1rem;">
                        Create League
                    </a>
                </div>
            `;
            return;
        }

        leaguesContainer.innerHTML = `
            <div class="card">
                <div class="table-container">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Description</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${leagues.map(league => `
                                <tr data-league-id="${league.id}">
                                    <td data-label="Name"><strong>${escapeHtml(league.name)}</strong></td>
                                    <td data-label="Description">${league.description ? escapeHtml(league.description.substring(0, 100)) + (league.description.length > 100 ? '...' : '') : '-'}</td>
                                    <td data-label="Actions">
                                        <div class="actions">
                                            <a href="/admin/leagues/${league.id}" class="btn btn-secondary action-btn" title="View details">
                                                <img src="https://cdn.jsdelivr.net/npm/remixicon@4.8.0/icons/System/eye-fill.svg" alt="View" style="width: 20px; height: 20px; display: block;">
                                            </a>
                                            <a href="/admin/leagues/${league.id}/edit" class="btn btn-secondary action-btn" title="Edit league">
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
        const leaguesContainer = container.querySelector('#leagues-container');
        leaguesContainer.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load leagues: ${error.message}</p>
            </div>
        `;
    }
}

export async function leagueDetailView(params) {
    const container = document.getElementById('view-container');
    const [, id] = params;

    if (!id) {
        router.navigate('/admin/leagues');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading league...</div>
        </div>
    `;

    try {
        const league = await api.getLeague(id);
        const teams = await api.getTeams();

        // Filter teams that belong to this league
        const leagueTeams = teams.filter(team => team.league_id === id);

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(league.name)}</h1>
                    <div class="actions">
                        <a href="/admin/leagues/${id}/edit" class="btn btn-primary">Edit</a>
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
                    </div>
                </div>

                <!-- League Information Card -->
                <div class="card">
                    <h2 class="card-header">League Information</h2>
                    <div>
                        <label class="form-label">Description</label>
                        <p style="white-space: pre-wrap; line-height: 1.6; color: var(--text-primary); margin-top: 0.25rem;">${league.description ? escapeHtml(league.description) : '<span style="color: var(--text-secondary); font-style: italic;">No description provided</span>'}</p>
                    </div>
                </div>

                <!-- League Teams Card -->
                <div class="card">
                    <h2 class="card-header">Teams (${leagueTeams.length})</h2>
                    ${leagueTeams.length === 0 ? `
                        <div class="empty-state">
                            <div class="empty-state-icon">👥</div>
                            <p>No teams assigned to this league</p>
                        </div>
                    ` : `
                        <div class="table-container">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Description</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${leagueTeams.map(team => `
                                        <tr class="clickable-row" data-team-id="${team.id}" style="cursor: pointer;">
                                            <td data-label="Name">
                                                <strong>${escapeHtml(team.name)}</strong>
                                            </td>
                                            <td data-label="Description">
                                                ${team.description ? escapeHtml(team.description.substring(0, 100)) + (team.description.length > 100 ? '...' : '') : '-'}
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
                        <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Delete League</h3>
                        <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                            Permanently delete this league. Teams will not be deleted but will be unassigned from this league.
                        </p>
                        <button type="button" class="btn btn-danger" id="delete-btn">
                            Delete League
                        </button>
                    </div>
                </div>
                ` : ''}
            </div>
        `;

        const deleteBtn = container.querySelector('#delete-btn');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', async () => {
                if (confirm('Are you sure you want to delete this league? Teams will not be deleted but will be unassigned from this league.')) {
                    try {
                        await api.deleteLeague(id);
                        router.navigate('/admin/leagues');
                    } catch (error) {
                        alert('Failed to delete league: ' + error.message);
                    }
                }
            });
        }

        // Make team rows clickable to navigate to team details
        const clickableRows = container.querySelectorAll('.clickable-row');
        clickableRows.forEach(row => {
            row.addEventListener('click', (e) => {
                const teamId = row.dataset.teamId;
                router.navigate(`/admin/teams/${teamId}`);
            });
        });
    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load league: ${error.message}</p>
                <a href="/admin/leagues" class="btn btn-secondary" style="margin-top: 1rem;">Back to Leagues</a>
            </div>
        `;
    }
}

export async function leagueFormView(params) {
    const container = document.getElementById('view-container');
    const [, id, action] = params;
    const isEdit = action === 'edit';

    let league = null;
    if (isEdit && id) {
        container.innerHTML = '<div class="loading">Loading league...</div>';
        try {
            league = await api.getLeague(id);
        } catch (error) {
            router.navigate('/admin/leagues');
            return;
        }
    }

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">${isEdit ? 'Edit' : 'Create'} League</h1>
            </div>

            <form id="league-form">
                <!-- League Information Card -->
                <div class="card">
                    <h2 class="card-header">League Information</h2>
                    <div class="form-group">
                        <label class="form-label" for="name">League Name *</label>
                        <input
                            type="text"
                            id="name"
                            class="form-input"
                            required
                            placeholder="e.g., Premier League, Division 1, Champions League"
                            value="${league ? escapeHtml(league.name) : ''}"
                        >
                    </div>
                    <div class="form-group" style="margin-bottom: 0;">
                        <label class="form-label" for="description">Description</label>
                        <textarea
                            id="description"
                            class="form-textarea"
                            placeholder="Describe the league's purpose and structure..."
                        >${league ? escapeHtml(league.description || '') : ''}</textarea>
                    </div>
                </div>

                <!-- Form Actions -->
                <div class="card">
                    <div style="display: flex; gap: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            ${isEdit ? 'Update' : 'Create'} League
                        </button>
                        <button type="button" class="btn btn-secondary" onclick="history.back()">Cancel</button>
                    </div>
                </div>
            </form>
        </div>
    `;

    const form = container.querySelector('#league-form');
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
                await api.updateLeague(id, data);
                router.navigate('/admin/leagues');
            } else {
                await api.createLeague(data);
                router.navigate('/admin/leagues');
            }
        } catch (error) {
            alert('Failed to save league: ' + error.message);
            submitBtn.disabled = false;
            submitBtn.textContent = `${isEdit ? 'Update' : 'Create'} League`;
        }
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
