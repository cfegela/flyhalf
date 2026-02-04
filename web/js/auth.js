import { api } from './api.js';

class AuthManager {
    constructor() {
        this.currentUser = null;
        this.listeners = [];
        this.refreshTimer = null;
    }

    subscribe(listener) {
        this.listeners.push(listener);
        return () => {
            this.listeners = this.listeners.filter(l => l !== listener);
        };
    }

    notify() {
        this.listeners.forEach(listener => listener(this.currentUser));
    }

    async init() {
        try {
            await this.refreshToken();
        } catch (error) {
            console.log('No valid session found');
        }
    }

    async login(email, password) {
        const data = await api.login(email, password);
        this.currentUser = data.user;
        this.scheduleTokenRefresh();
        this.notify();
        return data;
    }

    async logout() {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
            this.refreshTimer = null;
        }

        try {
            await api.logout();
        } catch (error) {
            console.error('Logout error:', error);
        }

        this.currentUser = null;
        this.notify();
    }

    async refreshToken() {
        const data = await api.refresh();
        this.currentUser = data.user;
        this.scheduleTokenRefresh();
        this.notify();
        return data;
    }

    scheduleTokenRefresh() {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
        }

        // Refresh token at 11 minutes (well before 15 minute expiry)
        // This provides 4 minute buffer for network delays and processing
        this.refreshTimer = setTimeout(async () => {
            try {
                await this.refreshToken();
            } catch (error) {
                console.error('Token refresh failed:', error);
                this.currentUser = null;
                this.notify();
            }
        }, 11 * 60 * 1000);
    }

    isAuthenticated() {
        return this.currentUser !== null;
    }

    isAdmin() {
        return this.currentUser && this.currentUser.role === 'admin';
    }

    getUser() {
        return this.currentUser;
    }
}

export const auth = new AuthManager();
