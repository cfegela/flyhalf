import { auth } from './auth.js';

class Router {
    constructor() {
        this.routes = [];
        this.currentRoute = null;
        this.viewContainer = document.getElementById('view-container');

        window.addEventListener('popstate', () => this.navigate());

        // Handle link clicks to prevent full page reloads
        document.addEventListener('click', (e) => {
            if (e.target.matches('a[href^="/"]') || e.target.closest('a[href^="/"]')) {
                const link = e.target.matches('a') ? e.target : e.target.closest('a');
                e.preventDefault();
                this.navigate(link.getAttribute('href'));
            }
        });
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
            window.history.pushState(null, '', path);
            // Trigger navigation after updating URL
            await this.navigate();
            return;
        }

        const pathname = window.location.pathname || '/';
        const match = this.matchRoute(pathname);

        if (!match) {
            this.viewContainer.innerHTML = '<div class="empty-state"><h2>Page not found</h2></div>';
            return;
        }

        const { route, params } = match;

        if (route.options.requireAuth && !auth.isAuthenticated()) {
            this.navigate('/login');
            return;
        }

        // Check if user must change password (unless already on password change page)
        if (route.options.requireAuth && auth.isAuthenticated() && !route.options.allowPasswordChange) {
            const user = auth.getUser();
            if (user && user.must_change_password) {
                this.navigate('/force-password-change');
                return;
            }
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

        // Dispatch custom event for components to react to navigation
        window.dispatchEvent(new CustomEvent('routechange'));
    }

    getParams() {
        const pathname = window.location.pathname || '/';
        const parts = pathname.split('/').filter(Boolean);
        return parts.slice(1);
    }
}

export const router = new Router();
