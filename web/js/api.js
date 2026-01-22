const API_BASE_URL = 'http://localhost:8081/api/v1';

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

    async getResources() {
        return this.request('/resources');
    }

    async getResource(id) {
        return this.request(`/resources/${id}`);
    }

    async createResource(resource) {
        return this.request('/resources', {
            method: 'POST',
            body: JSON.stringify(resource),
        });
    }

    async updateResource(id, resource) {
        return this.request(`/resources/${id}`, {
            method: 'PUT',
            body: JSON.stringify(resource),
        });
    }

    async deleteResource(id) {
        return this.request(`/resources/${id}`, {
            method: 'DELETE',
        });
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
