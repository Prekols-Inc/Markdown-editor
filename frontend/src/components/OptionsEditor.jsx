import { useState, useEffect } from 'react';

const DEFAULT_JSON = `{
    "async": false,
    "breaks": false,
    "extensions": null,
    "gfm": true,
    "hooks": null,
    "pedantic": false,
    "silent": false,
    "tokenizer": null,
    "walkTokens": null
}`;

export default function OptionsEditor({ value, onChange }) {
  const [text, setText] = useState(() => value ?? DEFAULT_JSON);
  const [error, setError] = useState(null);

  useEffect(() => {
    try {
      const obj = JSON.parse(text);
      setError(null);
      onChange(obj);
    } catch (e) {
      setError(e.message);
    }
  }, [text, onChange]);

  return (
    <div className="options-wrapper">
      <textarea
        className="editor"
        value={text}
        onChange={(e) => setText(e.target.value)}
        spellCheck={false}
      />
      {error && <p className="error">JSON error: {error}</p>}
    </div>
  );
}