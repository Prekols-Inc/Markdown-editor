export function isValidFilename(filename) {
    if (!filename || !filename.toLowerCase().endsWith('.md')) {
      return false;
    }

    const name = filename.slice(0, -3);
    if (!name.trim()) {
        return false;
    }
    
    const forbiddenPattern = /[\/\\:*?"<>|]/;
    if (forbiddenPattern.test(name)) {
      return false;
    }
  
    if (name.includes('..') || name.includes('.')) {
      return false;
    }  
    return true;
  }
  