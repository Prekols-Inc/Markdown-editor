import axios from 'axios';

const AUTH = axios.create({
  baseURL: import.meta.env.VITE_AUTH_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

const STORAGE = axios.create({
  baseURL: `${import.meta.env.VITE_STORAGE_API_BASE_URL}/api`,
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
      if (errorMessage && errorMessage === "Token has expired") {
        originalRequest._retry = true;
        try {
          const refreshResponse = await AUTH.post('/v1/refresh', {
            refreshToken: localStorage.getItem('refresh_token')
          });

          return AUTH(originalRequest);
        } catch (refreshError) {

          return Promise.reject(refreshError);
        }
      }
    }
    return Promise.reject(error);
  }
);

export default { AUTH, STORAGE };

