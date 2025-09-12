import { useState, useEffect, useCallback } from 'react';
import MarkdownEditor from './components/MarkdownEditor';
import OptionsEditor from './components/OptionsEditor';
import MarkdownPreview from './components/MarkdownPreview';

const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.
`;

export default function App() {
  const [markdown, setMarkdown] = useState(
    () => localStorage.getItem('md-draft') ?? DEFAULT_MD
  );

  useEffect(() => {
    const id = setTimeout(() => localStorage.setItem('md-draft', markdown), 400);
    return () => clearTimeout(id);
  }, [markdown]);

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

  const [tab, setTab] = useState('markdown');

  return (
    <div className="app-grid">
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
          <OptionsEditor value={JSON.stringify(options, null, 2)} onChange={handleOptionsChange} />
        )}
      </div>

      <MarkdownPreview markdown={markdown} options={options} />
    </div>
  );
}
