import { useMemo } from 'react';
import { marked, Renderer } from 'marked';
import markedKatex from 'marked-katex-extension';
import DOMPurify from 'dompurify';
import 'katex/dist/katex.min.css';

const renderer = new Renderer();
renderer.link = ({ href, title, text }) => {
  const titleAttr = title ? ` title="${title}"` : '';
  return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`;
};

marked.use(markedKatex({ throwOnError: false }));
marked.setOptions({ gfm: true, breaks: true, renderer });

interface MarkdownContentProps {
  content: string;
  className?: string;
}

// Renders LLM output as markdown. Plain text (no markdown syntax) renders
// identically to before — this only adds structure when the model actually
// uses it (code fences, lists, tables, etc.). Styling for the parsed HTML
// lives in the .markdown-body rules in style.css since we render raw HTML
// rather than per-element React components.
export default function MarkdownContent({ content, className = '' }: MarkdownContentProps) {
  const html = useMemo(() => {
    const rawHtml = marked.parse(content, { async: false }) as string;
    return DOMPurify.sanitize(rawHtml, {
      ADD_ATTR: ['target', 'rel'],
      // KaTeX emits MathML alongside its visual HTML spans for accessibility.
      ADD_TAGS: [
        'math', 'semantics', 'annotation', 'mrow', 'mi', 'mo', 'mn', 'ms',
        'msup', 'msub', 'msubsup', 'mfrac', 'msqrt', 'mroot', 'mtext',
        'mspace', 'mtable', 'mtr', 'mtd', 'menclose', 'mpadded', 'mphantom',
        'mstyle', 'merror', 'munderover', 'munder', 'mover',
      ],
    });
  }, [content]);

  return (
    <div
      className={`markdown-body text-sm break-words ${className}`}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
