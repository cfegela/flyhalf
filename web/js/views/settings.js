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

            <!-- Account Information Card -->
            <div class="card">
                <h2 class="card-header">Account Information</h2>
                <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1.5rem;">
                    <div>
                        <label class="form-label">Full Name</label>
                        <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary); font-weight: 500;">
                            ${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}
                        </p>
                    </div>
                    <div>
                        <label class="form-label">Email Address</label>
                        <p style="margin-top: 0.25rem; font-size: 1rem; color: var(--text-primary);">
                            ${escapeHtml(user.email)}
                        </p>
                    </div>
                    <div>
                        <label class="form-label">Role</label>
                        <div style="margin-top: 0.25rem;">
                            <span class="badge ${user.role === 'admin' ? 'badge-primary' : 'badge-success'}" style="font-size: 0.875rem; padding: 0.375rem 0.875rem;">
                                ${escapeHtml(user.role)}
                            </span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Security Card -->
            <div class="card">
                <h2 class="card-header">Security</h2>
                <p style="color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.6;">
                    Keep your account secure by using a strong password. We recommend changing your password regularly.
                </p>
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

            ${auth.isAdmin() ? `
            <!-- Danger Zone Card (Admin Only) -->
            <div class="card">
                <h2 class="card-header">Danger Zone</h2>
                <p style="color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.6;">
                    These actions are irreversible and will permanently delete data from the system.
                </p>
                <div style="margin-bottom: 1.5rem;">
                    <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Reset Demo Environment</h3>
                    <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                        This will delete ALL tickets, sprints, and projects. This action cannot be undone.
                    </p>
                    <button type="button" class="btn btn-danger" id="reset-demo-btn">
                        Reset Demo Environment
                    </button>
                </div>
                <div>
                    <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem; color: var(--text-primary);">Reseed Demo Environment</h3>
                    <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.875rem;">
                        This will create 1 sprint, 1 project, and 5 sample tickets for demonstration purposes.
                    </p>
                    <button type="button" class="btn btn-primary" id="reseed-demo-btn">
                        Reseed Demo Environment
                    </button>
                </div>
            </div>
            ` : ''}
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

    // Reset demo button (admin only)
    const resetDemoBtn = container.querySelector('#reset-demo-btn');
    if (resetDemoBtn) {
        resetDemoBtn.addEventListener('click', async () => {
            if (!confirm('Are you sure you want to reset the demo environment? This will delete ALL tickets, sprints, and projects. This action cannot be undone.')) {
                return;
            }

            resetDemoBtn.disabled = true;
            resetDemoBtn.textContent = 'Resetting...';

            try {
                const result = await api.resetDemo();
                alert(`Demo environment reset successfully.\n\nDeleted:\n- ${result.tickets_deleted} tickets\n- ${result.sprints_deleted} sprints\n- ${result.projects_deleted} projects`);
                router.navigate('/tickets');
            } catch (error) {
                alert('Failed to reset demo environment: ' + (error.message || 'Unknown error'));
                resetDemoBtn.disabled = false;
                resetDemoBtn.textContent = 'Reset Demo Environment';
            }
        });
    }

    // Reseed demo button (admin only)
    const reseedDemoBtn = container.querySelector('#reseed-demo-btn');
    if (reseedDemoBtn) {
        reseedDemoBtn.addEventListener('click', async () => {
            if (!confirm('Are you sure you want to reseed the demo environment? This will create 1 sprint, 1 project, and 5 sample tickets.')) {
                return;
            }

            reseedDemoBtn.disabled = true;
            reseedDemoBtn.textContent = 'Reseeding...';

            try {
                const result = await api.reseedDemo();
                alert(`Demo environment reseeded successfully.\n\nCreated:\n- ${result.tickets_created} tickets\n- ${result.sprints_created} sprint\n- ${result.projects_created} project`);
                router.navigate('/tickets');
            } catch (error) {
                alert('Failed to reseed demo environment: ' + (error.message || 'Unknown error'));
                reseedDemoBtn.disabled = false;
                reseedDemoBtn.textContent = 'Reseed Demo Environment';
            }
        });
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
