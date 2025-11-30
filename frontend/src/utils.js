const INVALID = /[<>:"/\\|?*+,!%@]/g;
const RESERVED = /^(con|prn|aux|nul|com[1-9]|lpt[1-9])$/i;

export function validateFilename(name) {
    const result = { ok: true, code: null, message: null, details: null };

    if (!name || !name.trim()) {
        return { ok: false, code: 'FILE_NAME_EMPTY', message: 'Имя файла не может быть пустым.' };
    }
    if (name.length > 255) {
        return { ok: false, code: 'FILE_NAME_TOO_LONG', message: 'Слишком длинное имя файла (макс. 255).' };
    }
    if (name.endsWith('.') || name.endsWith(' ')) {
        return { ok: false, code: 'FILE_NAME_TRAILING', message: 'Имя не должно заканчиваться точкой или пробелом.' };
    }
    if (name.includes('/') || name.includes('\\')) {
        return { ok: false, code: 'FILE_NAME_PATH', message: 'Имя не должно содержать путь (/, \\).' };
    }
    const m = name.match(INVALID);
    if (m) {
        const uniq = [...new Set(m)];
        return {
        ok: false,
        code: 'FILE_NAME_INVALID_CHARS',
        message: `Недопустимые символы: ${uniq.join(' ')}`,
        details: { invalid: uniq }
        };
    }
    const dot = name.lastIndexOf('.');
    const base = dot === -1 ? name : name.slice(0, dot);
    const ext  = dot === -1 ? ''   : name.slice(dot).toLowerCase();
    if (/^\.+$/.test(base)) {
        return {
            ok: false,
            code: "FILE_NAME_ONLY_DOTS",
            message: "Имя файла не может состоять только из точек."
        };
    }

    if (/\.+$/.test(base)) {
        return {
        ok: false,
        code: 'FILE_NAME_TRAILING_DOTS',
        message: 'Имя файла не должно заканчиваться точками перед расширением.'
        };
    }

    if (!base.trim()) {
        return { ok: false, code: 'FILE_NAME_EMPTY_BASE', message: 'Имя файла не может быть пустым.' };
    }

    if (RESERVED.test(base)) {
        return { ok: false, code: 'FILE_NAME_RESERVED', message: 'Это имя зарезервировано системой.' };
    }

    if (ext !== '.md' && ext !== '.markdown') {
        return {
        ok: false,
        code: 'FILE_EXTENSION_INVALID',
        message: 'Разрешены только расширения: .md, .markdown',
        details: { allowedExtensions: ['.md', '.markdown'] }
        };
    }
    return result;
}

export function isValidFilename(name) {
    return validateFilename(name).ok;
}
