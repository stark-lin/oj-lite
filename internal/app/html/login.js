const { collectElementsById, parseJSONSafe, ui } = window.OJLite;

document.getElementById('loginFieldsMount').outerHTML = [
  ui.loginField({
    label: 'Username',
    inputAttrs: { id: 'username', name: 'username', autocomplete: 'username', autofocus: true }
  }),
  ui.loginField({
    label: 'Password',
    password: true,
    inputAttrs: { id: 'password', name: 'password', autocomplete: 'current-password' }
  })
].join('');

const {
  loginForm: form,
  username,
  password,
  submitBtn,
  notice
} = collectElementsById('loginForm', 'username', 'password', 'submitBtn', 'notice');

function setNotice(text, isError) {
  notice.textContent = text;
  notice.className = isError ? 'notice notice--error' : 'notice';
}

function setSubmitting(active) {
  submitBtn.disabled = active;
  submitBtn.textContent = active ? 'Signing in...' : 'Sign in';
}

form.addEventListener('submit', async (event) => {
  event.preventDefault();

  const user = username.value.trim();
  const pass = password.value.trim();

  if (!user || !pass) {
    setNotice('Username and password are required.', true);
    return;
  }

  setSubmitting(true);
  setNotice('');

  try {
    const loginResponse = await fetch('/api/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      credentials: 'include',
      body: JSON.stringify({
        username: user,
        password: pass
      })
    });

    const loginPayload = await parseJSONSafe(loginResponse);
    if (!loginResponse.ok) {
      const message = loginPayload?.error?.message || 'Incorrect username or password.';
      setNotice(message, true);
      return;
    }

    const meResponse = await fetch('/api/me', {
      method: 'GET',
      credentials: 'include'
    });

    const mePayload = await parseJSONSafe(meResponse);
    if (!meResponse.ok) {
      setNotice('Login succeeded, but the session cookie was not accepted.', true);
      return;
    }

    const me = mePayload?.data?.user;
    if (!me) {
      setNotice('Login succeeded.', false);
      return;
    }

    if (me.role === 'student') {
      setNotice('Login successful. Redirecting...', false);
      window.location.assign('/student');
      return;
    }

    if (me.role === 'teacher') {
      setNotice('Login successful. Redirecting...', false);
      window.location.assign('/teacher');
      return;
    }

    setNotice(`Login successful. Signed in as ${me.username} (${me.role}).`, false);
  } catch (error) {
    setNotice('Network error. Please try again.', true);
  } finally {
    setSubmitting(false);
  }
});
