import { auth } from '../auth.js';
import { router } from '../router.js';

export async function loginView() {
    const container = document.getElementById('view-container');

    container.innerHTML = `
        <div class="login-container">
            <div class="login-card card">
                <div class="login-header">
                    <h1 class="login-title">Flyhalf</h1>
                    <p class="login-subtitle">Sign in to your account</p>
                </div>
                <form id="login-form">
                    <div class="form-group">
                        <label class="form-label" for="email">Email</label>
                        <input
                            type="email"
                            id="email"
                            class="form-input"
                            required
                            autocomplete="email"
                            placeholder="you@example.com"
                        >
                    </div>
                    <div class="form-group">
                        <label class="form-label" for="password">Password</label>
                        <input
                            type="password"
                            id="password"
                            class="form-input"
                            required
                            autocomplete="current-password"
                            placeholder="••••••••"
                        >
                    </div>
                    <div id="error-message" class="form-error" style="display: none;"></div>
                    <button type="submit" class="btn btn-primary" style="width: 100%;">
                        Sign In
                    </button>
                </form>
            </div>
        </div>
    `;

    const form = container.querySelector('#login-form');
    const errorMessage = container.querySelector('#error-message');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        errorMessage.style.display = 'none';

        const email = form.email.value.trim();
        const password = form.password.value;

        if (!email || !password) {
            return;
        }

        const submitBtn = form.querySelector('button[type="submit"]');
        submitBtn.disabled = true;
        submitBtn.textContent = 'Signing in...';

        try {
            await auth.login(email, password);
            const user = auth.getUser();
            if (user && user.must_change_password) {
                router.navigate('/force-password-change');
            } else {
                router.navigate('/');
            }
        } catch (error) {
            console.error('Login failed:', error);

            // Display user-friendly error message
            let errorText = 'Login failed. Please try again.';

            if (error.message) {
                const msg = error.message.toLowerCase();
                if (msg.includes('invalid credentials') || msg.includes('unauthorized')) {
                    errorText = 'Invalid email or password. Please check your credentials and try again.';
                } else if (msg.includes('network') || msg.includes('fetch')) {
                    errorText = 'Unable to connect to server. Please check your connection and try again.';
                } else if (msg.includes('inactive') || msg.includes('disabled')) {
                    errorText = 'Your account has been deactivated. Please contact your administrator.';
                } else {
                    errorText = error.message;
                }
            }

            errorMessage.textContent = errorText;
            errorMessage.style.display = 'block';
            submitBtn.disabled = false;
            submitBtn.textContent = 'Sign In';
        }
    });
}
