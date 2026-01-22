import { api } from '../api.js';
import { auth } from '../auth.js';
import { router } from '../router.js';

export async function forcePasswordChangeView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div class="login-container">
            <div class="login-card card">
                <div class="login-header">
                    <h1 class="login-title">Change Password Required</h1>
                    <p class="login-subtitle">You must change your password before continuing</p>
                </div>
                <form id="force-password-form">
                    <div class="form-group">
                        <label class="form-label" for="current_password">Current Password *</label>
                        <input
                            type="password"
                            id="current_password"
                            class="form-input"
                            required
                            autocomplete="current-password"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="new_password">New Password *</label>
                        <input
                            type="password"
                            id="new_password"
                            class="form-input"
                            required
                            minlength="8"
                            autocomplete="new-password"
                        >
                        <small style="color: var(--text-secondary);">Minimum 8 characters</small>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="confirm_password">Confirm New Password *</label>
                        <input
                            type="password"
                            id="confirm_password"
                            class="form-input"
                            required
                            minlength="8"
                            autocomplete="new-password"
                        >
                    </div>
                    <div id="error-message" class="form-error" style="display: none;"></div>
                    <button type="submit" class="btn btn-primary" style="width: 100%;">
                        Change Password
                    </button>
                </form>
            </div>
        </div>
    `;

    const form = container.querySelector('#force-password-form');
    const errorMessage = container.querySelector('#error-message');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        errorMessage.style.display = 'none';

        const currentPassword = form.current_password.value;
        const newPassword = form.new_password.value;
        const confirmPassword = form.confirm_password.value;

        if (newPassword !== confirmPassword) {
            errorMessage.textContent = 'New passwords do not match';
            errorMessage.style.display = 'block';
            return;
        }

        if (newPassword.length < 8) {
            errorMessage.textContent = 'Password must be at least 8 characters';
            errorMessage.style.display = 'block';
            return;
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = 'Changing...';

        try {
            await api.changePassword(currentPassword, newPassword);
            // Refresh user data to get updated must_change_password flag
            const userData = await api.getMe();
            auth.currentUser = userData;
            router.navigate('/tickets');
        } catch (error) {
            errorMessage.textContent = error.message || 'Failed to change password';
            errorMessage.style.display = 'block';
            submitBtn.disabled = false;
            submitBtn.textContent = 'Change Password';
        }
    });
}

export async function settingsView() {
    const container = document.getElementById('view-container');
    const user = auth.getUser();

    container.innerHTML = `
        <div>
            <div class="page-header">
                <h1 class="page-title">Settings</h1>
            </div>

            <div class="card">
                <div class="card-header">Account Information</div>
                <div style="display: grid; gap: 1rem;">
                    <div>
                        <label class="form-label">Name</label>
                        <p>${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</p>
                    </div>
                    <div>
                        <label class="form-label">Email</label>
                        <p>${escapeHtml(user.email)}</p>
                    </div>
                    <div>
                        <label class="form-label">Role</label>
                        <p>${escapeHtml(user.role)}</p>
                    </div>
                </div>
            </div>

            <div class="card">
                <div class="card-header">Change Password</div>
                <form id="change-password-form">
                    <div class="form-group">
                        <label class="form-label" for="current_password">Current Password *</label>
                        <input
                            type="password"
                            id="current_password"
                            class="form-input"
                            required
                            autocomplete="current-password"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="new_password">New Password *</label>
                        <input
                            type="password"
                            id="new_password"
                            class="form-input"
                            required
                            minlength="8"
                            autocomplete="new-password"
                        >
                        <small style="color: var(--text-secondary);">Minimum 8 characters</small>
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="confirm_password">Confirm New Password *</label>
                        <input
                            type="password"
                            id="confirm_password"
                            class="form-input"
                            required
                            minlength="8"
                            autocomplete="new-password"
                        >
                    </div>
                    <div id="error-message" class="form-error" style="display: none;"></div>
                    <div id="success-message" style="color: var(--success); font-size: 0.875rem; margin-top: 0.5rem; display: none;">
                        Password changed successfully
                    </div>
                    <div style="display: flex; gap: 1rem; margin-top: 1rem;">
                        <button type="submit" class="btn btn-primary">
                            Change Password
                        </button>
                    </div>
                </form>
            </div>
        </div>
    `;

    const form = container.querySelector('#change-password-form');
    const errorMessage = container.querySelector('#error-message');
    const successMessage = container.querySelector('#success-message');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        errorMessage.style.display = 'none';
        successMessage.style.display = 'none';

        const currentPassword = form.current_password.value;
        const newPassword = form.new_password.value;
        const confirmPassword = form.confirm_password.value;

        if (newPassword !== confirmPassword) {
            errorMessage.textContent = 'New passwords do not match';
            errorMessage.style.display = 'block';
            return;
        }

        if (newPassword.length < 8) {
            errorMessage.textContent = 'Password must be at least 8 characters';
            errorMessage.style.display = 'block';
            return;
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = 'Changing...';

        try {
            await api.changePassword(currentPassword, newPassword);
            successMessage.style.display = 'block';
            form.reset();
        } catch (error) {
            errorMessage.textContent = error.message || 'Failed to change password';
            errorMessage.style.display = 'block';
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Change Password';
        }
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
