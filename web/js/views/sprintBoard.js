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
    const [sprint, allTickets, users, currentUser] = await Promise.all([
      api.getSprint(sprintId),
      api.getTickets(),
      api.getUsersForAssignment(),
      api.getMe()
    ]);

    // Redirect to all tickets page if sprint is closed
    if (sprint.status === 'closed') {
      const { router } = await import('../router.js');
      router.navigate('/tickets');
      return;
    }

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

    // Sort each column by sprint_order (highest first)
    Object.keys(columns).forEach(col => {
      columns[col].sort((a, b) => b.sprint_order - a.sprint_order);
    });

    // Helper to parse date strings as local dates (avoiding timezone issues)
    const parseDate = (dateStr) => {
      const [year, month, day] = dateStr.split('T')[0].split('-').map(Number);
      return new Date(year, month - 1, day);
    };

    container.innerHTML = `
      <div class="page-header">
        <div>
          <h1 class="page-title">${sprint.name} - Board</h1>
          <div class="sprint-dates">
            ${parseDate(sprint.start_date).toLocaleDateString()} -
            ${parseDate(sprint.end_date).toLocaleDateString()}
          </div>
        </div>
        <div class="actions">
          <a href="/sprints/${sprintId}" class="btn btn-primary">Details</a>
          <a href="/sprints/${sprintId}/report" class="btn btn-primary">Report</a>
          <a href="/sprints/${sprintId}/retro" class="btn btn-primary">Retro</a>
          <button class="btn btn-secondary" onclick="history.back()">Back</button>
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
    initializeDragAndDrop(container, sprintId, currentUser);
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
      <div class="board-ticket" draggable="true" data-ticket-id="${ticket.id}" data-status="${ticket.status}" data-sprint-order="${ticket.sprint_order}">
        <div class="board-ticket-header">
          <span class="board-ticket-id">#${ticket.id.slice(0, 8)}</span>
          <span class="${badgeClass}" ${isUnderway ? 'data-ticket-id="' + ticket.id + '"' : ''}>${ticket.status}</span>
        </div>
        <div class="board-ticket-title">${ticket.title}</div>
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

function initializeDragAndDrop(container, sprintId, currentUser) {
  const tickets = container.querySelectorAll('.board-ticket');
  const columns = container.querySelectorAll('.board-column-content');

  let draggedTicket = null;
  let draggedFromColumn = null;

  // Add drag event listeners to tickets
  tickets.forEach(ticket => {
    ticket.addEventListener('dragstart', handleDragStart);
    ticket.addEventListener('dragend', handleDragEnd);
    ticket.addEventListener('dragover', handleTicketDragOver);
    ticket.addEventListener('drop', handleTicketDrop);
    ticket.addEventListener('dragenter', handleTicketDragEnter);
    ticket.addEventListener('dragleave', handleTicketDragLeave);
  });

  // Add drop event listeners to columns
  columns.forEach(column => {
    column.addEventListener('dragover', handleDragOver);
    column.addEventListener('drop', handleColumnDrop);
    column.addEventListener('dragenter', handleDragEnter);
    column.addEventListener('dragleave', handleDragLeave);
  });

  function handleDragStart(e) {
    draggedTicket = this;
    draggedFromColumn = this.closest('.board-column-content');
    this.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/html', this.innerHTML);
  }

  function handleDragEnd(e) {
    this.classList.remove('dragging');
    // Remove drag-over class from all columns and tickets
    columns.forEach(col => col.classList.remove('drag-over'));
    tickets.forEach(ticket => ticket.classList.remove('drag-over-ticket'));
  }

  function handleDragOver(e) {
    if (e.preventDefault) {
      e.preventDefault();
    }
    e.dataTransfer.dropEffect = 'move';
    return false;
  }

  function handleTicketDragOver(e) {
    if (e.preventDefault) {
      e.preventDefault();
    }
    e.dataTransfer.dropEffect = 'move';

    if (draggedTicket !== this) {
      this.classList.add('drag-over-ticket');
    }
    return false;
  }

  function handleTicketDragEnter(e) {
    if (draggedTicket !== this) {
      this.classList.add('drag-over-ticket');
    }
  }

  function handleTicketDragLeave(e) {
    this.classList.remove('drag-over-ticket');
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

  async function handleTicketDrop(e) {
    if (e.stopPropagation) {
      e.stopPropagation();
    }
    e.preventDefault();

    const targetTicket = e.currentTarget;
    const targetColumn = targetTicket.closest('.board-column-content');

    if (!draggedTicket || draggedTicket === targetTicket) {
      targetTicket.classList.remove('drag-over-ticket');
      return false;
    }

    targetTicket.classList.remove('drag-over-ticket');

    // Check if we're reordering within the same column
    if (draggedFromColumn === targetColumn) {
      // Reorder within column using fractional indexing
      const allTicketsInColumn = Array.from(targetColumn.querySelectorAll('.board-ticket'));
      const targetIndex = allTicketsInColumn.indexOf(targetTicket);
      const draggedIndex = allTicketsInColumn.indexOf(draggedTicket);

      let newSprintOrder;

      if (draggedIndex < targetIndex) {
        // Moving down - insert after the target ticket
        const nextTicket = allTicketsInColumn[targetIndex + 1];
        if (nextTicket) {
          const targetSprintOrder = parseFloat(targetTicket.dataset.sprintOrder || 0);
          const nextSprintOrder = parseFloat(nextTicket.dataset.sprintOrder || 0);
          newSprintOrder = (targetSprintOrder + nextSprintOrder) / 2.0;
        } else {
          const targetSprintOrder = parseFloat(targetTicket.dataset.sprintOrder || 0);
          newSprintOrder = targetSprintOrder - 1.0;
        }
      } else {
        // Moving up - insert before the target ticket
        const prevTicket = allTicketsInColumn[targetIndex - 1];
        if (prevTicket) {
          const prevSprintOrder = parseFloat(prevTicket.dataset.sprintOrder || 0);
          const targetSprintOrder = parseFloat(targetTicket.dataset.sprintOrder || 0);
          newSprintOrder = (prevSprintOrder + targetSprintOrder) / 2.0;
        } else {
          const targetSprintOrder = parseFloat(targetTicket.dataset.sprintOrder || 0);
          newSprintOrder = targetSprintOrder + 1.0;
        }
      }

      try {
        const ticketId = draggedTicket.dataset.ticketId;
        await api.updateTicketSprintOrder(ticketId, newSprintOrder);
        await loadSprintBoard(container, sprintId);
      } catch (error) {
        console.error('Failed to update ticket sprint order:', error);
        await loadSprintBoard(container, sprintId);
      }
    } else {
      // Moving to a different column - change status
      await handleStatusChange(targetColumn);
    }

    return false;
  }

  async function handleColumnDrop(e) {
    if (e.stopPropagation) {
      e.stopPropagation();
    }
    e.preventDefault();

    const targetColumn = e.currentTarget;
    if (!targetColumn || !draggedTicket) {
      return false;
    }

    targetColumn.classList.remove('drag-over');

    // Only handle status change if dropping on empty column area
    // (not on a ticket, which is handled by handleTicketDrop)
    if (!e.target.classList.contains('board-ticket') &&
        !e.target.closest('.board-ticket')) {
      await handleStatusChange(targetColumn);
    }

    return false;
  }

  async function handleStatusChange(targetColumn) {
    const ticketId = draggedTicket.dataset.ticketId;
    const oldStatus = draggedTicket.dataset.status;
    const targetColumnName = targetColumn.dataset.column;

    // Determine new status based on target column
    let newStatus;
    if (targetColumnName === 'committed') {
      newStatus = 'open';
    } else if (targetColumnName === 'underway') {
      // If already in underway column, keep the current status
      const underwayStatuses = ['in-progress', 'blocked', 'needs-review'];
      if (underwayStatuses.includes(oldStatus)) {
        newStatus = oldStatus;
      } else {
        // Default to in-progress when moving to underway from another column
        newStatus = 'in-progress';
      }
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

      // Check if trying to close ticket with incomplete acceptance criteria
      if (newStatus === 'closed') {
        const hasIncompleteCriteria = ticket.acceptance_criteria &&
          ticket.acceptance_criteria.some(criterion => !criterion.completed);
        if (hasIncompleteCriteria) {
          alert('Cannot close ticket: all acceptance criteria must be completed');
          return;
        }
      }

      ticket.status = newStatus;

      // If ticket is unassigned, assign it to the current user
      if (!ticket.assigned_to && currentUser) {
        ticket.assigned_to = currentUser.id;
      }

      await api.updateTicket(ticketId, ticket);

      // Reload the board to reflect changes
      await loadSprintBoard(container, sprintId);

    } catch (error) {
      // Reload to revert visual change
      await loadSprintBoard(container, sprintId);
    }
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
