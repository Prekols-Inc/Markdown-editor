import { useMemo } from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';

marked.setOptions({
  breaks: false,
  gfm: true,
  headerIds: false
});

export default function MarkdownPreview({ markdown }) {
  const html = useMemo(() => {
    const raw = marked.parse(markdown);
    return DOMPurify.sanitize(raw);
  }, [markdown]);

  return (
    <div
      className="preview"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}