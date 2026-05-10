import { useRef, useEffect } from 'react';
import { EditorView } from '@codemirror/view';
import { EditorState } from '@codemirror/state';
import { markdown, markdownLanguage } from '@codemirror/lang-markdown';
import { oneDark } from '@codemirror/theme-one-dark';
import { basicSetup } from 'codemirror';

interface CodeMirrorEditorProps {
  value: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  autoFocus?: boolean;
  className?: string;
}

export default function CodeMirrorEditor({
  value,
  onChange,
  autoFocus = false,
  className = '',
}: CodeMirrorEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const isExternalUpdate = useRef(false);

  useEffect(() => {
    if (!containerRef.current) return;

    const updateListener = EditorView.updateListener.of((update) => {
      if (update.docChanged && !isExternalUpdate.current) {
        onChange?.(update.state.doc.toString());
      }
    });

    const customTheme = EditorView.theme({
      '&': {
        fontSize: '13px',
        backgroundColor: 'transparent',
      },
      '.cm-content': {
        fontFamily: '"JetBrains Mono", "Fira Code", monospace',
        caretColor: 'hsl(var(--primary))',
        padding: '8px 0',
      },
      '.cm-cursor': {
        borderLeftColor: 'hsl(var(--primary))',
        borderLeftWidth: '2px',
      },
      '.cm-activeLine': {
        backgroundColor: 'hsl(var(--primary) / 0.05)',
      },
      '.cm-activeLineGutter': {
        backgroundColor: 'hsl(var(--primary) / 0.05)',
      },
      '.cm-gutters': {
        backgroundColor: 'transparent',
        borderRight: '1px solid hsl(var(--border))',
        color: 'hsl(var(--muted-foreground))',
        fontFamily: 'monospace',
        fontSize: '10px',
      },
      '.cm-lineNumbers .cm-gutterElement': {
        minWidth: '2.5em',
      },
      '.cm-selectionBackground, &.cm-focused .cm-selectionBackground': {
        background: 'hsl(var(--primary) / 0.15) !important',
      },
      '.cm-matchingBracket': {
        backgroundColor: 'hsl(var(--primary) / 0.25)',
        color: 'hsl(var(--foreground)) !important',
      },
      '.cm-searchMatch': {
        backgroundColor: 'hsl(var(--primary) / 0.2)',
      },
      '.cm-searchMatch.cm-searchMatch-selected': {
        backgroundColor: 'hsl(var(--primary) / 0.35)',
      },
      '.cm-placeholder': {
        color: 'hsl(var(--muted-foreground))',
        fontFamily: 'monospace',
        fontStyle: 'italic',
      },
    });

    const state = EditorState.create({
      doc: value,
      extensions: [
        basicSetup,
        markdown({ base: markdownLanguage }),
        oneDark,
        customTheme,
        updateListener,
        EditorView.lineWrapping,
      ],
    });

    const view = new EditorView({
      state,
      parent: containerRef.current,
    });

    viewRef.current = view;

    if (autoFocus) {
      setTimeout(() => view.focus(), 50);
    }

    return () => {
      view.destroy();
      viewRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const view = viewRef.current;
    if (!view) return;

    const currentDoc = view.state.doc.toString();
    if (currentDoc !== value) {
      isExternalUpdate.current = true;
      view.dispatch({
        changes: { from: 0, to: currentDoc.length, insert: value },
      });
      isExternalUpdate.current = false;
    }
  }, [value]);

  return <div ref={containerRef} className={`h-full overflow-auto ${className}`} />;
}
