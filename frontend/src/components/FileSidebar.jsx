import { useEffect, useState, forwardRef, useImperativeHandle, useRef } from 'react';
import { FilePlus2, Save, Download, LogOut, ChevronsLeft, ChevronsRight } from 'lucide-react';
import LogoutConfirmModal from "./LogoutConfirmModal";
import API from '../API';
import { UnauthStorage } from '../storage/unauthStorage';
import { toast } from 'react-hot-toast';
import AISummarizeButton from "./AISummarizeButton";
import { validateFilename } from '../utils';

const DEFAULT_MD = "# Новый Markdown файл\n\nНапишите здесь...";

const FileSidebar = forwardRef(function FileSidebar(
  {
    isUnauth,
    current,
    aiCurrent,
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
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
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

  const parseAPIError =
    (API && API.parseAPIError)
      ? API.parseAPIError
      : (e) => ({ code: 'GENERIC', message: e?.response?.data?.error || e?.message || 'Произошла ошибка' });

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

    if (isUnauth) {
      // Rename file in localStorage for unauth mode
      const files = UnauthStorage.load();
      if (files[oldName]) {
        files[newName] = files[oldName];
        delete files[oldName];
        UnauthStorage.save(files);

        setEntries(prev => prev.map(f => f.name === oldName ? { ...f, name: newName } : f));

        if (current?.name === oldName) {
          onOpenFile(files[newName], { name: newName });
        }
        toast.success('Файл переименован');
      }
    } else {
      try {
        const formData = new FormData();
        formData.append('oldName', oldName);
        formData.append('newName', newName);

        await API.STORAGE.put(`/rename/${encodeURIComponent(oldName)}/${encodeURIComponent(newName)}`, formData);

        setEntries(prev => prev.map(f => f.name === oldName ? { ...f, name: newName } : f));

        if (current?.name === oldName) {
          onOpenFile(localStorage.getItem(oldName) || "", { name: newName });
          localStorage.setItem(newName, localStorage.getItem(oldName));
          localStorage.removeItem(oldName);
        }
        toast.success('Файл переименован');
      } catch (err) {
        console.error("Ошибка при переименовании файла", err);
        const e = parseAPIError(err);
        if (e.code === 'FILE_NAME_INVALID_CHARS' && e.details?.invalid?.length) {
          toast.error(`Недопустимые символы: ${e.details.invalid.join(' ')}`);
        } else {
          toast.error(e.message || 'Не удалось переименовать файл');
        }
      }
    }
    cancelRename();
  };

  const duplicateFile = async (file) => {
    let base = file.name.replace(/\.md$/, '');
    let newName = base + '_copy.md';
    let idx = 1;
    while (entries.find(e => e.name === newName)) {
      newName = `${base}_copy${idx}.md`;
      idx++;
    }

    if (isUnauth) {
      // Duplicate file in localStorage for unauth mode
      const files = UnauthStorage.load();
      if (files[file.name]) {
        files[newName] = files[file.name];
        UnauthStorage.save(files);
        fetchFiles();
        toast.success('Файл дублирован');
      }
    } else {
      try {
        const response = await API.STORAGE.get(`/file/${encodeURIComponent(file.name)}`, { responseType: 'text' });
        const text = response.data;
        const blob = new Blob([text], { type: 'text/plain' });
        const formData = new FormData();
        formData.append('file', blob, newName);

        await API.STORAGE.post(`/file/${encodeURIComponent(newName)}`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        });

        fetchFiles();
      } catch (err) {
        alert('Не удалось дублировать файл');
        console.error('duplicate error', err);
      }
    }
    setContextMenu(ctx => ({ ...ctx, visible: false }));
  };

  const downloadFile = async (file) => {
    if (isUnauth) {
      // Download file from localStorage for unauth mode
      const files = UnauthStorage.load();
      const content = files[file.name];
      if (content) {
        const blob = new Blob([content], { type: 'text/plain' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = file.name;
        document.body.appendChild(a);
        a.click();
        a.remove();
        window.URL.revokeObjectURL(url);
      }
    } else {
      try {
        const resp = await API.STORAGE.get(`/file/${encodeURIComponent(file.name)}`, { responseType: 'blob' });
        let filename = file.name;
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
    }
  };

  const fetchFiles = async () => {
    if (isUnauth) {
      // Load files from localStorage for unauth mode
      const files = UnauthStorage.load();
      const fileList = Object.keys(files).map(name => ({ name }));
      setEntries(fileList);

      if (fileList.length === 0) {
        onOpenFile("Нет открытых файлов. Нажмите «New» для создания.", null);
        setUnsaved(false);
      } else if (!current) {
        openFile(fileList[0]);
      }
    } else {
      try {
        const response = await API.STORAGE.get('/files');
        const files = response.data.files.map(name => ({ name }));
        setEntries(files);

        if (files.length === 0) {
          onOpenFile("Нет открытых файлов. Нажмите «New» для создания.", null);
          setUnsaved(false);
        } else if (!current) {
          openFile(files[0]);
        }
      } catch (err) {
        console.error('Ошибка загрузки файлов', err);
        const e = parseAPIError(err);
        toast.error(e.message || 'Не удалось загрузить список файлов');
      }
    }
  };

  const openFile = async (file) => {
    if (isUnauth) {
      // Load file from localStorage for unauth mode
      const files = UnauthStorage.load();
      const content = files[file.name] || DEFAULT_MD;
      onOpenFile(content, { name: file.name });
      setUnsaved(false);
    } else {
      try {
        const cachedFile = localStorage.getItem(file.name);
        if (cachedFile != null) {
          onOpenFile(cachedFile, { name: file.name });
          setUnsaved(false);
        } else {
          const response = await API.STORAGE.get(`/file/${encodeURIComponent(file.name)}`, { responseType: 'text' });
          onOpenFile(response.data, { name: file.name });
          setUnsaved(false);
          localStorage.setItem(file.name, response.data);
        }
      } catch (err) {
        console.error('Ошибка загрузки файла', err);
        const e = parseAPIError(err);
        toast.error(e.message || 'Не удалось открыть файл');
      }
    }
  };

  const deleteFile = async (file) => {
    if (isUnauth) {
      // Delete file from localStorage for unauth mode
      UnauthStorage.remove(file.name);
      const newList = entries.filter(x => x.name !== file.name);
      setEntries(newList);

      if (current?.name === file.name) {
        if (newList.length > 0) openFile(newList[0]);
        else {
          onOpenFile("Нет открытых файлов. Нажмите «New» для создания.", null);
          setUnsaved(false);
        }
      }
      toast.success('Файл удалён');
    } else {
      try {
        await API.STORAGE.delete(`/file/${encodeURIComponent(file.name)}`);
        const newList = entries.filter(x => x.name !== file.name);
        setEntries(newList);
        localStorage.removeItem(file.name);

        if (current?.name === file.name) {
          if (newList.length > 0) openFile(newList[0]);
          else {
            onOpenFile("Нет открытых файлов. Нажмите «New» для создания.", null);
            setUnsaved(false);
          }
        }
        toast.success('Файл удалён');
      } catch (err) {
        console.error('Ошибка удаления файла', err);
        const e = parseAPIError(err);
        toast.error(e.message || 'Не удалось удалить файл');
      }
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

  const handleLogout = async () => {
    if (isUnauth) {
      // For unauth mode, just remove the editorMode to allow choosing again
      localStorage.removeItem('editorMode');
      window.location.href = '/login';
    } else {
      try {
        await API.AUTH.post('/v1/logout');
        localStorage.removeItem('editorMode');
        window.location.href = '/login';
      } catch (err) {
        const e = parseAPIError(err);
        toast.error(e.message || 'Не удалось выполнить выход');
      }
    }
  };

  return (
    <aside
      className={collapsed ? 'sidebar collapsed' : 'sidebar'}
      style={{
        width: collapsed ? 48 : 260,
        minWidth: collapsed ? 48 : 260,
        borderRight: collapsed ? '1px solid #ddd' : undefined,
        overflow: 'visible'
      }}
    >
      {/* Sidebar Header */}
      <div className="sidebar-header">
        <button
          className="sidebar-toggle"
          onClick={onToggle}
          title={collapsed ? 'Развернуть боковую панель' : 'Свернуть боковую панель'}
        >
          {collapsed ? <ChevronsRight size={20} /> : <ChevronsLeft size={20} />}
        </button>

        {!collapsed && (
          <div className="sidebar-actions">
            <button
              className="sidebar-btn"
              onClick={onNewFile}
              title="Новый файл"
            >
              <FilePlus2 size={18} />
            </button>

            <div className="save-group" ref={saveGroupRef}>
              <button
                className="sidebar-btn"
                onClick={() => setSaveMenuOpen(!saveMenuOpen)}
                title="Сохранить"
              >
                <Save size={18} />
              </button>

              {saveMenuOpen && (
                <div className="save-dropdown">
                  <button onClick={() => { onSave(); setSaveMenuOpen(false); }}>
                    Сохранить
                  </button>
                  <button onClick={() => { onSave(() => { }); setSaveMenuOpen(false); }}>
                    Сохранить как...
                  </button>
                </div>
              )}
            </div>

            <button
              className="sidebar-btn"
              onClick={handleLogout}
              title={isUnauth ? 'Переключиться на авторизованный режим' : 'Выйти'}
            >
              <LogOut size={18} />
            </button>
          </div>
        )}
      </div>

      {/* File List */}
      {!collapsed && (
        <div className="file-list">
          {entries.map(file => (
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
        </div>
      )}

      {/* Context Menu */}
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
          {isUnauth && (
            <button
              className="dropdown-item"
              style={{ background: "none", border: "none", width: "100%", padding: "8px 16px", textAlign: "left", cursor: "pointer", color: "#333" }}
              onClick={() => { downloadFile(contextMenu.file); setContextMenu(ctx => ({ ...ctx, visible: false })); }}
            >
              Скачать
            </button>
          )}
        </div>
      )}

      {/* AI Summarize Button (only in auth mode) */}
      {!collapsed && !isUnauth && (
        <AISummarizeButton current={aiCurrent} />
      )}

      {/* Logout Confirmation Modal */}
      <LogoutConfirmModal
        open={showLogoutConfirm}
        onConfirm={handleLogout}
        onCancel={() => setShowLogoutConfirm(false)}
      />
    </aside>
  );
});

export default FileSidebar;
