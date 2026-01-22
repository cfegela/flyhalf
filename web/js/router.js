import { auth } from './auth.js';

class Router {
    constructor() {
        this.routes = [];
        this.currentRoute = null;
        this.viewContainer = document.getElementById('view-container');

        window.addEventListener('hashchange', () => this.navigate());
    }

    addRoute(path, handler, options = {}) {
        this.routes.push({ path, handler, options });
    }

    matchRoute(urlPath) {
        for (const route of this.routes) {
            const routeParts = route.path.split('/').filter(Boolean);
            const urlParts = urlPath.split('/').filter(Boolean);

            if (routeParts.length !== urlParts.length) {
                continue;
            }

            const params = [];
            let isMatch = true;

            for (let i = 0; i < routeParts.length; i++) {
                if (routeParts[i].startsWith(':')) {
                    params.push(urlParts[i]);
                } else if (routeParts[i] !== urlParts[i]) {
                    isMatch = false;
                    break;
                }
            }

            if (isMatch) {
                return { route, params };
            }
        }

        return null;
    }

    async navigate(path) {
        if (path) {
            window.location.hash = path;
            return;
        }

        const hash = window.location.hash.slice(1) || '/';
        const match = this.matchRoute(hash);

        if (!match) {
            this.viewContainer.innerHTML = '<div class="empty-state"><h2>Page not found</h2></div>';
            return;
        }

        const { route, params } = match;

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
        await route.handler(params);
    }

    getParams() {
        const hash = window.location.hash.slice(1) || '/';
        const parts = hash.split('/').filter(Boolean);
        return parts.slice(1);
    }
}

export const router = new Router();
