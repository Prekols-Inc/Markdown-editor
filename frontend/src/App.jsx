import { useState, useEffect } from 'react';
import MarkdownEditor from './components/MarkdownEditor';
import MarkdownPreview from './components/MarkdownPreview';

const DEFAULT_MD = `# Marked - Markdown Parser

> Введите Markdown слева — результат увидите справа.

- Поддержка \`**жирного**\`, \`_курсива_\`, списков, заголовков и т.д.
- Безопасность обеспечивается **DOMPurify**.

\`\`\`js
// Пример кода
function hello() {
  console.log('Привет, Markdown!');
}
\`\`\`
`;

export default function App() {
  const [markdown, setMarkdown] = useState(
    () => localStorage.getItem('md-draft') ?? DEFAULT_MD
  );

  // Автосохранение в localStorage
  useEffect(() => {
    const id = setTimeout(() => localStorage.setItem('md-draft', markdown), 400);
    return () => clearTimeout(id);
  }, [markdown]);

  return (
    <div className="app-grid">
      <MarkdownEditor value={markdown} onChange={setMarkdown} />
      <MarkdownPreview markdown={markdown} />
    </div>
  );
}