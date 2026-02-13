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

    async updateTicketPriority(id, priority) {
        return this.request(`/tickets/${id}/priority`, {
            method: 'PATCH',
            body: JSON.stringify({ priority }),
        });
    }

    async updateTicketSprintOrder(id, sprintOrder) {
        return this.request(`/tickets/${id}/sprint-order`, {
            method: 'PATCH',
            body: JSON.stringify({ sprint_order: sprintOrder }),
        });
    }

    async updateAcceptanceCriteriaCompleted(ticketId, criteriaId, completed) {
        return this.request(`/tickets/${ticketId}/acceptance-criteria/${criteriaId}`, {
            method: 'PATCH',
            body: JSON.stringify({ completed }),
        });
    }

    async createTicketUpdate(ticketId, content) {
        return this.request(`/tickets/${ticketId}/updates`, {
            method: 'POST',
            body: JSON.stringify({ content }),
        });
    }

    async deleteAcceptanceCriteria(ticketId, criteriaId) {
        return this.request(`/tickets/${ticketId}/acceptance-criteria/${criteriaId}`, {
            method: 'DELETE',
        });
    }

    async deleteTicketUpdate(ticketId, updateId) {
        return this.request(`/tickets/${ticketId}/updates/${updateId}`, {
            method: 'DELETE',
        });
    }

    async getProjects() {
        return this.request('/projects');
    }

    async getProject(id) {
        return this.request(`/projects/${id}`);
    }

    async createProject(project) {
        return this.request('/projects', {
            method: 'POST',
            body: JSON.stringify(project),
        });
    }

    async updateProject(id, project) {
        return this.request(`/projects/${id}`, {
            method: 'PUT',
            body: JSON.stringify(project),
        });
    }

    async deleteProject(id) {
        return this.request(`/projects/${id}`, {
            method: 'DELETE',
        });
    }

    async getSprints() {
        return this.request('/sprints');
    }

    async getSprint(id) {
        return this.request(`/sprints/${id}`);
    }

    async getSprintTickets(id) {
        return this.request(`/sprints/${id}/tickets`);
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

    async closeSprint(id) {
        return this.request(`/sprints/${id}/close`, {
            method: 'POST',
        });
    }

    async getSprintReport(id) {
        return this.request(`/sprints/${id}/report`);
    }

    async getRetroItems(sprintId) {
        return this.request(`/sprints/${sprintId}/retro`);
    }

    async createRetroItem(sprintId, content, category) {
        return this.request(`/sprints/${sprintId}/retro`, {
            method: 'POST',
            body: JSON.stringify({ content, category }),
        });
    }

    async updateRetroItem(id, content, category) {
        return this.request(`/retro-items/${id}`, {
            method: 'PUT',
            body: JSON.stringify({ content, category }),
        });
    }

    async deleteRetroItem(id) {
        return this.request(`/retro-items/${id}`, {
            method: 'DELETE',
        });
    }

    async voteRetroItem(id) {
        return this.request(`/retro-items/${id}/vote`, {
            method: 'POST',
        });
    }

    async unvoteRetroItem(id) {
        return this.request(`/retro-items/${id}/vote`, {
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

    async getTeams() {
        return this.request('/admin/teams');
    }

    async getTeam(id) {
        return this.request(`/admin/teams/${id}`);
    }

    async createTeam(team) {
        return this.request('/admin/teams', {
            method: 'POST',
            body: JSON.stringify(team),
        });
    }

    async updateTeam(id, team) {
        return this.request(`/admin/teams/${id}`, {
            method: 'PUT',
            body: JSON.stringify(team),
        });
    }

    async deleteTeam(id) {
        return this.request(`/admin/teams/${id}`, {
            method: 'DELETE',
        });
    }

    async getLeagues() {
        return this.request('/admin/leagues');
    }

    async getLeague(id) {
        return this.request(`/admin/leagues/${id}`);
    }

    async createLeague(league) {
        return this.request('/admin/leagues', {
            method: 'POST',
            body: JSON.stringify(league),
        });
    }

    async updateLeague(id, league) {
        return this.request(`/admin/leagues/${id}`, {
            method: 'PUT',
            body: JSON.stringify(league),
        });
    }

    async deleteLeague(id) {
        return this.request(`/admin/leagues/${id}`, {
            method: 'DELETE',
        });
    }

    async resetDemo() {
        return this.request('/admin/reset-demo', {
            method: 'POST',
        });
    }

    async reseedDemo() {
        return this.request('/admin/reseed-demo', {
            method: 'POST',
        });
    }
}

export const api = new APIClient();
