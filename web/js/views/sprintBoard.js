import { api } from '../api.js';

export async function sprintBoardView(params) {
  const container = document.getElementById('view-container');
  const [id] = params;

  container.innerHTML = `
    <div class="sprint-board-container">
      <div class="loading">Loading sprint board...</div>
    </div>
  `;

  const boardContainer = container.querySelector('.sprint-board-container');
  await loadSprintBoard(boardContainer, id);
}

async function loadSprintBoard(container, sprintId) {
  try {
    const [sprint, allTickets, users] = await Promise.all([
      api.getSprint(sprintId),
      api.getTickets(),
      api.getUsersForAssignment()
    ]);

    // Create a map of user_id to user for quick lookup
    const userMap = {};
    users.forEach(user => {
      userMap[user.id] = user;
    });

    // Filter tickets for this sprint
    const sprintTickets = allTickets.filter(t => t.sprint_id === sprintId);

    // Group tickets by column
    const columns = {
      committed: sprintTickets.filter(t => t.status === 'open'),
      underway: sprintTickets.filter(t =>
        t.status === 'in-progress' ||
        t.status === 'blocked' ||
        t.status === 'needs-review'
      ),
      completed: sprintTickets.filter(t => t.status === 'closed')
    };

    // Sort each column by priority (highest first)
    Object.keys(columns).forEach(col => {
      columns[col].sort((a, b) => b.priority - a.priority);
    });

    container.innerHTML = `
      <div class="page-header">
        <div>
          <h1>${sprint.name} - Board</h1>
          <div class="sprint-dates">
            ${new Date(sprint.start_date).toLocaleDateString()} -
            ${new Date(sprint.end_date).toLocaleDateString()}
          </div>
        </div>
        <div>
          <a href="/sprints/${sprintId}" class="btn btn-secondary">Back to Details</a>
        </div>
      </div>

      <div class="board">
        <div class="board-column" data-column="committed" data-status="open">
          <div class="board-column-header">
            <h2>Committed</h2>
            <span class="ticket-count">${columns.committed.length}</span>
          </div>
          <div class="board-column-content" data-column="committed">
            ${renderTickets(columns.committed, userMap)}
          </div>
        </div>

        <div class="board-column" data-column="underway">
          <div class="board-column-header">
            <h2>Underway</h2>
            <span class="ticket-count">${columns.underway.length}</span>
          </div>
          <div class="board-column-content" data-column="underway">
            ${renderTickets(columns.underway, userMap)}
          </div>
        </div>

        <div class="board-column" data-column="completed" data-status="closed">
          <div class="board-column-header">
            <h2>Completed</h2>
            <span class="ticket-count">${columns.completed.length}</span>
          </div>
          <div class="board-column-content" data-column="completed">
            ${renderTickets(columns.completed, userMap)}
          </div>
        </div>
      </div>
    `;

    // Initialize drag and drop
    initializeDragAndDrop(container, sprintId);

    // Initialize status badge clicks for underway tickets
    initializeStatusBadgeClicks(container, sprintId);

  } catch (error) {
    container.innerHTML = `
      <div class="error-state">
        <p>Error loading sprint board: ${error.message}</p>
        <a href="/sprints" class="btn">Back to Sprints</a>
      </div>
    `;
  }
}

function renderTickets(tickets, userMap) {
  if (tickets.length === 0) {
    return '<div class="board-empty-state">No tickets</div>';
  }

  return tickets.map(ticket => {
    const isUnderway = ['in-progress', 'blocked', 'needs-review'].includes(ticket.status);
    const badgeClass = isUnderway ? 'badge badge-' + ticket.status + ' badge-clickable' : 'badge badge-' + ticket.status;
    const assignee = ticket.assigned_to ? userMap[ticket.assigned_to] : null;
    const assigneeDisplay = assignee ? `${assignee.first_name} ${assignee.last_name}` : 'Unassigned';
    const sizeDisplay = getSizeLabel(ticket.size);

    return `
      <div class="board-ticket" draggable="true" data-ticket-id="${ticket.id}" data-status="${ticket.status}">
        <div class="board-ticket-header">
          <span class="board-ticket-id">#${ticket.id.slice(0, 8)}</span>
          <span class="${badgeClass}" ${isUnderway ? 'data-ticket-id="' + ticket.id + '"' : ''}>${ticket.status}</span>
        </div>
        <div class="board-ticket-title">${ticket.title}</div>
        ${ticket.description ? `<div class="board-ticket-description">${ticket.description.substring(0, 100)}${ticket.description.length > 100 ? '...' : ''}</div>` : ''}
        <div class="board-ticket-meta">
          <span class="board-ticket-assignee">${assigneeDisplay}</span>
          <span class="board-ticket-size">${sizeDisplay}</span>
        </div>
        <div class="board-ticket-footer">
          <a href="/tickets/${ticket.id}" class="board-ticket-link" onclick="event.stopPropagation()">View</a>
        </div>
      </div>
    `;
  }).join('');
}

function getSizeLabel(size) {
  if (!size) return '-';
  switch (size) {
    case 1: return 'S';
    case 2: return 'M';
    case 3: return 'L';
    case 5: return 'XL';
    case 8: return 'XXL';
    default: return '-';
  }
}

function initializeDragAndDrop(container, sprintId) {
  const tickets = container.querySelectorAll('.board-ticket');
  const columns = container.querySelectorAll('.board-column-content');

  let draggedTicket = null;

  // Add drag event listeners to tickets
  tickets.forEach(ticket => {
    ticket.addEventListener('dragstart', handleDragStart);
    ticket.addEventListener('dragend', handleDragEnd);
  });

  // Add drop event listeners to columns
  columns.forEach(column => {
    column.addEventListener('dragover', handleDragOver);
    column.addEventListener('drop', handleDrop);
    column.addEventListener('dragenter', handleDragEnter);
    column.addEventListener('dragleave', handleDragLeave);
  });

  function handleDragStart(e) {
    draggedTicket = this;
    this.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/html', this.innerHTML);
  }

  function handleDragEnd(e) {
    this.classList.remove('dragging');
    // Remove drag-over class from all columns
    columns.forEach(col => col.classList.remove('drag-over'));
  }

  function handleDragOver(e) {
    if (e.preventDefault) {
      e.preventDefault();
    }
    e.dataTransfer.dropEffect = 'move';
    return false;
  }

  function handleDragEnter(e) {
    if (e.target.classList.contains('board-column-content')) {
      e.target.classList.add('drag-over');
    }
  }

  function handleDragLeave(e) {
    if (e.target.classList.contains('board-column-content')) {
      e.target.classList.remove('drag-over');
    }
  }

  async function handleDrop(e) {
    if (e.stopPropagation) {
      e.stopPropagation();
    }
    e.preventDefault();

    const targetColumn = e.target.closest('.board-column-content');
    if (!targetColumn || !draggedTicket) {
      return false;
    }

    targetColumn.classList.remove('drag-over');

    // Get ticket data
    const ticketId = draggedTicket.dataset.ticketId;
    const oldStatus = draggedTicket.dataset.status;
    const targetColumnName = targetColumn.dataset.column;

    // Determine new status based on target column
    let newStatus;
    if (targetColumnName === 'committed') {
      newStatus = 'open';
    } else if (targetColumnName === 'underway') {
      // Default to in-progress when moving to underway
      newStatus = 'in-progress';
    } else if (targetColumnName === 'completed') {
      newStatus = 'closed';
    }

    // Don't update if status hasn't changed
    if (oldStatus === newStatus) {
      return false;
    }

    try {
      // Update ticket status via API
      const ticket = await api.getTicket(ticketId);
      ticket.status = newStatus;
      await api.updateTicket(ticketId, ticket);

      // Reload the board to reflect changes
      loadSprintBoard(container, sprintId);

    } catch (error) {
      // Reload to revert visual change
      loadSprintBoard(container, sprintId);
    }

    return false;
  }
}

function initializeStatusBadgeClicks(container, sprintId) {
  const clickableBadges = container.querySelectorAll('.badge-clickable');

  clickableBadges.forEach(badge => {
    badge.addEventListener('click', function(e) {
      e.stopPropagation();

      // Close any existing dropdowns
      const existingDropdown = document.querySelector('.status-dropdown');
      if (existingDropdown) {
        existingDropdown.remove();
      }

      const ticketId = this.dataset.ticketId;
      const currentStatus = this.textContent;
      const rect = this.getBoundingClientRect();

      // Create dropdown
      const dropdown = document.createElement('div');
      dropdown.className = 'status-dropdown';
      dropdown.innerHTML = `
        <div class="status-dropdown-item" data-status="in-progress">in-progress</div>
        <div class="status-dropdown-item" data-status="blocked">blocked</div>
        <div class="status-dropdown-item" data-status="needs-review">needs-review</div>
      `;

      // Position dropdown
      dropdown.style.position = 'fixed';
      dropdown.style.top = `${rect.bottom + 5}px`;
      dropdown.style.left = `${rect.left}px`;

      document.body.appendChild(dropdown);

      // Handle dropdown item clicks
      const items = dropdown.querySelectorAll('.status-dropdown-item');
      items.forEach(item => {
        item.addEventListener('click', async function(e) {
          e.stopPropagation();
          const newStatus = this.dataset.status;

          if (newStatus !== currentStatus) {
            try {
              // Update ticket status
              const ticket = await api.getTicket(ticketId);
              ticket.status = newStatus;
              await api.updateTicket(ticketId, ticket);

              // Reload the board
              await loadSprintBoard(container, sprintId);
            } catch (error) {
              console.error('Failed to update ticket status:', error);
            }
          }

          dropdown.remove();
        });
      });

      // Close dropdown when clicking outside
      const closeDropdown = (e) => {
        if (!dropdown.contains(e.target) && e.target !== badge) {
          dropdown.remove();
          document.removeEventListener('click', closeDropdown);
        }
      };
      setTimeout(() => document.addEventListener('click', closeDropdown), 0);
    });
  });
}
