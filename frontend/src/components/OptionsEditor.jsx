import { useState, useEffect } from 'react';

// Description of all settings + tips
const FIELDS = [

  {
    key: 'breaks',
    type: 'boolean',
    label: 'Breaks',
    hint: 'Ставить <br> при одинарных переводах строк.'
  },
  {
    key: 'gfm',
    type: 'boolean',
    label: 'GFM',
    hint: 'GitHub Flavored Markdown (таблицы, чекбоксы и т.д.).'
  },
  {
    key: 'pedantic',
    type: 'boolean',
    label: 'Pedantic',
    hint: 'Поведение, максимально совместимое с markdown.pl.'
  },
  {
    key: 'silent',
    type: 'boolean',
    label: 'Silent',
    hint: 'Не выбрасывать ошибки, а возвращать исходный Markdown.'
  }
];

export default function OptionsEditor({ value = {}, onChange }) {
  /* UI: form | json */
  const [mode, setMode] = useState('form');

  const [options, setOptions] = useState(value);

  const [raw, setRaw] = useState(JSON.stringify(value, null, 2));

  const [error, setError] = useState(null);

  useEffect(() => {
    setOptions(value);
    setRaw(JSON.stringify(value, null, 2));
  }, [value]);

  const pushChange = updated => {
    setOptions(updated);
    setRaw(JSON.stringify(updated, null, 2));
    onChange?.(updated);
  };

  const handleBoolean = (key, strVal) => {
    const updated = { ...options, [key]: strVal === 'true' };
    pushChange(updated);
  };

  const handleJsonField = (key, text) => {
    let newVal = null;
    if (text.trim().length) {
      try {
        newVal = JSON.parse(text);
      } catch (_) {
        return;
      }
    }
    const updated = { ...options, [key]: newVal };
    pushChange(updated);
  };

  const handleRawChange = e => {
    const txt = e.target.value;
    setRaw(txt);
    try {
      const parsed = JSON.parse(txt);
      setError(null);
      setOptions(parsed);
      onChange?.(parsed);
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="options-wrapper">
      {/* mode switch */}
      <div className="mode-switch">
        <button
          className={mode === 'form' ? 'tab active' : 'tab'}
          onClick={() => setMode('form')}
        >
          UI
        </button>
        <button
          className={mode === 'json' ? 'tab active' : 'tab'}
          onClick={() => setMode('json')}
        >
          JSON
        </button>
      </div>

      {mode === 'form' ? (
        <div className="options-form">
          {FIELDS.map(f => (
            <div className="option-row" key={f.key}>
              <label htmlFor={f.key}>{f.label}</label>

              {f.type === 'boolean' && (
                <select
                  id={f.key}
                  value={String(options[f.key])}
                  onChange={e => handleBoolean(f.key, e.target.value)}
                >
                  <option value="true">true</option>
                  <option value="false">false</option>
                </select>
              )}

              {f.type === 'json' && (
                <textarea
                  id={f.key}
                  rows={3}
                  placeholder="null"
                  value={
                    options[f.key] == null
                      ? ''
                      : JSON.stringify(options[f.key], null, 2)
                  }
                  onChange={e => handleJsonField(f.key, e.target.value)}
                />
              )}

              <p className="hint">{f.hint}</p>
            </div>
          ))}
        </div>
      ) : (
        <>
          <textarea
            className="editor"
            value={raw}
            onChange={handleRawChange}
            spellCheck={false}
          />
          {error && <p className="error">JSON error: {error}</p>}
        </>
      )}
    </div>
  );
}
