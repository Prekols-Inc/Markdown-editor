import { useState, useEffect, useCallback, useRef } from 'react';
import MarkdownEditor from './MarkdownEditor';
import FileSidebar from './FileSidebar';
import MarkdownPreview from './MarkdownPreview';
import AppTopBar from './AppTopBar';
import API from '../API';
import NewFileModal from './NewFileModal';
import { validateFilename } from "../utils";
import { toast, Toaster } from 'react-hot-toast';

export const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.
`;

const DEFAULT_LEFT = Math.round(window.innerWidth * 0.4);

const DEFAULT_OPTIONS = {
    breaks: false,
    gfm: true,
    pedantic: false,
    silent: false
};

export default function App() {
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [showPreview, setShowPreview] = useState(true);
    const toggleSidebar = () => setSidebarOpen(o => !o);
    const [leftWidth, setLeftWidth] = useState(DEFAULT_LEFT);
    const isResizing = useRef(false);
    const [markdown, setMarkdown] = useState(
        () => localStorage.getItem('md-draft') ?? DEFAULT_MD
    );

    const autoSaveTimeout = useRef(null);
    useEffect(() => {
        if (!fileHandle) return;

        if (autoSaveTimeout.current) {
            clearTimeout(autoSaveTimeout.current);
        }

        // Auto-save after 10 seconds of inactivity
        autoSaveTimeout.current = setTimeout(() => {
            handleSave();
        }, 10000);

        return () => clearTimeout(autoSaveTimeout.current);
    }, [markdown]);

    const parseAPIError =
        (API && API.parseAPIError)
            ? API.parseAPIError
            : (e) => {
                const data = e?.response?.data;
                const err = data?.error;
                if (!err) return { code: 'GENERIC', message: e?.message || 'Ошибка сети' };
                if (typeof err === 'string') return { code: 'GENERIC', message: err };
                return { code: err.code || 'GENERIC', message: err.message || 'Ошибка', field: err.field, details: err.details };
            };

    useEffect(() => {
        const id = setTimeout(() => localStorage.setItem('md-draft', markdown), 400);
        return () => clearTimeout(id);
    }, [markdown]);

    const handleMouseDown = () => (isResizing.current = true);

    useEffect(() => {
        const handleMouseMove = e => {
            if (!isResizing.current) return;
            const sidebar = sidebarOpen ? 260 : 0;
            const min = 220;
            const max = window.innerWidth - sidebar - 220;
            const next = Math.min(Math.max(e.clientX - sidebar, min), max);
            setLeftWidth(next);
        };
        const stop = () => (isResizing.current = false);

        window.addEventListener('mousemove', handleMouseMove);
        window.addEventListener('mouseup', stop);
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            window.removeEventListener('mouseup', stop);
        };
    }, [sidebarOpen]);

    const [options, setOptions] = useState(() => {
        try {
            const stored = JSON.parse(localStorage.getItem('md-options'));
            if (!stored || typeof stored !== 'object' || !Object.keys(stored).length) {
                return DEFAULT_OPTIONS;
            }
            return stored;
        } catch {
            return DEFAULT_OPTIONS;
        }
    });

    const handleOptionsChange = useCallback((obj) => {
        setOptions(obj);
        localStorage.setItem('md-options', JSON.stringify(obj));
    }, []);

    const [fileHandle, setFileHandle] = useState(null);
    const [unsaved, setUnsaved] = useState(false);
    // Used to show/hide the unsaved dot
    const [savedSnapshot, setSavedSnapshot] = useState("");

    const sidebarRef = useRef(null);

    useEffect(() => {
        if (!fileHandle) {
            setUnsaved(false);
            return;
        }
        setUnsaved(markdown !== savedSnapshot);
    }, [markdown, fileHandle, savedSnapshot]);

    const handleOpenFile = useCallback((text, handle) => {
        setMarkdown(text);
        setFileHandle(handle);
        setSavedSnapshot(text);
        setUnsaved(false);
    }, []);

    const [isNewFileModalOpen, setIsNewFileModalOpen] = useState(false);

    const handleNewFile = useCallback(async (inputName) => {
        try {
            let filename = inputName?.trim() || 'untitled.md';
            if (!filename.endsWith('.md') && !filename.endsWith('.markdown')) {
                filename += '.md';
            }

            const v = validateFilename(filename);
            if (!v.ok) {
                toast.error(v.message);
                return;
            }

            const blob = new Blob([DEFAULT_MD], { type: 'text/plain' });
            const formData = new FormData();
            formData.append('file', blob, filename);

            await API.STORAGE.post(`/file/${encodeURIComponent(filename)}`, formData, {
                headers: { 'Content-Type': 'multipart/form-data' },
            });

            setMarkdown(DEFAULT_MD);
            setFileHandle({ name: filename });
            setSavedSnapshot(DEFAULT_MD);
            setUnsaved(false);

            sidebarRef.current?.refresh?.();

            toast.success('Файл создан');
        } catch (err) {
            console.error('Ошибка создания файла', err);
            const e = parseAPIError(err);
            if (e.code === 'FILE_ALREADY_EXISTS') {
                toast.error('Файл с таким именем уже существует. Выберите другое имя.');
            } else if (e.code === 'FILE_COUNT_LIMIT') {
                toast.error('Превышен лимит количества файлов. Удалите лишние.');
            } else if (e.code === 'USER_SPACE_FULL') {
                toast.error('Недостаточно места в хранилище пользователя.');
            } else if (e.code === 'FILE_NAME_INVALID_CHARS' && e.details?.invalid?.length) {
                toast.error(`Недопустимые символы: ${e.details.invalid.join(' ')}`);
            } else {
                toast.error(e.message || 'Не удалось создать файл');
            }
        }
    }, [toast, parseAPIError]);

    const handleFileUpload = useCallback(async (content, originalFilename) => {
        try {
            let filename = originalFilename || 'uploaded.md';

            // Ensure the file has the correct extension
            if (!filename.endsWith('.md') && !filename.endsWith('.markdown')) {
                filename = filename.replace(/\.[^/.]+$/, "") + '.md';
            }

            const v = validateFilename(filename);
            if (!v.ok) {
                toast.error(v.message);
                return;
            }

            const blob = new Blob([content], { type: 'text/plain' });
            const formData = new FormData();
            formData.append('file', blob, filename);

            await API.STORAGE.post(`/file/${encodeURIComponent(filename)}`, formData, {
                headers: { 'Content-Type': 'multipart/form-data' },
            });

            setMarkdown(content);
            setFileHandle({ name: filename });
            setSavedSnapshot(content);
            setUnsaved(false);

            sidebarRef.current?.refresh?.();

            toast.success(`Файл "${filename}" создан из загруженного файла`);
        } catch (err) {
            console.error('Ошибка создания файла из загруженного', err);
            const e = parseAPIError(err);
            if (e.code === 'FILE_ALREADY_EXISTS') {
                toast.error('Файл с таким именем уже существует. Переименуйте загружаемый файл.');
            } else if (e.code === 'FILE_COUNT_LIMIT') {
                toast.error('Превышен лимит количества файлов. Удалите лишние.');
            } else if (e.code === 'USER_SPACE_FULL') {
                toast.error('Недостаточно места в хранилище пользователя.');
            } else if (e.code === 'FILE_NAME_INVALID_CHARS' && e.details?.invalid?.length) {
                toast.error(`Недопустимые символы в имени файла: ${e.details.invalid.join(' ')}`);
            } else {
                toast.error(e.message || 'Не удалось создать файл из загруженного');
            }
        }
    }, [toast, parseAPIError]);

    const handleSave = useCallback(
        async (refreshFiles) => {
            try {
                let filename = fileHandle?.name;

                if (!filename) {
                    const asked = prompt('Введите имя файла', 'untitled.md');
                    if (!asked) return;
                    filename = asked.trim();
                }

                if (!filename.endsWith('.md') && !filename.endsWith('.markdown')) {
                    toast.info('Сохраняем как .md');
                    filename += '.md';
                }

                const v = validateFilename(filename);
                if (!v.ok) {
                    toast.error(v.message);
                    return;
                }

                const content = markdown;

                localStorage.setItem(filename, content);

                const blob = new Blob([content], { type: 'text/plain' });
                const formData = new FormData();
                formData.append('file', blob, filename);

                await API.STORAGE.put(`/file/${encodeURIComponent(filename)}`, formData, {
                    headers: { 'Content-Type': 'multipart/form-data' },
                });

                setFileHandle({ name: filename });
                setSavedSnapshot(content);
                setUnsaved(false);

                if (typeof refreshFiles === 'function') {
                    refreshFiles();
                }

                toast.success('Файл сохранён');
            } catch (err) {
                console.error('Ошибка сохранения файла', err);
                const e = parseAPIError(err);
                if (e.code === 'FILE_NOT_FOUND') {
                    toast.error('Файл не найден (возможно был удалён). Создайте заново.');
                } else if (e.code === 'USER_SPACE_FULL') {
                    toast.error('Недостаточно места в хранилище пользователя.');
                } else if (e.code === 'FILE_NAME_INVALID_CHARS' && e.details?.invalid?.length) {
                    toast.error(`Недопустимые символы: ${e.details.invalid.join(' ')}`);
                } else {
                    toast.error(e.message || 'Не удалось сохранить файл');
                }
            }
        },
        [markdown, fileHandle, toast, parseAPIError]
    );

    const handleDownloadCurrent = useCallback(async () => {
        if (!fileHandle?.name) return;
        try {
            const resp = await API.STORAGE.get(`/file/${encodeURIComponent(fileHandle.name)}`, { responseType: 'blob' });
            let filename = fileHandle.name;
            const cd = resp.headers?.['content-disposition'];
            if (cd) {
                const m = /filename\*=UTF-8''([^;]+)|filename="?([^"]+)"?/i.exec(cd);
                if (m) filename = decodeURIComponent(m[1] || m[2]);
            }
            const url = window.URL.createObjectURL(resp.data);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            console.error('File download error', err);
            const e = parseAPIError(err);
            toast.error(e.message || 'Не удалось скачать файл');
        }
    }, [fileHandle, toast, parseAPIError]);

    const handleLogout = useCallback(async () => {
        try {
            await API.AUTH.post('/v1/logout');
            window.location.href = '/login';
        } catch (err) {
            const e = parseAPIError(err);
            toast.error(e.message || 'Не удалось выполнить выход');
        }
    }, [toast, parseAPIError]);

    return (
        <>
            <div className="app-shell">
                <AppTopBar
                    sidebarOpen={sidebarOpen}
                    onToggleSidebar={toggleSidebar}
                    showPreview={showPreview}
                    onTogglePreview={() => setShowPreview(p => !p)}
                    current={fileHandle}
                    unsaved={unsaved}
                    onNewFile={() => setIsNewFileModalOpen(true)}
                    onSave={() => handleSave(() => sidebarRef.current?.refresh?.())}
                    onDownload={handleDownloadCurrent}
                    onLogout={handleLogout}
                    options={options}
                    onOptionsChange={handleOptionsChange}
                />

                <div
                    className="app-grid"
                    style={{
                        gridTemplateColumns: showPreview
                            ? `${sidebarOpen ? 260 : 0}px ${leftWidth}px 5px 1fr`
                            : `${sidebarOpen ? 260 : 0}px 1fr`
                    }}
                >
                <Toaster position="top-right" reverseOrder={false} />
                <FileSidebar
                    ref={sidebarRef}
                    current={fileHandle}
                    onOpenFile={handleOpenFile}
                    onSave={handleSave}
                    onNewFile={() => setIsNewFileModalOpen(true)}
                    unsaved={unsaved}
                    setUnsaved={setUnsaved}
                    collapsed={!sidebarOpen}
                    onToggle={toggleSidebar}
                    aiCurrent={fileHandle ? { name: fileHandle.name, text: markdown } : null}
                />

                <div className="left-panel">
                    <MarkdownEditor
                        value={markdown}
                        onChange={setMarkdown}
                        onFileUpload={handleFileUpload}
                    />
                </div >

                {showPreview && (
                    <>
                        <div
                            className="resizer"
                            onMouseDown={handleMouseDown}
                        />
                        <MarkdownPreview markdown={markdown} options={options} />
                    </>
                )}
                </div>
            </div>

            <NewFileModal
                open={isNewFileModalOpen}
                onClose={() => setIsNewFileModalOpen(false)}
                onConfirm={(filename) => {
                    setIsNewFileModalOpen(false);
                    handleNewFile(filename);
                }}
            />
        </>
    );
}
