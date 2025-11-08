export function isValidFilename(filename) {
    if (!filename) return false;

    const trimmed = filename.trim();
    if (!trimmed.toLowerCase().endsWith('.md')) {
        return false;
    }

    const name = trimmed.slice(0, -3);
    if (!name) return false;

    const forbiddenPattern = /[\/\0]/;
    if (forbiddenPattern.test(name)) return false;

    if (name === '.' || name === '..') return false;

    if (name.includes('.') || name.includes('/') || name.includes(' ')) return false;

    return true;
}
