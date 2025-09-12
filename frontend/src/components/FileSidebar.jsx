import { useEffect, useState } from 'react';
import { fileOpen, directoryOpen, fileSave } from 'browser-fs-access';

export default function FileSidebar({
  current,
  onOpenFile,
  onSave,
  unsaved,
  setUnsaved
}) {
  const [dirHandle, setDirHandle] = useState(null);
  const [entries, setEntries] = useState([]);

  /* выбрать папку */
  async function openFolder() {
    try {
      const handle = await directoryOpen({ id: 'md-dir', recursive: true });
      setDirHandle(handle);
    } catch (_) { }
  }

  /* читаем список md-файлов */
  useEffect(() => {
    if (!dirHandle?.values) return;
    (async () => {
        const tmp = [];
        const iterable = dirHandle.values ? dirHandle.values() : dirHandle;

        for await (const entry of iterable) {
            const fileLike = entry.kind === 'file' ? await entry.getFile() : entry;
            if (fileLike && /\.(md|markdown|txt)$/i.test(fileLike.name)) {
                tmp.push({ source: entry, file: fileLike });
            }
        }
        setEntries(tmp);
    })();
  }, [dirHandle]);

  /* выбрать файл из списка */
  async function clickFile(item) {
    const file = 'getFile' in item.source ? await item.source.getFile() : item.file;
    const text = await file.text();
    onOpenFile(text, item.source);
    setUnsaved(false);
  }

  /* сохранить */
  async function handleSave(format) {
    await onSave(format);
    setUnsaved(false);
  }

  return (
    <div className="sidebar">
      <div className="toolbar">
        <button onClick={openFolder}>Open Folder</button>
        <button disabled={!current} onClick={() => handleSave('md')}>
          Save .md
        </button>
        <button disabled={!current} onClick={() => handleSave('html')}>
          Save .html
        </button>
      </div>

      {entries.map((e) => (
        <div
            key={e.file.name}
            className={
                'fs-item' + (current?.name === e.file.name ? ' active' : '')
          }
          title={e.file.name}
          onClick={() => clickFile(e)}
        >
            {e.file.name}
            {unsaved && current?.name === e.file.name && '  ●'}
        </div>
      ))}
    </div>
  );
}
