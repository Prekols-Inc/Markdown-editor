import { useState, useEffect, useCallback, useRef } from 'react';
import MarkdownEditor from './MarkdownEditor';
import FileSidebar from './FileSidebar';
import OptionsEditor from './OptionsEditor';
import MarkdownPreview from './MarkdownPreview';
import { marked } from 'marked';
import API from '../API';

const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.
`;

const DEFAULT_OPTIONS = {
    async: false,
    breaks: false,
    extensions: null,
    gfm: true,
    hooks: null,
    pedantic: false,
    silent: false,
    tokenizer: null,
    walkTokens: null
};

export default function App() {
    const [sidebarOpen, setSidebarOpen] = useState(true);
    const toggleSidebar = () => setSidebarOpen(o => !o);
    const [markdown, setMarkdown] = useState(
        () => localStorage.getItem('md-draft') ?? DEFAULT_MD
    );

    useEffect(() => {
        const id = setTimeout(() => localStorage.setItem('md-draft', markdown), 400);
        return () => clearTimeout(id);
    }, [markdown]);

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

    const handleNewFile = useCallback(async () => {
        try {
            let filename = prompt('Введите имя нового файла', 'untitled.md');
            if (!filename) return;

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


    return (
        <div
            className="app-grid"
            style={{ gridTemplateColumns: `${sidebarOpen ? 260 : 48}px 1fr 1fr` }}
        >
            <FileSidebar
                ref={sidebarRef}
                current={fileHandle}
                onOpenFile={handleOpenFile}
                onSave={handleSave}
                onNewFile={handleNewFile}
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
                    <OptionsEditor
                        value={JSON.stringify(options, null, 2)}
                        onChange={handleOptionsChange}
                    />
                )}
            </div>

            <MarkdownPreview markdown={markdown} options={options} />
        </div>
    );
}
