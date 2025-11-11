import { useState, useEffect, useCallback, useRef } from 'react';
import MarkdownEditor from './MarkdownEditor';
import FileSidebar from './FileSidebar';
import OptionsEditor from './OptionsEditor';
import MarkdownPreview from './MarkdownPreview';
import { marked } from 'marked';
import API from '../API';
import NewFileModal from './NewFileModal';
import { isValidFilename } from '../utils';

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

    const handleNewFile = useCallback(async (filename) => {
        try {
            if (!filename.endsWith('.md')) {
                filename += '.md';
            }
            if (!isValidFilename(filename)) {
                alert('Недопустимое имя файла!'); // todo: change with notification
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

            sidebarRef.current?.refresh();
        } catch (err) {
            console.error('Ошибка создания файла', err);
            alert('Не удалось создать файл'); // todo: change with notification
        }
    }, []);

    const handleSave = useCallback(
        async (refreshFiles) => {
            try {
                let filename = fileHandle?.name;
                if (!filename) {
                    filename = prompt('Введите имя файла', 'untitled.md');
                    if (!filename) return;
                }

                if (!/\.(md|markdown|txt|html)$/i.test(filename)) {
                    filename += '.md';
                }

                let content;
                if (filename.endsWith('.html')) {
                    content = marked.parse(markdown, options);
                } else {
                    content = markdown;
                }

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
            } catch (err) {
                console.error('Ошибка сохранения файла', err);
                alert('Не удалось сохранить файл'); // todo: change with notification
            }
        },
        [markdown, options, fileHandle]
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
                            className="tab right"
                            onClick={() => setShowPreview(p => !p)}
                            title={'Показать/Скрыть превью'}
                        >
                            {showPreview ? '🡸 Hide Preview' : '🡺 Show Preview'}
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
