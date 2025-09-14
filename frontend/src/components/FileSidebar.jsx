import { useEffect, useState } from 'react';
import { directoryOpen, fileOpen } from 'browser-fs-access';

export default function FileSidebar({
  current,
  onOpenFile,
  onSave,
  unsaved,
  setUnsaved,
  collapsed = false,
  onToggle
}) {
  const [dirHandle, setDirHandle] = useState(null);
  const [entries,   setEntries]   = useState([]);


  /* Open folder */
  async function openFolder() {
    try {
      const handle = await directoryOpen({ id: 'md-dir', recursive: true });
      setDirHandle(handle);
    } catch {
      try {
        const file = await fileOpen({
          id: 'md-single',
          mimeTypes: ['text/markdown', 'text/plain'],
          extensions: ['.md', '.markdown', '.txt']
        });
        setDirHandle([file]);
      } catch {}
    }
  }

  /* Open single file */
  async function openSingleFile() {
    try {
    const file = await fileOpen({
    id: 'md-single',
    mimeTypes: ['text/markdown', 'text/plain'],
    extensions: ['.md', '.markdown', '.txt']
    });
    onOpenFile(await file.text(), file);
    setUnsaved(false);
    setEntries(list => {
      const exists = list.some(e => e.file.name === file.name);
      if (exists) return list;
      return [...list, { source: file, file }];
    });
    setDirHandle(null);
    } catch {}
    }

  useEffect(() => {
    if (!dirHandle) return;

    (async () => {
      const iterable = dirHandle.values ? dirHandle.values() : dirHandle;
      const list = [];
      for await (const entry of iterable) {
        const file = entry.kind === 'file' ? await entry.getFile() : entry;
        if (/\.(md|markdown|txt)$/i.test(file.name)) list.push({ source: entry, file });
      }
      setEntries(list);
    })();
  }, [dirHandle]);

  async function clickFile(item) {
    const file = 'getFile' in item.source ? await item.source.getFile() : item.file;
    onOpenFile(await file.text(), item.source);
    setUnsaved(false);
  }

  /* save */
  async function handleSave(ext) {
    await onSave(ext);
    setUnsaved(false);
  }

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
            <button className="btn" onClick={openFolder}>Open Folder</button>
            <button className="btn" onClick={openSingleFile}>Open File</button>
            <button className="btn" disabled={!current} onClick={() => handleSave('md')}>Save .md</button>
            <button className="btn" disabled={!current} onClick={() => handleSave('html')}>Save .html</button>
          </>
        )}
      </div>

      {!collapsed && entries.map(e => (
        <div
          key={e.file.name}
          className={'fs-item' + (current?.name === e.file.name ? ' active' : '')}
          title={e.file.name}
          onClick={() => clickFile(e)}
        >
          <span
            className="fs-name"
            title={e.file.name}
            onClick={() => clickFile(e)}
          >
            {e.file.name}
            {unsaved && current?.name === e.file.name && ' ●'}
          </span>
          <button
            className="fs-close"
            title="Remove from sidebar"
            onClick={() =>
              setEntries(list => list.filter(x => x.file.name !== e.file.name))
            }
          >
            ×
          </button>
        </div>
      ))}
    </aside>
  );
}