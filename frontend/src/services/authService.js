import API from '../API';

export async function hashPassword(password, salt = '') {
  const enc = new TextEncoder();
  const data = enc.encode(password + '|' + salt);
  const digest = await crypto.subtle.digest('SHA-256', data);
  const bytes = Array.from(new Uint8Array(digest));
  return bytes.map(b => b.toString(16).padStart(2, '0')).join('');
}

const LOCAL_KEY = 'markdown_editor_local_users';

function getUsers() {
  try {
    const raw = localStorage.getItem(LOCAL_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveUsers(users) {
  localStorage.setItem(LOCAL_KEY, JSON.stringify(users));
}

export function isRemoteAuthConfigured() {
  return Boolean(import.meta.env.VITE_AUTH_API_BASE_URL);
}

export async function registerLocal({ username, password, email = '' }) {
  const users = getUsers();
  const exists = users.some(u => u.username.toLowerCase() === String(username).toLowerCase());
  if (exists) {
    throw new Error('Пользователь с таким именем уже существует');
  }
  const salt = String(Date.now());
  const passwordHash = await hashPassword(password, salt);
  users.push({ username, email, passwordHash, salt, createdAt: new Date().toISOString() });
  saveUsers(users);
  return { ok: true };
}

export async function loginLocal({ username, password }) {
  const users = getUsers();
  const user = users.find(u => u.username.toLowerCase() === String(username).toLowerCase());
  if (!user) return { ok: false, message: 'Пользователь не найден' };
  const candidate = await hashPassword(password, user.salt);
  if (candidate !== user.passwordHash) return { ok: false, message: 'Неверный пароль' };
  return { ok: true };
}

export async function register({ username, password, email = '' }) {
  if (isRemoteAuthConfigured()) {
    try {
      const resp = await API.AUTH.post('/v1/register', { username, password, email });
      if (resp.status >= 200 && resp.status < 300) return { ok: true, remote: true };
    } catch {
    }
  }
  await registerLocal({ username, password, email });
  return { ok: true, remote: false };
}

export async function login({ username, password }) {
  if (isRemoteAuthConfigured()) {
    try {
      const resp = await API.AUTH.post('/v1/login', { username, password });
      if (resp.status >= 200 && resp.status < 300) return { ok: true, remote: true };
    } catch {
    }
  }
  const res = await loginLocal({ username, password });
  return { ok: res.ok, remote: false, message: res.message };
}