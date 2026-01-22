import { auth } from '../auth.js';
import { router } from '../router.js';
import { toast } from './toast.js';

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
            toast.success('Logged out successfully');
            router.navigate('/login');
        } catch (error) {
            toast.error('Logout failed: ' + error.message);
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
                    <a href="#/" class="navbar-link">Dashboard</a>
                    <a href="#/resources" class="navbar-link">Resources</a>
                    ${isAdmin ? '<a href="#/admin/users" class="navbar-link">Users</a>' : ''}
                    <span class="navbar-link">${user.first_name} ${user.last_name}</span>
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
