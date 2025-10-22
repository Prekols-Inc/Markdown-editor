import { useEffect, useState, forwardRef, useImperativeHandle } from 'react';
import API from '../API';

const FileSidebar = forwardRef(function FileSidebar(
  {
    current,
    onOpenFile,
    unsaved,
    setUnsaved,
    collapsed = false,
  },
  ref
) {
  const [entries, setEntries] = useState([]);

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
      } else {
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

  useImperativeHandle(ref, () => ({
    refresh: fetchFiles,
  }));

  return (
    <aside
      className={collapsed ? 'sidebar collapsed' : 'sidebar'}
    >
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
    </aside>
  );
});

export default FileSidebar;
