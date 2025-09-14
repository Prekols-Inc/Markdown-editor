import axios from 'axios';

const API = axios.create({
  baseURL: import.meta.env.VITE_AUTH_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export default API;
