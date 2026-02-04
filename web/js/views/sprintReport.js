import { api } from '../api.js';
import { router } from '../router.js';

// Load Chart.js from CDN
function loadChartJS() {
    return new Promise((resolve, reject) => {
        if (window.Chart) {
            resolve();
            return;
        }

        const script = document.createElement('script');
        script.src = 'https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js';
        script.onload = resolve;
        script.onerror = reject;
        document.head.appendChild(script);
    });
}

export async function sprintReportView(params) {
    const container = document.getElementById('view-container');
    const [id] = params;

    if (!id) {
        router.navigate('/sprints');
        return;
    }

    container.innerHTML = `
        <div>
            <div class="loading">Loading sprint report...</div>
        </div>
    `;

    try {
        // Load Chart.js if not already loaded
        await loadChartJS();

        // Fetch sprint report data
        const report = await api.getSprintReport(id);
        const sprint = report.sprint;

        // Calculate sprint progress
        // Parse dates as local dates to avoid timezone issues
        const parseDate = (dateStr) => {
            const [year, month, day] = dateStr.split('T')[0].split('-').map(Number);
            return new Date(year, month - 1, day);
        };
        const startDate = parseDate(sprint.start_date);
        const endDate = parseDate(sprint.end_date);
        const today = new Date();
        const totalDays = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24));
        const daysElapsed = Math.max(0, Math.min(totalDays, Math.ceil((today - startDate) / (1000 * 60 * 60 * 24))));
        const daysRemaining = Math.max(0, totalDays - daysElapsed);
        const isActive = sprint.status === 'active';
        const isCompleted = sprint.status === 'completed';
        const isUpcoming = sprint.status === 'upcoming';

        // Calculate velocity (points per day)
        const velocity = daysElapsed > 0 ? (report.completed_points / daysElapsed).toFixed(2) : 0;

        container.innerHTML = `
            <div>
                <div class="page-header">
                    <h1 class="page-title">${escapeHtml(sprint.name)} - Report</h1>
                    <div class="actions">
                        <a href="/sprints/${id}/board" class="btn btn-primary">Board</a>
                        <a href="/sprints/${id}/retro" class="btn btn-primary">Retro</a>
                        <button class="btn btn-secondary" onclick="history.back()">Back</button>
                    </div>
                </div>

                <!-- Progress Metrics -->
                <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1.5rem;">
                    <!-- Story Points Card -->
                    <div class="card">
                        <h2 class="card-header">Story Points</h2>
                        <div style="display: flex; flex-direction: column; gap: 1rem;">
                            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem;">
                                <div>
                                    <label class="form-label">Total Points</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--text-primary); margin: 0;">
                                        ${report.total_points}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Committed</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--primary); margin: 0;">
                                        ${report.committed_points}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Adopted</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--danger); margin: 0;">
                                        ${report.adopted_points}
                                    </p>
                                </div>
                            </div>
                            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem;">
                                <div>
                                    <label class="form-label">Completed</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--success); margin: 0;">
                                        ${report.completed_points}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Remaining</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--warning); margin: 0;">
                                        ${report.remaining_points}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Velocity</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--primary); margin: 0;">
                                        ${velocity}
                                    </p>
                                    <p style="font-size: 0.75rem; color: var(--text-secondary); margin: 0.25rem 0 0 0;">
                                        pts/day
                                    </p>
                                </div>
                            </div>
                            <div>
                                <div style="background: var(--border); height: 8px; border-radius: 4px; overflow: hidden;">
                                    <div style="background: var(--success); height: 100%; width: ${report.total_points > 0 ? (report.completed_points / report.total_points * 100) : 0}%; transition: width 0.3s ease;"></div>
                                </div>
                                <p style="font-size: 0.875rem; color: var(--text-secondary); margin-top: 0.5rem;">
                                    ${report.total_points > 0 ? Math.round(report.completed_points / report.total_points * 100) : 0}% complete
                                </p>
                            </div>
                        </div>
                    </div>

                    <!-- Tickets Card -->
                    <div class="card">
                        <h2 class="card-header">Tickets</h2>
                        <div style="display: flex; flex-direction: column; gap: 1rem;">
                            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem;">
                                <div>
                                    <label class="form-label">Total Tickets</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--text-primary); margin: 0;">
                                        ${report.total_tickets}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Committed</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--primary); margin: 0;">
                                        ${report.committed_tickets}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Adopted</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--danger); margin: 0;">
                                        ${report.adopted_tickets}
                                    </p>
                                </div>
                            </div>
                            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 1rem;">
                                <div>
                                    <label class="form-label">Completed</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--success); margin: 0;">
                                        ${report.completed_tickets}
                                    </p>
                                </div>
                                <div>
                                    <label class="form-label">Remaining</label>
                                    <p style="font-size: 2rem; font-weight: bold; color: var(--warning); margin: 0;">
                                        ${report.total_tickets - report.completed_tickets}
                                    </p>
                                </div>
                            </div>
                            <div>
                                <div style="background: var(--border); height: 8px; border-radius: 4px; overflow: hidden;">
                                    <div style="background: var(--success); height: 100%; width: ${report.total_tickets > 0 ? (report.completed_tickets / report.total_tickets * 100) : 0}%; transition: width 0.3s ease;"></div>
                                </div>
                                <p style="font-size: 0.875rem; color: var(--text-secondary); margin-top: 0.5rem;">
                                    ${report.total_tickets > 0 ? Math.round(report.completed_tickets / report.total_tickets * 100) : 0}% complete
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Burndown Chart Card -->
                <div class="card" style="margin-top: 1.5rem;">
                    <h2 class="card-header">Burndown Chart</h2>
                    <div style="position: relative; height: 400px; margin-top: 1rem;">
                        <canvas id="burndown-chart"></canvas>
                    </div>
                </div>
            </div>
        `;

        // Render the burndown chart
        renderBurndownChart(report, today, startDate, endDate);

    } catch (error) {
        container.innerHTML = `
            <div class="card">
                <p style="color: var(--danger);">Failed to load sprint report: ${error.message}</p>
                <a href="/sprints" class="btn btn-secondary" style="margin-top: 1rem;">Back to Sprints</a>
            </div>
        `;
    }
}

function renderBurndownChart(report, today, startDate, endDate) {
    const ctx = document.getElementById('burndown-chart');
    if (!ctx) return;

    // Helper function to parse date string as local date (avoiding timezone issues)
    const parseLocalDate = (dateStr) => {
        const [year, month, day] = dateStr.split('-').map(Number);
        return new Date(year, month - 1, day); // month is 0-indexed
    };

    // Prepare data for the chart
    const labels = report.ideal_burndown.map((point, index) => {
        // Use "Start" for the first day (day before sprint begins)
        if (index === 0) {
            return 'Start';
        }
        const date = parseLocalDate(point.date);
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    });

    const idealData = report.ideal_burndown.map(point => point.points);

    // Use actual burndown data from backend
    const actualData = report.actual_burndown.map(point => point.points);

    // Determine where to cut off the actual data (only show up to today)
    const todayIndex = report.ideal_burndown.findIndex(point => {
        const pointDate = parseLocalDate(point.date);
        pointDate.setHours(0, 0, 0, 0);
        const compareDate = new Date(today);
        compareDate.setHours(0, 0, 0, 0);
        return pointDate >= compareDate;
    });

    // Set future days to null so they don't display
    const displayActualData = actualData.map((points, index) => {
        if (todayIndex !== -1 && index > todayIndex) {
            return null;
        }
        return points;
    });

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Remaining Points',
                    data: displayActualData,
                    borderColor: 'rgb(59, 130, 246)',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    borderWidth: 3,
                    pointRadius: 4,
                    pointHoverRadius: 6,
                    tension: 0.1
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return context.parsed.y + ' points';
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Story Points'
                    },
                    ticks: {
                        precision: 0
                    }
                },
                x: {
                    title: {
                        display: true,
                        text: 'Sprint Days'
                    }
                }
            }
        }
    });
}

function renderStatusBreakdown(ticketsByStatus, pointsByStatus) {
    const statuses = ['open', 'in-progress', 'blocked', 'needs-review', 'closed'];
    const statusLabels = {
        'open': 'Open',
        'in-progress': 'In Progress',
        'blocked': 'Blocked',
        'needs-review': 'Needs Review',
        'closed': 'Closed'
    };

    return statuses.map(status => {
        const tickets = ticketsByStatus[status] || 0;
        const points = pointsByStatus[status] || 0;

        if (tickets === 0) return '';

        return `
            <tr>
                <td data-label="Status">
                    <span class="badge ${getStatusBadgeClass(status)}">
                        ${statusLabels[status]}
                    </span>
                </td>
                <td data-label="Tickets">${tickets}</td>
                <td data-label="Story Points">${points}</td>
            </tr>
        `;
    }).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateString) {
    // Parse date string as local date to avoid timezone conversion issues
    const [year, month, day] = dateString.split('T')[0].split('-').map(Number);
    const date = new Date(year, month - 1, day);
    return date.toLocaleDateString();
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
