import { auth } from './auth.js';
import { router } from './router.js';
import { initNavbar } from './components/navbar.js';
import { loginView } from './views/login.js';
import {
    ticketsListView,
    ticketDetailView,
    ticketFormView
} from './views/tickets.js';
import {
    projectsListView,
    projectDetailView,
    projectFormView
} from './views/projects.js';
import {
    sprintsListView,
    sprintDetailView,
    sprintFormView
} from './views/sprints.js';
import { sprintBoardView } from './views/sprintBoard.js';
import { sprintReportView } from './views/sprintReport.js';
import { sprintRetroView } from './views/sprintRetro.js';
import {
    usersListView,
    userDetailView,
    userFormView
} from './views/admin.js';
import {
    teamsListView,
    teamDetailView,
    teamFormView
} from './views/teams.js';
import {
    leaguesListView,
    leagueDetailView,
    leagueFormView
} from './views/leagues.js';
import { settingsView, forcePasswordChangeView } from './views/settings.js';

async function initApp() {
    initNavbar();

    router.addRoute('/login', loginView, { guestOnly: true });
    router.addRoute('/force-password-change', forcePasswordChangeView, { requireAuth: true, allowPasswordChange: true });
    router.addRoute('/', ticketsListView, { requireAuth: true });
    router.addRoute('/tickets', ticketsListView, { requireAuth: true });
    router.addRoute('/tickets/new', (params) => ticketFormView(['new']), { requireAuth: true });
    router.addRoute('/tickets/:id/edit', (params) => ticketFormView([params[0], 'edit']), { requireAuth: true });
    router.addRoute('/tickets/:id', ticketDetailView, { requireAuth: true });
    router.addRoute('/projects', projectsListView, { requireAuth: true });
    router.addRoute('/projects/new', (params) => projectFormView(['new']), { requireAuth: true });
    router.addRoute('/projects/:id/edit', (params) => projectFormView([params[0], 'edit']), { requireAuth: true });
    router.addRoute('/projects/:id', projectDetailView, { requireAuth: true });
    router.addRoute('/sprints', sprintsListView, { requireAuth: true });
    router.addRoute('/sprints/new', (params) => sprintFormView(['new']), { requireAuth: true });
    router.addRoute('/sprints/:id/board', sprintBoardView, { requireAuth: true });
    router.addRoute('/sprints/:id/report', sprintReportView, { requireAuth: true });
    router.addRoute('/sprints/:id/retro', sprintRetroView, { requireAuth: true });
    router.addRoute('/sprints/:id/edit', (params) => sprintFormView([params[0], 'edit']), { requireAuth: true });
    router.addRoute('/sprints/:id', sprintDetailView, { requireAuth: true });
    router.addRoute('/settings', settingsView, { requireAuth: true });
    router.addRoute('/admin/users', usersListView, { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/new', (params) => userFormView(['admin', 'new']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/:id/edit', (params) => userFormView(['admin', params[0], 'edit']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/:id', (params) => userDetailView(['admin', params[0]]), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/teams', teamsListView, { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/teams/new', (params) => teamFormView(['admin', 'new']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/teams/:id/edit', (params) => teamFormView(['admin', params[0], 'edit']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/teams/:id', (params) => teamDetailView(['admin', params[0]]), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/leagues', leaguesListView, { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/leagues/new', (params) => leagueFormView(['admin', 'new']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/leagues/:id/edit', (params) => leagueFormView(['admin', params[0], 'edit']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/leagues/:id', (params) => leagueDetailView(['admin', params[0]]), { requireAuth: true, requireAdmin: true });

    await auth.init();

    if (!auth.isAuthenticated() && !window.location.pathname.includes('login')) {
        router.navigate('/login');
    } else {
        router.navigate();
    }
}

initApp();
