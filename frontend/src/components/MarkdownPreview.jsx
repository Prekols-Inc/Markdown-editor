import { useMemo } from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';

export default function MarkdownPreview({ markdown, options = {} }) {
  const html = useMemo(() => {
    try {
      const raw = marked.parse(markdown, options);
      return DOMPurify.sanitize(raw);
    } catch (e) {
      return `<pre style="color:red;">Render error: ${e.message}</pre>`;
    }
  }, [markdown, options]);

  return (
    <div
      className="preview"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
