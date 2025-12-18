const STORAGE_KEY = 'unauth_files';

export const UnauthStorage = {
  load() {
    try {
      return JSON.parse(localStorage.getItem(STORAGE_KEY)) || {};
    } catch {
      return {};
    }
  },

  save(files) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(files));
  },

  remove(name) {
    const files = this.load();
    delete files[name];
    this.save(files);
  }
};
