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
    usersListView,
    userDetailView,
    userFormView
} from './views/admin.js';

async function initApp() {
    initNavbar();

    router.addRoute('/login', loginView, { guestOnly: true });
    router.addRoute('/', ticketsListView, { requireAuth: true });
    router.addRoute('/tickets', ticketsListView, { requireAuth: true });
    router.addRoute('/tickets/new', (params) => ticketFormView(['new']), { requireAuth: true });
    router.addRoute('/tickets/:id/edit', (params) => ticketFormView([params[0], 'edit']), { requireAuth: true });
    router.addRoute('/tickets/:id', ticketDetailView, { requireAuth: true });
    router.addRoute('/admin/users', usersListView, { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/new', (params) => userFormView(['admin', 'new']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/:id/edit', (params) => userFormView(['admin', params[0], 'edit']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/:id', (params) => userDetailView(['admin', params[0]]), { requireAuth: true, requireAdmin: true });

    await auth.init();

    if (!auth.isAuthenticated() && !window.location.hash.includes('login')) {
        router.navigate('/login');
    } else {
        router.navigate();
    }
}

initApp();
