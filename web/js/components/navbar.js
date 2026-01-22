import { auth } from '../auth.js';
import { router } from '../router.js';

class Navbar {
    constructor() {
        this.container = document.getElementById('navbar');
        this.render();

        auth.subscribe(() => this.render());
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
                <a href="#/" class="navbar-brand">Flyhalf</a>
                <div class="navbar-menu">
                    <a href="#/tickets" class="navbar-link">Tickets</a>
                    ${isAdmin ? '<a href="#/admin/users" class="navbar-link">Users</a>' : ''}
                    <a href="#/settings" class="navbar-link">${user.first_name} ${user.last_name}</a>
                    <button class="btn btn-secondary btn-sm" id="logout-btn">Logout</button>
                </div>
            </div>
        `;

        const logoutBtn = this.container.querySelector('#logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', (e) => this.handleLogout(e));
        }

        this.updateActiveLink();
    }

    updateActiveLink() {
        const hash = window.location.hash || '#/';
        const links = this.container.querySelectorAll('.navbar-link');

        links.forEach(link => {
            const href = link.getAttribute('href');
            if (href === hash || (hash === '#/' && href === '#/')) {
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
