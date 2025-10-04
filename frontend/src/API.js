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

export default { AUTH, STORAGE };

