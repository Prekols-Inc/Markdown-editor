import { useEffect, useState } from 'react';
import { directoryOpen } from 'browser-fs-access';


export default function FileSidebar({
  current,
  onOpenFile,
  onSave,
  unsaved,
  setUnsaved,
  collapsed,
  onToggle
}) {
  const [dirHandle, setDirHandle] = useState(null);
  const [entries,   setEntries]   = useState([]);

  /* выбрать папку */
  async function openFolder() {
    try {
      const handle = await directoryOpen({ id: 'md-dir', recursive: true });
      setDirHandle(handle);
    } catch {}
  }

  /* прочитать .md-файлы */
  useEffect(() => {
    if (!dirHandle) return;

    (async () => {
      const iterable = dirHandle.values ? dirHandle.values() : dirHandle;
      const list = [];

      for await (const entry of iterable) {
        const file = entry.kind === 'file' ? await entry.getFile() : entry;
        if (/\\.(md|markdown|txt)$/i.test(file.name)) list.push({ source: entry, file });
      }
      setEntries(list);
    })();
  }, [dirHandle]);

  /* клик по файлу */
  async function clickFile(item) {
    const file = 'getFile' in item.source ? await item.source.getFile() : item.file;
    onOpenFile(await file.text(), item.source);
    setUnsaved(false);
  }

  /* сохранение */
  async function handleSave(ext) {
    await onSave(ext);
    setUnsaved(false);
  }

  return (
    <aside
      className={collapsed ? 'sidebar collapsed' : 'sidebar'}
      style={{ width: collapsed ? 48 : 260 }}
    >
      {/* toolbar */}
      <div className="toolbar">
        {/* кнопка-стрелка (видна всегда) */}
        <button
          className="btn secondary"
          onClick={onToggle}
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          style={{ width: 32 }}
        >
          {collapsed ? '»' : '«'}
        </button>

        {/* остальные действия показываем в развёрнутом виде */}
        {!collapsed && (
          <>
            <button className="btn" onClick={openFolder}>Open&nbsp;Folder</button>
            <button className="btn" disabled={!current} onClick={() => handleSave('md')}>Save&nbsp;.md</button>
            <button className="btn" disabled={!current} onClick={() => handleSave('html')}>Save&nbsp;.html</button>
          </>
        )}
      </div>

      {/* писок файлов */}
      {!collapsed && entries.map(e => (
        <div
          key={e.file.name}
          className={'fs-item' + (current?.name === e.file.name ? ' active' : '')}
          title={e.file.name}
          onClick={() => clickFile(e)}
        >
          {e.file.name}
          {unsaved && current?.name === e.file.name && ' ●'}
        </div>
      ))}
    </aside>
  );
}
