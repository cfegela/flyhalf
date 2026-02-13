import { auth } from '../auth.js';
import { router } from '../router.js';

class Navbar {
    constructor() {
        this.container = document.getElementById('navbar');
        this.render();

        auth.subscribe(() => this.render());

        // Update active link whenever the route changes
        window.addEventListener('popstate', () => this.updateActiveLink());
        window.addEventListener('routechange', () => this.updateActiveLink());
    }

    async handleLogout(e) {
        e.preventDefault();
        try {
            await auth.logout();
            router.navigate('/login');
        } catch (error) {
            console.error('Logout failed:', error);
        }
    }

    render() {
        const user = auth.getUser();

        if (!auth.isAuthenticated()) {
            this.container.innerHTML = '';
            return;
        }

        const isAdmin = auth.isAdmin();

        this.container.innerHTML = `
            <div class="navbar">
                <a href="/" class="navbar-brand">Flyhalf</a>
                <button class="navbar-toggle" id="navbar-toggle" aria-label="Toggle navigation">
                    <span></span>
                    <span></span>
                    <span></span>
                </button>
                <div class="navbar-menu" id="navbar-menu">
                    <a href="/tickets" class="navbar-link">Tickets</a>
                    <a href="/sprints" class="navbar-link">Sprints</a>
                    <a href="/projects" class="navbar-link">Projects</a>
                    ${isAdmin ? '<a href="/admin/users" class="navbar-link">Users</a>' : ''}
                    ${isAdmin ? '<a href="/admin/teams" class="navbar-link">Teams</a>' : ''}
                    ${isAdmin ? '<a href="/admin/leagues" class="navbar-link">Leagues</a>' : ''}
                    <a href="/settings" class="navbar-link">${user.first_name} ${user.last_name}</a>
                    <button class="btn btn-secondary btn-sm" id="logout-btn">Logout</button>
                </div>
            </div>
        `;

        const navbarToggle = this.container.querySelector('#navbar-toggle');
        const navbarMenu = this.container.querySelector('#navbar-menu');

        if (navbarToggle && navbarMenu) {
            navbarToggle.addEventListener('click', () => {
                navbarMenu.classList.toggle('active');
                navbarToggle.classList.toggle('active');
            });

            // Close menu when clicking a link
            const navbarLinks = navbarMenu.querySelectorAll('.navbar-link');
            navbarLinks.forEach(link => {
                link.addEventListener('click', () => {
                    navbarMenu.classList.remove('active');
                    navbarToggle.classList.remove('active');
                });
            });

            // Close menu when clicking logout button
            const logoutBtn = this.container.querySelector('#logout-btn');
            if (logoutBtn) {
                logoutBtn.addEventListener('click', (e) => {
                    navbarMenu.classList.remove('active');
                    navbarToggle.classList.remove('active');
                    this.handleLogout(e);
                });
            }
        }

        this.updateActiveLink();
    }

    updateActiveLink() {
        const pathname = window.location.pathname || '/';
        const links = this.container.querySelectorAll('.navbar-link');

        links.forEach(link => {
            const href = link.getAttribute('href');

            // Check if current pathname matches this link
            // For home link, only match exact '/'
            // For other links, match if pathname starts with the link path
            const isActive = href === '/'
                ? pathname === '/'
                : pathname.startsWith(href);

            if (isActive) {
                link.classList.add('active');
            } else {
                link.classList.remove('active');
            }
        });
    }
}

export function initNavbar() {
    return new Navbar();
}
