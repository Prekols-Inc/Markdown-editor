import { useEffect, useState } from 'react';
import API from '../API';

export default function FileSidebar({
  current,
  onOpenFile,
  onSave,
  unsaved,
  setUnsaved,
  collapsed = false,
  onToggle
}) {
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
      const response = await API.STORAGE.get(`/download/${encodeURIComponent(item.name)}`);
      // response.data — это текст файла
      onOpenFile(response.data, { name: item.name });
      setUnsaved(false);
    } catch (err) {
      console.error('Ошибка загрузки файла', err);
    }
  };

  useEffect(() => {
    fetchFiles();
  }, []);

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
            <button className="btn" disabled={!current} onClick={() => handleSave('md')}>Save .md</button>
            <button className="btn" disabled={!current} onClick={() => handleSave('html')}>Save .html</button>
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
          <span
            className="fs-name"
            title={e.name}
            onClick={() => clickFile(e)}
          >
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
}
