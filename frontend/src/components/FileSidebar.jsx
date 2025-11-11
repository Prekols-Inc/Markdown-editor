import { useEffect, useState, forwardRef, useImperativeHandle, useRef } from 'react';
import LogoutConfirmModal from "./LogoutConfirmModal";
import API from '../API';
import { validateFilename } from '../utils';
import { useToast } from './ToastProvider';

const FileSidebar = forwardRef(function FileSidebar(
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
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [saveMenuOpen, setSaveMenuOpen] = useState(false);
  const saveGroupRef = useRef(null);
  const [editingFile, setEditingFile] = useState(null);
  const [newFileName, setNewFileName] = useState("");

  const toast = useToast();
  const parseAPIError =
    (API && API.parseAPIError)
      ? API.parseAPIError
      : (e) => ({ code: 'GENERIC', message: e?.response?.data?.error || e?.message || 'Произошла ошибка' });

  const startRename = (file) => {
    setEditingFile(file.name);
    setNewFileName(file.name);
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
      await API.STORAGE.put(`/rename/${encodeURIComponent(oldName)}/${encodeURIComponent(newName)}`);
      setEntries((prev) =>
        prev.map((f) => (f.name === oldName ? { ...f, name: newName } : f))
      );
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
    } finally {
      cancelRename();
    }
  };

  const downloadFile = async (file) => {
    try {
      const resp = await API.STORAGE.get(
        `/file/${encodeURIComponent(file.name)}`,
        { responseType: 'blob' }
      );

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
  };

  const fetchFiles = async () => {
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
  };

  const openFile = async (file) => {
    try {
      const cachedFile = localStorage.getItem(file.name);
      if (cachedFile != null) {
        onOpenFile(cachedFile, { name: file.name });
        setUnsaved(true);
      }
      else {
        const response = await API.STORAGE.get(`/file/${encodeURIComponent(file.name)}`);
        onOpenFile(response.data, { name: file.name });
        setUnsaved(false);
        localStorage.setItem(file.name, response.data);
      }
    } catch (err) {
      console.error('Ошибка загрузки файла', err);
      const e = parseAPIError(err);
      toast.error(e.message || 'Не удалось открыть файл');
    }
  };

  const deleteFile = async (file) => {
    try {
      await API.STORAGE.delete(`/file/${encodeURIComponent(file.name)}`);
      const newList = entries.filter(x => x.name !== file.name);
      setEntries(newList);
      localStorage.removeItem(file.name);

      if (current?.name === file.name) {
        if (newList.length > 0) {
          openFile(newList[0]);
        } else {
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
  };

  useEffect(() => {
    fetchFiles();
  }, []);

  useEffect(() => {
    const onDocClick = (e) => {
      if (saveGroupRef.current && !saveGroupRef.current.contains(e.target)) {
        setSaveMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', onDocClick);
    return () => document.removeEventListener('mousedown', onDocClick);
  }, []);

  useImperativeHandle(ref, () => ({
    refresh: fetchFiles,
  }));

  const handleLogout = async () => {
    try {
      await API.AUTH.post('/v1/logout');
      window.location.href = '/login';
    } catch (err) {
      const e = parseAPIError(err);
      toast.error(e.message || 'Не удалось выполнить выход');
    }
  };

  return (
    <aside
      className={collapsed ? 'sidebar collapsed' : 'sidebar'}
      style={{ width: collapsed ? 48 : 260 }}
    >
      <div className="toolbar">
        <button
          className="btn secondary"
          style={{ width: 32 }}
          onClick={onToggle}
          title={collapsed ? 'Expand' : 'Collapse'}
        >
          {collapsed ? '»' : '«'}
        </button>

        {!collapsed && (
          <>
            <button className="btn" onClick={onNewFile}>New</button>
            <div className="btn-group" ref={saveGroupRef}>
              <button
                className="btn split-main"
                disabled={!current && !unsaved}
                onClick={() => {
                  onSave(fetchFiles);
                  setSaveMenuOpen(false);
                }}
              >
                Save
              </button>
              <button
                className="btn split-toggle"
                disabled={!current && !unsaved}
                onClick={() => setSaveMenuOpen(v => !v)}
                aria-haspopup="menu"
                aria-expanded={saveMenuOpen}
              />
              <div
                className={`menu ${saveMenuOpen ? 'open' : ''}`}
                role="menu"
              >
                <button
                  className="menu-item"
                  role="menuitem"
                  onClick={() => {
                    if (current) downloadFile(current);
                    setSaveMenuOpen(false);
                  }}
                  disabled={!current}
                >
                  Download
                </button>
              </div>
            </div>

            <button
              className="btn danger"
              style={{ backgroundColor: "#e74c3c", color: "white" }}
              onClick={() => setShowLogoutConfirm(true)}
            >
              Logout
            </button>

            <LogoutConfirmModal
              open={showLogoutConfirm}
              onClose={() => setShowLogoutConfirm(false)}
              onConfirm={handleLogout}
            />
          </>
        )}
      </div>

      {!collapsed && entries.map(file => (
        <div
          key={file.name}
          className={'fs-item' + (current?.name === file.name ? ' active' : '')}
          title={file.name}
          onClick={() => openFile(file)}
          onDoubleClick={() => startRename(file)}
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
                style={{
                  width: "100%",
                  background: "transparent",
                  color: "inherit",
                  border: "1px solid #555",
                  borderRadius: 4,
                  padding: "2px 4px",
                }}
              />
            ) : (
              <>
                {file.name}
                {unsaved && current?.name === file.name && ' ●'}
              </>
            )}
          </span>
          <button
            className="fs-close"
            title="Удалить файл"
            onClick={(ev) => {
              ev.stopPropagation();
              deleteFile(file);
            }}
          >
            ×
          </button>
        </div>
      ))}

    </aside >
  );
});

export default FileSidebar;
