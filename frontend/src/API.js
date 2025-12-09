import axios from 'axios';

const AUTH = axios.create({
  baseURL: import.meta.env.VITE_AUTH_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

const REFRESH = axios.create({
  baseURL: import.meta.env.VITE_AUTH_API_BASE_URL,
  withCredentials: true,
  headers: { 
    'Content-Type': 'application/json',
  },
});

const STORAGE = axios.create({
  baseURL: `${import.meta.env.VITE_STORAGE_API_BASE_URL}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

const GIGACHAT_PROXY = axios.create({
  baseURL: import.meta.env.VITE_GIGACHAT_PROXY_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

AUTH.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      const errorMessage = error.response.data?.error;
      
      if (errorMessage && 
        (errorMessage === "Token has expired" || errorMessage === "Missing access token")) {
        originalRequest._retry = true;
        try {
          const refreshResponse = await REFRESH.post('/v1/refresh');

          return AUTH(originalRequest);
        } catch (refreshError) {

          return Promise.reject(refreshError);
        }
      }
    }
    return Promise.reject(error);
  }
);

STORAGE.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      const errorMessage = error.response.data?.error;
      if (errorMessage && 
        (errorMessage === "JWT not provided")) {
        originalRequest._retry = true;
        try {
          const refreshResponse = await REFRESH.post('/v1/refresh');

          return STORAGE(originalRequest);
        } catch (refreshError) {

          return Promise.reject(refreshError);
        }
      }
    }
    return Promise.reject(error);
  }
);

export function parseAPIError(e) {
  const data = e?.response?.data;
  const raw = data?.error;

  if (!raw) {
    return { code: 'GENERIC', message: e?.message || 'Ошибка сети' };
  }
  if (typeof raw === 'string') {
    return { code: 'GENERIC', message: raw };
  }
  const code = typeof raw.code === 'string' ? raw.code : 'GENERIC';
  let message = raw.message;

  if (typeof message !== 'string') {
    try { message = JSON.stringify(raw); }
    catch { message = 'Произошла ошибка'; }
  }
  return { code, message, field: raw.field, details: raw.details };
}
export default { AUTH, STORAGE, GIGACHAT_PROXY };

