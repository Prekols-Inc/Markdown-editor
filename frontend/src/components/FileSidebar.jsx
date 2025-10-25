import { useEffect, useState, forwardRef, useImperativeHandle, cache, useRef } from 'react';
import API from '../API';

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
  const [saveMenuOpen, setSaveMenuOpen] = useState(false);
  const saveGroupRef = useRef(null);

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
      console.error('Ошибка скачивания файла', err);
      alert('Не удалось скачать файл');
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
    } catch (err) {
      console.error('Ошибка удаления файла', err);
      alert('Не удалось удалить файл');
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
          </>
        )}
      </div>

      {!collapsed && entries.map(file => (
        <div
          key={file.name}
          className={'fs-item' + (current?.name === file.name ? ' active' : '')}
          title={file.name}
          onClick={() => openFile(file)}
        >
          <span className="fs-name">
            {file.name}
            {unsaved && current?.name === file.name && ' ●'}
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
