import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import MarkdownEditor from './MarkdownEditor';
import FileSidebarUnauth from './FileSidebarUnauth';
import MarkdownPreview from './MarkdownPreview';
import AppTopBarUnauth from './AppTopBarUnauth';
import NewFileModal from './NewFileModal';
import { validateFilename } from "../utils";
import { toast, Toaster } from 'react-hot-toast';

export const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.

**Режим без авторизации** - файлы сохраняются локально в браузере.
`;

const DEFAULT_LEFT = Math.round(window.innerWidth * 0.4);

const DEFAULT_OPTIONS = {
    breaks: false,
    gfm: true,
    pedantic: false,
    silent: false
};

const UNAUTH_FILES_KEY = 'md_unauth_files';
const UNAUTH_CURRENT_FILE_KEY = 'md_unauth_current_file';

export default function MarkdownAppUnauth() {
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const [showPreview, setShowPreview] = useState(true);
    const toggleSidebar = () => setSidebarOpen(o => !o);
    const [leftWidth, setLeftWidth] = useState(DEFAULT_LEFT);
    const isResizing = useRef(false);
    const navigate = useNavigate();

    const [markdown, setMarkdown] = useState(DEFAULT_MD);
    const [fileHandle, setFileHandle] = useState(null);
    const [unsaved, setUnsaved] = useState(false);
    const [savedSnapshot, setSavedSnapshot] = useState("");
    const [isNewFileModalOpen, setIsNewFileModalOpen] = useState(false);

    const sidebarRef = useRef(null);

    const getStoredFiles = useCallback(() => {
        try {
            const stored = localStorage.getItem(UNAUTH_FILES_KEY);
            return stored ? JSON.parse(stored) : {};
        } catch {
            return {};
        }
    }, []);

    const setStoredFiles = useCallback((files) => {
        localStorage.setItem(UNAUTH_FILES_KEY, JSON.stringify(files));
    }, []);

    useEffect(() => {
        const files = getStoredFiles();
        const currentFileName = localStorage.getItem(UNAUTH_CURRENT_FILE_KEY);

        if (currentFileName && files[currentFileName]) {
            setMarkdown(files[currentFileName]);
            setFileHandle({ name: currentFileName });
            setSavedSnapshot(files[currentFileName]);
        } else if (Object.keys(files).length > 0) {
            const firstFileName = Object.keys(files)[0];
            setMarkdown(files[firstFileName]);
            setFileHandle({ name: firstFileName });
            setSavedSnapshot(files[firstFileName]);
        } else {
            const defaultFileName = 'welcome.md';
            files[defaultFileName] = DEFAULT_MD;
            setStoredFiles(files);
            setMarkdown(DEFAULT_MD);
            setFileHandle({ name: defaultFileName });
            setSavedSnapshot(DEFAULT_MD);
            localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, defaultFileName);
        }
    }, [getStoredFiles, setStoredFiles]);

    useEffect(() => {
        if (!fileHandle) {
            setUnsaved(false);
            return;
        }
        setUnsaved(markdown !== savedSnapshot);
    }, [markdown, fileHandle, savedSnapshot]);

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

    const handleOpenFile = useCallback((content, handle) => {
        setMarkdown(content);
        setFileHandle(handle);
        setSavedSnapshot(content);
        setUnsaved(false);
        if (handle) {
            localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, handle.name);
        }
    }, []);

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

            const files = getStoredFiles();

            if (files[filename]) {
                toast.error('Файл с таким именем уже существует. Выберите другое имя.');
                return;
            }

            files[filename] = DEFAULT_MD;
            setStoredFiles(files);

            setMarkdown(DEFAULT_MD);
            setFileHandle({ name: filename });
            setSavedSnapshot(DEFAULT_MD);
            setUnsaved(false);
            localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, filename);

            sidebarRef.current?.refresh?.();
            toast.success('Файл создан');
        } catch (err) {
            console.error('Ошибка создания файла', err);
            toast.error('Не удалось создать файл');
        }
    }, [getStoredFiles, setStoredFiles]);

    const handleFileUpload = useCallback(async (content, originalFilename) => {
        try {
            let filename = originalFilename || 'uploaded.md';

            if (!filename.endsWith('.md') && !filename.endsWith('.markdown')) {
                filename = filename.replace(/\.[^/.]+$/, "") + '.md';
            }

            const v = validateFilename(filename);
            if (!v.ok) {
                toast.error(v.message);
                return;
            }

            const files = getStoredFiles();

            if (files[filename]) {
                toast.error('Файл с таким именем уже существует. Переименуйте загружаемый файл.');
                return;
            }

            files[filename] = content;
            setStoredFiles(files);

            setMarkdown(content);
            setFileHandle({ name: filename });
            setSavedSnapshot(content);
            setUnsaved(false);
            localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, filename);

            sidebarRef.current?.refresh?.();
            toast.success(`Файл "${filename}" создан из загруженного файла`);
        } catch (err) {
            console.error('Ошибка создания файла из загруженного', err);
            toast.error('Не удалось создать файл из загруженного');
        }
    }, [getStoredFiles, setStoredFiles]);

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

                const files = getStoredFiles();
                files[filename] = markdown;
                setStoredFiles(files);

                setFileHandle({ name: filename });
                setSavedSnapshot(markdown);
                setUnsaved(false);
                localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, filename);

                if (typeof refreshFiles === 'function') {
                    refreshFiles();
                }

                toast.success('Файл сохранён');
            } catch (err) {
                console.error('Ошибка сохранения файла', err);
                toast.error('Не удалось сохранить файл');
            }
        },
        [markdown, fileHandle, getStoredFiles, setStoredFiles]
    );

    const handleDownloadCurrent = useCallback(async () => {
        if (!fileHandle?.name) return;
        try {
            const files = getStoredFiles();
            const content = files[fileHandle.name];

            if (!content) {
                toast.error('Файл не найден');
                return;
            }

            const blob = new Blob([content], { type: 'text/markdown' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = fileHandle.name;
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            console.error('File download error', err);
            toast.error('Не удалось скачать файл');
        }
    }, [fileHandle, getStoredFiles]);

    const handleBackToLogin = useCallback(() => {
        navigate('/login');
    }, [navigate]);

    return (
        <>
            <div className="app-shell">
                <AppTopBarUnauth
                    sidebarOpen={sidebarOpen}
                    onToggleSidebar={toggleSidebar}
                    showPreview={showPreview}
                    onTogglePreview={() => setShowPreview(p => !p)}
                    current={fileHandle}
                    unsaved={unsaved}
                    onNewFile={() => setIsNewFileModalOpen(true)}
                    onSave={() => handleSave(() => sidebarRef.current?.refresh?.())}
                    onDownload={handleDownloadCurrent}
                    onBackToLogin={handleBackToLogin}
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
                    <FileSidebarUnauth
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
