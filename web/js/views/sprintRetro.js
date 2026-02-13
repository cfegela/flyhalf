import { api } from '../api.js';
import { router } from '../router.js';
import { auth } from '../auth.js';
import { escapeHtml, formatDate } from '../utils/formatting.js';
import { getStatusBadgeClass } from '../utils/helpers.js';

export async function sprintRetroView(params) {
    const container = document.getElementById('view-container');
    const [sprintId] = params;

    if (!sprintId) {
        router.navigate('/sprints');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading retrospective...</div>
        </div>
    `;

    try {
        const sprint = await api.getSprint(sprintId);
        const items = await api.getRetroItems(sprintId);

        const goodItems = items.filter(item => item.category === 'good');
        const badItems = items.filter(item => item.category === 'bad');
        const isClosed = sprint.status === 'closed';

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">Sprint Retrospective: ${escapeHtml(sprint.name)}</h1>
                    <div class="actions">
                        <a href="/sprints/${sprintId}" class="btn btn-primary">Details</a>
                        ${!isClosed ? `<a href="/sprints/${sprintId}/board" class="btn btn-primary">Board</a>` : ''}
                        <a href="/sprints/${sprintId}/report" class="btn btn-primary">Report</a>
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
                    </div>
                </div>

                ${isClosed ? `
                <div class="card" style="background: var(--warning-light); border-color: var(--warning); margin-bottom: 1.5rem;">
                    <p style="margin: 0; color: var(--text-primary); font-weight: 500;">
                        ⚠️ This sprint is closed and the retrospective is read-only.
                    </p>
                </div>
                ` : ''}

                <div class="board retro-board-2col">
                    <div class="board-column">
                        <div class="board-column-header">
                            <h2><span style="color: var(--success);">✓</span> Good</h2>
                            <span class="ticket-count">${goodItems.length}</span>
                        </div>
                        <div class="board-column-content" id="good-items">
                            ${renderItems(goodItems, 'good', isClosed)}
                        </div>
                        ${!isClosed ? `
                        <div class="retro-add-item">
                            <textarea
                                id="new-good-item"
                                class="form-input"
                                placeholder="Add something that went well..."
                                maxlength="500"
                                rows="3"
                            ></textarea>
                            <button class="btn btn-primary" onclick="window.addRetroItem('good')">Add</button>
                        </div>
                        ` : ''}
                    </div>

                    <div class="board-column">
                        <div class="board-column-header">
                            <h2><span style="color: var(--danger);">✗</span> Bad</h2>
                            <span class="ticket-count">${badItems.length}</span>
                        </div>
                        <div class="board-column-content" id="bad-items">
                            ${renderItems(badItems, 'bad', isClosed)}
                        </div>
                        ${!isClosed ? `
                        <div class="retro-add-item">
                            <textarea
                                id="new-bad-item"
                                class="form-input"
                                placeholder="Add something that could improve..."
                                maxlength="500"
                                rows="3"
                            ></textarea>
                            <button class="btn btn-primary" onclick="window.addRetroItem('bad')">Add</button>
                        </div>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;

        // Store sprint ID for global functions
        window.currentSprintId = sprintId;

        // Add event listeners for edit/delete buttons
        setupItemActions();

    } catch (error) {
        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">Sprint Retrospective</h1>
                </div>
                <div class="card">
                    <p style="color: var(--danger);">Failed to load retrospective: ${error.message}</p>
                </div>
            </div>
        `;
    }
}

function renderItems(items, category, isClosed) {
    if (items.length === 0) {
        return '<div class="board-empty-state">No items yet. Add one below!</div>';
    }

    const currentUser = auth.getUser();
    const isAdmin = auth.isAdmin();

    return items.map(item => {
        const canEdit = !isClosed && (isAdmin || item.user_id === currentUser.id);

        return `
            <div class="board-ticket retro-item" data-item-id="${item.id}">
                <div class="retro-item-content" id="content-${item.id}">
                    <div class="board-ticket-title">${escapeHtml(item.content)}</div>
                </div>
                ${!isClosed ? `
                <div class="retro-item-edit" id="edit-${item.id}" style="display: none;">
                    <textarea
                        class="form-input"
                        id="edit-content-${item.id}"
                        maxlength="500"
                        rows="3"
                    >${escapeHtml(item.content)}</textarea>
                    <div class="retro-item-actions">
                        <button class="btn btn-primary btn-sm" onclick="window.saveRetroItem('${item.id}', '${category}')">Save</button>
                        <button class="btn btn-secondary btn-sm" onclick="window.cancelEditRetroItem('${item.id}')">Cancel</button>
                    </div>
                </div>
                ` : ''}
                <div class="board-ticket-footer retro-item-actions" id="actions-${item.id}">
                    ${!isClosed ? `
                    <div style="display: flex; align-items: center; gap: 0.25rem; font-size: 0.75rem;">
                        <span style="color: var(--text-secondary);">Votes:</span>
                        <span id="vote-count-${item.id}" style="color: var(--text-secondary);">${item.vote_count}</span>
                        <span style="color: var(--text-secondary);">[</span>
                        <a href="#" class="board-ticket-link" onclick="event.preventDefault(); window.voteRetroItem('${item.id}')">add</a>
                        <span style="color: var(--text-secondary);">|</span>
                        <a href="#" class="board-ticket-link" onclick="event.preventDefault(); window.unvoteRetroItem('${item.id}')">remove</a>
                        <span style="color: var(--text-secondary);">]</span>
                    </div>
                    ` : `
                    <div style="display: flex; align-items: center; gap: 0.25rem; font-size: 0.75rem;">
                        <span style="color: var(--text-secondary);">Votes: ${item.vote_count}</span>
                    </div>
                    `}
                    <div style="display: flex; gap: 1rem;">
                        ${canEdit ? `
                            <a href="#" class="board-ticket-link" onclick="event.preventDefault(); window.editRetroItem('${item.id}')">Edit</a>
                            <a href="#" class="board-ticket-link" style="color: var(--danger);" onclick="event.preventDefault(); window.deleteRetroItem('${item.id}')">Delete</a>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

function setupItemActions() {
    // Global functions for item management
    window.addRetroItem = async (category) => {
        const textarea = document.getElementById(`new-${category}-item`);
        const content = textarea.value.trim();

        if (!content) {
            alert('Please enter some content');
            return;
        }

        if (content.length > 500) {
            alert('Content must be 500 characters or less');
            return;
        }

        try {
            await api.createRetroItem(window.currentSprintId, content, category);
            // Reload the view
            await sprintRetroView([window.currentSprintId]);
        } catch (error) {
            alert(`Failed to add item: ${error.message}`);
        }
    };

    window.editRetroItem = (itemId) => {
        document.getElementById(`content-${itemId}`).style.display = 'none';
        document.getElementById(`actions-${itemId}`).style.display = 'none';
        document.getElementById(`edit-${itemId}`).style.display = 'block';
    };

    window.cancelEditRetroItem = (itemId) => {
        document.getElementById(`content-${itemId}`).style.display = 'block';
        document.getElementById(`actions-${itemId}`).style.display = 'flex';
        document.getElementById(`edit-${itemId}`).style.display = 'none';
    };

    window.saveRetroItem = async (itemId, category) => {
        const textarea = document.getElementById(`edit-content-${itemId}`);
        const content = textarea.value.trim();

        if (!content) {
            alert('Please enter some content');
            return;
        }

        if (content.length > 500) {
            alert('Content must be 500 characters or less');
            return;
        }

        try {
            await api.updateRetroItem(itemId, content, category);
            // Reload the view
            await sprintRetroView([window.currentSprintId]);
        } catch (error) {
            alert(`Failed to update item: ${error.message}`);
        }
    };

    window.deleteRetroItem = async (itemId) => {
        if (!confirm('Are you sure you want to delete this item?')) {
            return;
        }

        try {
            await api.deleteRetroItem(itemId);
            // Reload the view
            await sprintRetroView([window.currentSprintId]);
        } catch (error) {
            alert(`Failed to delete item: ${error.message}`);
        }
    };

    window.voteRetroItem = async (itemId) => {
        try {
            const updatedItem = await api.voteRetroItem(itemId);
            // Update the vote count in the UI without reloading
            const voteCountElement = document.getElementById(`vote-count-${itemId}`);
            if (voteCountElement) {
                voteCountElement.textContent = updatedItem.vote_count;
            }
        } catch (error) {
            alert(`Failed to vote: ${error.message}`);
        }
    };

    window.unvoteRetroItem = async (itemId) => {
        try {
            const updatedItem = await api.unvoteRetroItem(itemId);
            // Update the vote count in the UI without reloading
            const voteCountElement = document.getElementById(`vote-count-${itemId}`);
            if (voteCountElement) {
                voteCountElement.textContent = updatedItem.vote_count;
            }
        } catch (error) {
            alert(`Failed to unvote: ${error.message}`);
        }
    };
}
