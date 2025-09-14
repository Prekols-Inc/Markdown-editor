export default function MarkdownEditor({ value, onChange }) {
  return (
    <textarea
      className="editor"
      value={value}
      onChange={e => onChange(e.target.value)}
      spellCheck={false}
    />
  );
}