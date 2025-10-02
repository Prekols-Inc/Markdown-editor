import { useState, useEffect } from 'react';

// Only the four simple boolean options we decided to keep
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
    hint: 'GitHub Flavored Markdown (таблицы, чекбоксы и т.д.).'
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
  // local state of options
  const [options, setOptions] = useState(value);

  // If the parent passed new options, we update the local state.
  useEffect(() => {
    setOptions(value);
  }, [value]);

  // we update the local state and propagate the change outward
  const pushChange = updated => {
    setOptions(updated);
    onChange?.(updated);
  };

  // Boolean value switching processing
  const handleBoolean = (key, strVal) => {
    const updated = { ...options, [key]: strVal === 'true' };
    pushChange(updated);
  };

  return (
    <div className="options-wrapper">
      <div className="options-form">
        {FIELDS.map(f => (
          <div className="option-row" key={f.key}>
            <label htmlFor={f.key}>{f.label}</label>

            {/* boolean select */}
            <select
              id={f.key}
              value={String(options[f.key])}
              onChange={e => handleBoolean(f.key, e.target.value)}
            >
              <option value="true">true</option>
              <option value="false">false</option>
            </select>

            <p className="hint">{f.hint}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
