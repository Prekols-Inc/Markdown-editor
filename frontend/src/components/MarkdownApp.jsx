import { useState, useEffect, useCallback, useRef } from 'react';
import { PanelRightOpen, PanelRightClose } from 'lucide-react';
import MarkdownEditor from './MarkdownEditor';
import FileSidebar from './FileSidebar';
import OptionsEditor from './OptionsEditor';
import MarkdownPreview from './MarkdownPreview';
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
            const sidebar = sidebarOpen ? 260 : 48;
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

    const [tab, setTab] = useState('markdown');

    const [fileHandle, setFileHandle] = useState(null);
    const [unsaved, setUnsaved] = useState(false);

    const sidebarRef = useRef(null);

    useEffect(() => {
        if (fileHandle) setUnsaved(true);
    }, [markdown, fileHandle]);

    const handleOpenFile = useCallback((text, handle) => {
        setMarkdown(text);
        setFileHandle(handle);
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
    return (
        <>
            <div
                className="app-grid"
                style={{
                    gridTemplateColumns: showPreview
                        ? `${sidebarOpen ? 260 : 48}px ${leftWidth}px 5px 1fr`
                        : `${sidebarOpen ? 260 : 48}px 1fr`
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
                />

                <div className="left-panel">
                    <div className="tabs">
                        <button
                            className={tab === 'markdown' ? 'tab active' : 'tab'}
                            onClick={() => setTab('markdown')}
                        >
                            Markdown
                        </button>
                        <button
                            className={tab === 'options' ? 'tab active' : 'tab'}
                            onClick={() => setTab('options')}
                        >
                            Options
                        </button>

                        <button
                            className="tab right flex items-center gap-2"
                            onClick={() => setShowPreview(p => !p)}
                            title={showPreview ? 'Скрыть превью' : 'Показать превью'}
                        >
                            {showPreview ? (
                                <PanelRightClose size={22} strokeWidth={1.75} />
                            ) : (
                                <PanelRightOpen size={22} strokeWidth={1.75} />
                            )}
                        </button>
                    </div>

                    {tab === 'markdown' ? (
                        <MarkdownEditor value={markdown} onChange={setMarkdown} />
                    ) : (
                        <OptionsEditor
                            value={options}
                            onChange={handleOptionsChange}
                        />
                    )}
                </div>

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
