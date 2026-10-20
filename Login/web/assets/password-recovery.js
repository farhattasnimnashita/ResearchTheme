'use strict';
// Uses existing ResearchBridge API endpoints and Supabase OTP-based recovery backend.
const isReset = location.pathname === '/reset-password';
const root = document.getElementById('recovery-content');
const qs = new URLSearchParams(location.search);
// The email is passed as a query parameter for convenience; no password or OTP is included.
const suggestedEmail = qs.get('email') || '';
root.innerHTML = isReset ? `
<p class="eyebrow">ACCOUNT RECOVERY · STEP 2</p>
<h1>Create a new password</h1>
<p class="subtle">Enter the recovery code sent to your email, then choose a new password.</p>
<div class="hint">✉ <span>Check your inbox and Spam. Use the <strong>recovery OTP</strong> in the email, not a previous signup verification code.</span></div>
<form id="recovery-form">
<label class="field">Registered email<input type="email" name="email" maxlength="254" autocomplete="email" required></label>
<label class="field">Recovery code (OTP)<input type="text" name="token" inputmode="numeric" pattern="[0-9]{6,10}" maxlength="10" autocomplete="one-time-code" placeholder="6-digit code" required></label>
<label class="field">New password<input type="password" name="password" minlength="8" maxlength="72" autocomplete="new-password" required></label>
<label class="field">Confirm new password<input type="password" name="confirm" minlength="8" maxlength="72" autocomplete="new-password" required></label>
<button class="button" type="submit">Reset password →</button>
</form>
<p id="recovery-status" class="status" role="status" aria-live="polite"></p>
<a href="/forgot-password" class="extra">Didn't get a code? Request another</a>
<a href="/" id="login-after-reset" class="button secondary" style="display:none;text-decoration:none;text-align:center;padding:16px;margin-top:18px">Return to login →</a>
` : `
<p class="eyebrow">ACCOUNT RECOVERY · STEP 1</p>
<h1>Forgot your password?</h1>
<p class="subtle">Enter your registered email to receive a Supabase recovery code. Already received one? Continue directly to reset your password.</p>
<form id="recovery-form">
<label class="field">Email address<input type="email" name="email" maxlength="254" autocomplete="email" placeholder="name@example.com" required></label>
<button class="button" type="submit">✉ Send recovery code →</button>
</form>
<p id="recovery-status" class="status" role="status" aria-live="polite"></p>
<div class="separator">Already have a code?</div>
<a href="/reset-password" id="go-reset" class="button secondary" style="display:block;text-decoration:none;text-align:center;padding:16px">Enter recovery code →</a>
`;
document.getElementById(isReset ? 'step-reset' : 'step-email').classList.add('active');
const form = document.getElementById('recovery-form');
const status = document.getElementById('recovery-status');
if (isReset && suggestedEmail && suggestedEmail.length <= 254) form.elements.email.value = suggestedEmail;
function message(text, type) { status.textContent = text; status.className = 'status ' + (type || ''); }
form.addEventListener('submit', async (event) => {
  event.preventDefault();
  const values = Object.fromEntries(new FormData(form).entries());
  if (isReset && values.password !== values.confirm) return message('Passwords do not match.', 'error');
  if (isReset && new TextEncoder().encode(values.password).length > 72) return message('Password must not exceed 72 bytes.', 'error');
  const button = form.querySelector('button[type="submit"]');
  button.disabled = true;
  message(isReset ? 'Updating password…' : 'Sending recovery code…');
  try {
    const payload = isReset ? { email: values.email, token: values.token.trim(), password: values.password } : { email: values.email };
    const response = await fetch(isReset ? '/api/reset-password' : '/api/forgot-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'ResearchBridge' },
      credentials: 'same-origin', body: JSON.stringify(payload)
    });
    const result = await response.json();
    if (!response.ok) throw new Error(result.error || 'Request failed. Please try again.');
    message(result.message || (isReset ? 'Password updated.' : 'Recovery requested. Check your email.'), 'success');
    if (isReset) {
      form.reset(); form.hidden = true;
      document.getElementById('login-after-reset').style.display = 'block';
    } else {
      // Keep the next step prominent after a successful send. No OTP is stored.
      const next = document.getElementById('go-reset');
      next.href = '/reset-password?email=' + encodeURIComponent(values.email.trim());
      next.textContent = 'Continue to enter your OTP →';
      button.disabled = false;
    }
  } catch (error) {
    message(error.message || 'Unexpected error.', 'error'); button.disabled = false;
  }
});
