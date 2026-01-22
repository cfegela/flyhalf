import { auth } from './auth.js';

class Router {
    constructor() {
        this.routes = [];
        this.currentRoute = null;
        this.viewContainer = document.getElementById('view-container');

        window.addEventListener('hashchange', () => this.navigate());
        window.addEventListener('load', () => this.navigate());
    }

    addRoute(path, handler, options = {}) {
        this.routes.push({ path, handler, options });
    }

    async navigate(path) {
        if (path) {
            window.location.hash = path;
            return;
        }

        const hash = window.location.hash.slice(1) || '/';
        const [routePath, ...paramParts] = hash.split('/').filter(Boolean);
        const fullPath = '/' + (routePath || '');

        let route = this.routes.find(r => r.path === fullPath);

        if (!route) {
            route = this.routes.find(r => r.path === '/404');
        }

        if (!route) {
            this.viewContainer.innerHTML = '<div class="empty-state"><h2>Page not found</h2></div>';
            return;
        }

        if (route.options.requireAuth && !auth.isAuthenticated()) {
            this.navigate('/login');
            return;
        }

        if (route.options.requireAdmin && !auth.isAdmin()) {
            this.navigate('/');
            return;
        }

        if (route.options.guestOnly && auth.isAuthenticated()) {
            this.navigate('/');
            return;
        }

        this.currentRoute = route;

        const params = paramParts.length > 0 ? paramParts : [];
        await route.handler(params);
    }

    getParams() {
        const hash = window.location.hash.slice(1) || '/';
        const parts = hash.split('/').filter(Boolean);
        return parts.slice(1);
    }
}

export const router = new Router();
