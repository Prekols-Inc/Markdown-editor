import { useMemo, useEffect } from 'react';
import { marked } from 'marked';
import { markedHighlight } from 'marked-highlight';
import hljs from 'highlight.js';
import DOMPurify from 'dompurify';

// Настраиваем marked с highlight.js для подсветки синтаксиса
marked.use(markedHighlight({
  langPrefix: 'hljs language-',
  highlight(code, lang) {
    const language = hljs.getLanguage(lang) ? lang : 'plaintext';
    return hljs.highlight(code, { language }).value;
  }
}));

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
