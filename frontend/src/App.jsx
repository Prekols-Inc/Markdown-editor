import { useState, useEffect, useCallback } from 'react';
import MarkdownEditor from './components/MarkdownEditor';
import FileSidebar from './components/FileSidebar';
import OptionsEditor from './components/OptionsEditor';
import MarkdownPreview from './components/MarkdownPreview';
import { marked } from 'marked';
import { fileSave } from 'browser-fs-access';

const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.
`;

const DEFAULT_OPTIONS = {
  async: false,
  breaks: false,
  extensions: null,
  gfm: true,
  hooks: null,
  pedantic: false,
  silent: false,
  tokenizer: null,
  walkTokens: null
};

export default function App() {
  /* Markdown */
  const [markdown, setMarkdown] = useState(
    () => localStorage.getItem('md-draft') ?? DEFAULT_MD
  );

  // автосохранение черновика
  useEffect(() => {
    const id = setTimeout(() => localStorage.setItem('md-draft', markdown), 400);
    return () => clearTimeout(id);
  }, [markdown]);

  /* Options */
  const [options, setOptions] = useState(() => {
    try {
      const stored = JSON.parse(localStorage.getItem('md-options'));
      if (!stored || typeof stored !== 'object' || !Object.keys(stored).length) {
        return DEFAULT_OPTIONS;
      }
      return stored;
    } catch {
      return DEFAULT_OPTIONS;
    }
  });

  const handleOptionsChange = useCallback((obj) => {
    setOptions(obj);
    localStorage.setItem('md-options', JSON.stringify(obj));
  }, []);

  /* Tabs */
  const [tab, setTab] = useState('markdown');

  /* File handling */
  const [fileHandle, setFileHandle] = useState(null);
  const [unsaved, setUnsaved] = useState(false);

  useEffect(() => {
    if (fileHandle) setUnsaved(true);
  }, [markdown, fileHandle]);

  const handleOpenFile = useCallback((text, handle) => {
    setMarkdown(text);
    setFileHandle(handle);
    setUnsaved(false);
  }, []);

  const handleSave = useCallback(
    async (format) => {
      try {
        let blob;
        if (format === 'md') {
          blob = new Blob([markdown], { type: 'text/markdown' });
        } else {
          const html = marked.parse(markdown, options);
          blob = new Blob([html], { type: 'text/html' });
        }

        await fileSave(
          blob,
          {
            fileName: fileHandle?.name
              ? fileHandle.name.replace(/\.(md|markdown|txt)$/i, `.${format}`)
              : `untitled.${format}`,
            extensions: [`.${format}`],
          },
          fileHandle
        );
        setUnsaved(false);
      } catch (_) {
      }
    },
    [markdown, options, fileHandle]
  );

  return (
    <div className="app-grid">
      {/* Sidebar with file explorer */}
      <FileSidebar
        current={fileHandle}
        onOpenFile={handleOpenFile}
        onSave={handleSave}
        unsaved={unsaved}
        setUnsaved={setUnsaved}
      />

      {/* Tabs + editors */}
      <div className="left-panel">
        <div className="tabs">
          <button
            className={tab === 'markdown' ? 'tab active' : 'tab'}
            onClick={() => setTab('markdown')}
          >
            Markdown
          </button>
          <button
            className={tab === 'options' ? 'tab active' : 'tab'}
            onClick={() => setTab('options')}
          >
            Options
          </button>
        </div>

        {tab === 'markdown' ? (
          <MarkdownEditor value={markdown} onChange={setMarkdown} />
        ) : (
          <OptionsEditor
            value={JSON.stringify(options, null, 2)}
            onChange={handleOptionsChange}
          />
        )}
      </div>

      {/* Preview */}
      <MarkdownPreview markdown={markdown} options={options} />
    </div>
  );
}
