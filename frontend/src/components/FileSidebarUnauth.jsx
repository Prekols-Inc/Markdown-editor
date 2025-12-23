import { useEffect, useState, forwardRef, useImperativeHandle, useRef } from 'react';
import { FilePlus2, Save, Download, LogOut, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { toast } from 'react-hot-toast';
import { validateFilename } from '../utils';

const DEFAULT_MD = "# Новый Markdown файл\n\nНапишите здесь...";
const UNAUTH_FILES_KEY = 'md_unauth_files';
const UNAUTH_CURRENT_FILE_KEY = 'md_unauth_current_file';

const FileSidebarUnauth = forwardRef(function FileSidebarUnauth(
    {
        current,
        onOpenFile,
        onSave,
        onNewFile,
        unsaved,
        setUnsaved,
        collapsed = false,
        onToggle
    },
    ref
) {
    const [entries, setEntries] = useState([]);
    const [saveMenuOpen, setSaveMenuOpen] = useState(false);
    const saveGroupRef = useRef(null);
    const [editingFile, setEditingFile] = useState(null);
    const [newFileName, setNewFileName] = useState("");
    const [contextMenu, setContextMenu] = useState({
        visible: false,
        x: 0,
        y: 0,
        file: null
    });
    const menuRef = useRef(null);

    const getStoredFiles = () => {
        try {
            const stored = localStorage.getItem(UNAUTH_FILES_KEY);
            return stored ? JSON.parse(stored) : {};
        } catch {
            return {};
        }
    };

    const setStoredFiles = (files) => {
        localStorage.setItem(UNAUTH_FILES_KEY, JSON.stringify(files));
    };

    const startRename = (file) => {
        setEditingFile(file.name);
        setNewFileName(file.name);
        setContextMenu(ctx => ({ ...ctx, visible: false }));
    };

    const cancelRename = () => {
        setEditingFile(null);
        setNewFileName("");
    };

    const confirmRename = async (oldName, newName) => {
        if (!newName.endsWith('.md') && !newName.endsWith('.markdown')) {
            newName += '.md';
        }
        if (!newName.trim() || newName === oldName) {
            cancelRename();
            return;
        }
        const v = validateFilename(newName);
        if (!v.ok) {
            toast.error(v.message);
            cancelRename();
            return;
        }

        try {
            const files = getStoredFiles();

            if (files[newName]) {
                toast.error('Файл с таким именем уже существует');
                cancelRename();
                return;
            }

            files[newName] = files[oldName];
            delete files[oldName];
            setStoredFiles(files);

            setEntries(prev => prev.map(f => f.name === oldName ? { ...f, name: newName } : f));

            if (current?.name === oldName) {
                onOpenFile(files[newName], { name: newName });
                localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, newName);
            }

            toast.success('Файл переименован');
        } catch (err) {
            console.error("Ошибка при переименовании файла", err);
            toast.error('Не удалось переименовать файл');
        } finally {
            cancelRename();
        }
    };

    const duplicateFile = async (file) => {
        let base = file.name.replace(/\.md$/, '');
        let newName = base + '_copy.md';
        let idx = 1;
        while (entries.find(e => e.name === newName)) {
            newName = `${base}_copy${idx}.md`;
            idx++;
        }
        try {
            const files = getStoredFiles();
            const content = files[file.name];

            if (content) {
                files[newName] = content;
                setStoredFiles(files);
                fetchFiles();
                toast.success('Файл дублирован');
            } else {
                toast.error('Не удалось найти содержимое файла');
            }
        } catch (err) {
            toast.error('Не удалось дублировать файл');
            console.error('duplicate error', err);
        }
        setContextMenu(ctx => ({ ...ctx, visible: false }));
    };

    const downloadFile = async (file) => {
        try {
            const files = getStoredFiles();
            const content = files[file.name];

            if (!content) {
                toast.error('Файл не найден');
                return;
            }

            const blob = new Blob([content], { type: 'text/markdown' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = file.name;
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            console.error('File download error', err);
            toast.error('Не удалось скачать файл');
        }
    };

    const fetchFiles = () => {
        try {
            const files = getStoredFiles();
            const fileNames = Object.keys(files).map(name => ({ name }));
            setEntries(fileNames);

            if (fileNames.length === 0) {
                onOpenFile("Нет открытых файлов. Нажмите «New» для создания.", null);
                setUnsaved(false);
            } else if (!current) {
                openFile(fileNames[0]);
            }
        } catch (err) {
            console.error('Ошибка загрузки файлов', err);
            toast.error('Не удалось загрузить список файлов');
        }
    };

    const openFile = (file) => {
        try {
            const files = getStoredFiles();
            const content = files[file.name];

            if (content !== undefined) {
                onOpenFile(content, { name: file.name });
                setUnsaved(false);
                localStorage.setItem(UNAUTH_CURRENT_FILE_KEY, file.name);
            } else {
                toast.error('Файл не найден');
            }
        } catch (err) {
            console.error('Ошибка загрузки файла', err);
            toast.error('Не удалось открыть файл');
        }
    };

    const deleteFile = (file) => {
        try {
            const files = getStoredFiles();
            delete files[file.name];
            setStoredFiles(files);

            const newList = entries.filter(x => x.name !== file.name);
            setEntries(newList);

            if (current?.name === file.name) {
                if (newList.length > 0) {
                    openFile(newList[0]);
                } else {
                    onOpenFile("Нет открытых файлов. Нажмите «New» для создания.", null);
                    setUnsaved(false);
                    localStorage.removeItem(UNAUTH_CURRENT_FILE_KEY);
                }
            }
            toast.success('Файл удалён');
        } catch (err) {
            console.error('Ошибка удаления файла', err);
            toast.error('Не удалось удалить файл');
        }
    };

    useEffect(() => { fetchFiles(); }, []);

    useEffect(() => {
        const onDocClick = (e) => {
            if (saveGroupRef.current && !saveGroupRef.current.contains(e.target)) {
                setSaveMenuOpen(false);
            }
        };
        document.addEventListener('click', onDocClick);
        return () => document.removeEventListener('click', onDocClick);
    }, [saveGroupRef]);

    useEffect(() => {
        const handleClickOutside = (e) => {
            if (contextMenu.visible && menuRef.current && !menuRef.current.contains(e.target)) {
                setContextMenu(ctx => ({ ...ctx, visible: false }));
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [contextMenu.visible]);

    useImperativeHandle(ref, () => ({ refresh: fetchFiles }));

    return (
        <aside
            className={collapsed ? 'sidebar collapsed' : 'sidebar'}
            style={{
                width: collapsed ? 0 : 260,
                minWidth: collapsed ? 0 : 260,
                borderRight: collapsed ? 'none' : undefined,
                overflow: collapsed ? 'hidden' : undefined
            }}
        >
            {!collapsed && entries.map(file => (
                <div
                    key={file.name}
                    className={'fs-item' + (current?.name === file.name ? ' active' : '')}
                    title={file.name}
                    onClick={() => openFile(file)}
                    onDoubleClick={() => startRename(file)}
                    onContextMenu={(e) => {
                        e.preventDefault();
                        setContextMenu({ visible: true, x: e.clientX, y: e.clientY, file });
                    }}
                >
                    <span className="fs-name" onDoubleClick={(e) => { e.stopPropagation(); startRename(file); }}>
                        {editingFile === file.name ? (
                            <input
                                type="text"
                                value={newFileName}
                                autoFocus
                                onChange={(e) => setNewFileName(e.target.value)}
                                onBlur={() => confirmRename(file.name, newFileName)}
                                onKeyDown={(e) => {
                                    if (e.key === "Enter") confirmRename(file.name, newFileName);
                                    if (e.key === "Escape") cancelRename();
                                }}
                                className="rename-input"
                                style={{ width: "100%", background: "transparent", color: "inherit", border: "1px solid #555", borderRadius: 4, padding: "2px 4px" }}
                            />
                        ) : (
                            <>
                                {file.name}{unsaved && current?.name === file.name && ' ●'}
                            </>
                        )}
                    </span>

                    <button className="fs-close" title="Удалить файл" onClick={(ev) => { ev.stopPropagation(); deleteFile(file); }}>×</button>
                </div>
            ))}

            {!collapsed && contextMenu.visible && (
                <div
                    ref={menuRef}
                    className="dropdown-menu"
                    style={{
                        position: 'fixed',
                        left: contextMenu.x + 2,
                        top: contextMenu.y + 2,
                        zIndex: 9999,
                        background: '#fff',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.13)',
                        borderRadius: 6,
                        padding: "4px 0",
                        border: "1px solid #e3e3e3",
                        minWidth: 140
                    }}
                >
                    <button
                        className="dropdown-item"
                        style={{ background: "none", border: "none", width: "100%", padding: "8px 16px", textAlign: "left", cursor: "pointer", color: "#333" }}
                        onClick={() => { startRename(contextMenu.file); setContextMenu(ctx => ({ ...ctx, visible: false })); }}
                    >
                        Переименовать
                    </button>
                    <button
                        className="dropdown-item"
                        style={{ background: "none", border: "none", width: "100%", padding: "8px 16px", textAlign: "left", cursor: "pointer", color: "#333" }}
                        onClick={() => { duplicateFile(contextMenu.file); setContextMenu(ctx => ({ ...ctx, visible: false })); }}
                    >
                        Дублировать
                    </button>
                    <button
                        className="dropdown-item"
                        style={{ background: "none", border: "none", width: "100%", padding: "8px 16px", textAlign: "left", cursor: "pointer", color: "#333" }}
                        onClick={() => { downloadFile(contextMenu.file); setContextMenu(ctx => ({ ...ctx, visible: false })); }}
                    >
                        Скачать
                    </button>
                </div>
            )}
        </aside>
    );
});

export default FileSidebarUnauth;
