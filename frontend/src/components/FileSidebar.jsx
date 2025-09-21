import { useEffect, useState, forwardRef, useImperativeHandle } from 'react';
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

  const fetchFiles = async () => {
    try {
      const response = await API.STORAGE.get('/files');
      const files = response.data.files.map(name => ({ name }));
      setEntries(files);
    } catch (err) {
      console.error('Ошибка загрузки файлов', err);
    }
  };

  const clickFile = async (item) => {
    try {
      const response = await API.STORAGE.get(`/file/${encodeURIComponent(item.name)}`);
      onOpenFile(response.data, { name: item.name });
      setUnsaved(false);
    } catch (err) {
      console.error('Ошибка загрузки файла', err);
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
            <button
              className="btn"
              disabled={!current && !unsaved}
              onClick={() => onSave(fetchFiles)}
            >
              Save
            </button>
          </>
        )}
      </div>

      {!collapsed && entries.map(e => (
        <div
          key={e.name}
          className={'fs-item' + (current?.name === e.name ? ' active' : '')}
          title={e.name}
          onClick={() => clickFile(e)}
        >
          <span className="fs-name">
            {e.name}
            {unsaved && current?.name === e.name && ' ●'}
          </span>
          <button
            className="fs-close"
            title="Remove from sidebar"
            onClick={() =>
              setEntries(list => list.filter(x => x.name !== e.name))
            }
          >
            ×
          </button>
        </div>
      ))}
    </aside>
  );
});

export default FileSidebar;
