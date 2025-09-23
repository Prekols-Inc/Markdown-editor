import axios from 'axios';

const getRuntimeConfig = (key) => {
  if (window.APP_CONFIG && window.APP_CONFIG[key] !== undefined) {
    return window.APP_CONFIG[key];
  }

  return import.meta.env[key];
};

const createApiInstance = (envKey) => {
  let instance = null;
  
  return new Proxy({}, {
    get(target, prop) {
      if (!instance) {
        instance = axios.create({
          baseURL: getRuntimeConfig(envKey),
          headers: {
            'Content-Type': 'application/json',
          },
        });
      }
      
      if (instance[prop] !== undefined) {
        return typeof instance[prop] === 'function' 
          ? instance[prop].bind(instance) 
          : instance[prop];
      }
      
      return target[prop];
    },
    
    set(target, prop, value) {
      if (!instance) {
        instance = axios.create({
          baseURL: getRuntimeConfig(envKey),
          headers: {
            'Content-Type': 'application/json',
          },
        });
      }
      instance[prop] = value;
      return true;
    }
  });
};

const AUTH = createApiInstance('VITE_AUTH_API_BASE_URL');
const STORAGE = createApiInstance('VITE_STORAGE_API_BASE_URL');

export default { AUTH, STORAGE };
