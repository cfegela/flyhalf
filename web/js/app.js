import { auth } from './auth.js';
import { router } from './router.js';
import { initNavbar } from './components/navbar.js';
import { loginView } from './views/login.js';
import { dashboardView } from './views/dashboard.js';
import {
    resourcesListView,
    resourceDetailView,
    resourceFormView
} from './views/resources.js';
import {
    usersListView,
    userDetailView,
    userFormView
} from './views/admin.js';

async function initApp() {
    initNavbar();

    router.addRoute('/login', loginView, { guestOnly: true });
    router.addRoute('/', dashboardView, { requireAuth: true });
    router.addRoute('/resources', resourcesListView, { requireAuth: true });
    router.addRoute('/resources/new', () => resourceFormView(['new']), { requireAuth: true });
    router.addRoute('/resources/:id', resourceDetailView, { requireAuth: true });
    router.addRoute('/resources/:id/edit', resourceFormView, { requireAuth: true });
    router.addRoute('/admin/users', usersListView, { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/new', () => userFormView(['admin', 'new']), { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/:id', userDetailView, { requireAuth: true, requireAdmin: true });
    router.addRoute('/admin/users/:id/edit', userFormView, { requireAuth: true, requireAdmin: true });

    await auth.init();

    if (!auth.isAuthenticated() && !window.location.hash.includes('login')) {
        router.navigate('/login');
    } else {
        router.navigate();
    }
}

initApp();
