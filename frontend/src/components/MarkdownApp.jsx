import { useState, useEffect, useCallback, useRef } from 'react';
import MarkdownEditor from './MarkdownEditor';
import FileSidebar from './FileSidebar';
import OptionsEditor from './OptionsEditor';
import MarkdownPreview from './MarkdownPreview';
import { marked } from 'marked';
import API from '../API';
import NewFileModal from './NewFileModal';

export const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.
`;

const DEFAULT_LEFT = Math.round(window.innerWidth * 0.4);
const DEFAULT_OPTIONS = { breaks: false, gfm: true, pedantic: false, silent: false };

function TopBar({ onNew, onSave, onToggleSidebar, onLogout, sidebarOpen, unsaved }) {
    return (
        <header
            className="topbar"
            style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                background: '#f7f7f7',
                borderBottom: '1px solid #ddd',
                padding: '8px 16px',
                height: '48px',
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
            }}
        >
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <button
                    className="btn secondary"
                    onClick={onToggleSidebar}
                    title={sidebarOpen ? 'Collapse sidebar' : 'Expand sidebar'}
                    style={{ width: 32 }}
                >
                    {sidebarOpen ? '«' : '»'}
                </button>

                <button className="btn" onClick={onNew}>New</button>

                <button
                    className="btn"
                    disabled={!unsaved}
                    onClick={onSave}
                >
                    Save
                </button>
            </div>

            <button
                className="btn logout"
                onClick={onLogout}
                style={{
                    backgroundColor: '#ff5252',
                    color: 'white',
                    border: 'none',
                    padding: '6px 14px',
                    borderRadius: '6px',
                    cursor: 'pointer',
                    fontWeight: 500,
                }}
            >
                Logout
            </button>
        </header>
    );
}

export default function App() {
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const toggleSidebar = () => setSidebarOpen(o => !o);
    const [leftWidth, setLeftWidth] = useState(DEFAULT_LEFT);
    const isResizing = useRef(false);
    const [markdown, setMarkdown] = useState(() => localStorage.getItem('md-draft') ?? DEFAULT_MD);
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
            if (!/\.(md|markdown|txt|html)$/i.test(filename)) {
                filename += '.md';
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
            alert('Не удалось создать файл');
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
                alert('Не удалось сохранить файл');
            }
        },
        [markdown, options, fileHandle]
    );

    const handleLogout = useCallback(async () => {
        try {
            await API.AUTH.post('/v1/logout');
            window.location.href = '/login';
        } catch (err) {
            alert('Не удалось выполнить выход');
        }
    }, []);

    return (
        <div className="app-wrapper" style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
            <TopBar
                onNew={() => setIsNewFileModalOpen(true)}
                onSave={() => handleSave(sidebarRef.current?.refresh)}
                onLogout={handleLogout}
                onToggleSidebar={toggleSidebar}
                sidebarOpen={sidebarOpen}
                unsaved={unsaved}
            />

            <div
                className="app-grid"
                style={{
                    flex: 1,
                    display: 'grid',
                    gridTemplateColumns: `${sidebarOpen ? 260 : 48}px ${leftWidth}px 5px 1fr`,
                    height: 'calc(100vh - 48px)',
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
                    </div>

                    {tab === 'markdown' ? (
                        <MarkdownEditor value={markdown} onChange={setMarkdown} />
                    ) : (
                        <OptionsEditor value={options} onChange={handleOptionsChange} />
                    )}
                </div>

                <div className="resizer" onMouseDown={handleMouseDown} />
                <MarkdownPreview markdown={markdown} options={options} />
            </div>

            <NewFileModal
                open={isNewFileModalOpen}
                onClose={() => setIsNewFileModalOpen(false)}
                onConfirm={(filename) => {
                    setIsNewFileModalOpen(false);
                    handleNewFile(filename);
                }}
            />
        </div>
    );
}
