import { config } from './config.js';

const API_BASE_URL = config.apiBaseUrl;

class APIClient {
    constructor() {
        this.accessToken = null;
    }

    setAccessToken(token) {
        this.accessToken = token;
    }

    clearAccessToken() {
        this.accessToken = null;
    }

    async request(endpoint, options = {}) {
        const url = `${API_BASE_URL}${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (this.accessToken) {
            headers['Authorization'] = `Bearer ${this.accessToken}`;
        }

        const config = {
            ...options,
            headers,
            credentials: 'include',
        };

        try {
            const response = await fetch(url, config);

            if (response.status === 204) {
                return null;
            }

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Request failed');
            }

            return data;
        } catch (error) {
            throw error;
        }
    }

    async login(email, password) {
        const data = await this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
        });
        this.setAccessToken(data.access_token);
        return data;
    }

    async refresh() {
        try {
            const data = await this.request('/auth/refresh', {
                method: 'POST',
            });
            this.setAccessToken(data.access_token);
            return data;
        } catch (error) {
            this.clearAccessToken();
            throw error;
        }
    }

    async logout() {
        try {
            await this.request('/auth/logout', {
                method: 'POST',
            });
        } finally {
            this.clearAccessToken();
        }
    }

    async getMe() {
        return this.request('/auth/me');
    }

    async changePassword(currentPassword, newPassword) {
        return this.request('/auth/password', {
            method: 'PUT',
            body: JSON.stringify({
                current_password: currentPassword,
                new_password: newPassword
            }),
        });
    }

    async getTickets() {
        return this.request('/tickets');
    }

    async getTicket(id) {
        return this.request(`/tickets/${id}`);
    }

    async createTicket(ticket) {
        return this.request('/tickets', {
            method: 'POST',
            body: JSON.stringify(ticket),
        });
    }

    async updateTicket(id, ticket) {
        return this.request(`/tickets/${id}`, {
            method: 'PUT',
            body: JSON.stringify(ticket),
        });
    }

    async deleteTicket(id) {
        return this.request(`/tickets/${id}`, {
            method: 'DELETE',
        });
    }

    async promoteTicket(id) {
        return this.request(`/tickets/${id}/promote`, {
            method: 'POST',
        });
    }

    async demoteTicket(id) {
        return this.request(`/tickets/${id}/demote`, {
            method: 'POST',
        });
    }

    async getEpics() {
        return this.request('/epics');
    }

    async getEpic(id) {
        return this.request(`/epics/${id}`);
    }

    async createEpic(epic) {
        return this.request('/epics', {
            method: 'POST',
            body: JSON.stringify(epic),
        });
    }

    async updateEpic(id, epic) {
        return this.request(`/epics/${id}`, {
            method: 'PUT',
            body: JSON.stringify(epic),
        });
    }

    async deleteEpic(id) {
        return this.request(`/epics/${id}`, {
            method: 'DELETE',
        });
    }

    async getSprints() {
        return this.request('/sprints');
    }

    async getSprint(id) {
        return this.request(`/sprints/${id}`);
    }

    async createSprint(sprint) {
        return this.request('/sprints', {
            method: 'POST',
            body: JSON.stringify(sprint),
        });
    }

    async updateSprint(id, sprint) {
        return this.request(`/sprints/${id}`, {
            method: 'PUT',
            body: JSON.stringify(sprint),
        });
    }

    async deleteSprint(id) {
        return this.request(`/sprints/${id}`, {
            method: 'DELETE',
        });
    }

    async getUsersForAssignment() {
        return this.request('/users');
    }

    async getUsers() {
        return this.request('/admin/users');
    }

    async getUser(id) {
        return this.request(`/admin/users/${id}`);
    }

    async createUser(user) {
        return this.request('/admin/users', {
            method: 'POST',
            body: JSON.stringify(user),
        });
    }

    async updateUser(id, user) {
        return this.request(`/admin/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(user),
        });
    }

    async deleteUser(id) {
        return this.request(`/admin/users/${id}`, {
            method: 'DELETE',
        });
    }
}

export const api = new APIClient();
