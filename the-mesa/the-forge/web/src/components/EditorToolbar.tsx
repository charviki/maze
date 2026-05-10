import {
  Bold,
  Italic,
  Link,
  List,
  ListOrdered,
  Heading1,
  Heading2,
  Heading3,
  Code,
  Quote,
  Image,
  Paperclip,
} from 'lucide-react';

interface EditorToolbarProps {
  onInsert: (before: string, after?: string) => void;
  onAttach?: () => void;
}

const tools = [
  { icon: Heading1, label: 'H1', before: '# ', after: '' },
  { icon: Heading2, label: 'H2', before: '## ', after: '' },
  { icon: Heading3, label: 'H3', before: '### ', after: '' },
  { icon: Bold, label: 'Bold', before: '**', after: '**' },
  { icon: Italic, label: 'Italic', before: '_', after: '_' },
  { icon: Code, label: 'Code', before: '`', after: '`' },
  { icon: Link, label: 'Link', before: '[', after: '](url)' },
  { icon: Image, label: 'Image', before: '![', after: '](url)' },
  { icon: List, label: 'UL', before: '- ', after: '' },
  { icon: ListOrdered, label: 'OL', before: '1. ', after: '' },
  { icon: Quote, label: 'Quote', before: '> ', after: '' },
];

export default function EditorToolbar({ onInsert, onAttach }: EditorToolbarProps) {
  return (
    <div className="flex items-center gap-0.5 px-2 py-1.5 bg-card border-b border-border">
      {tools.map((tool) => (
        <button
          key={tool.label}
          onClick={() => onInsert(tool.before, tool.after)}
          className="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded transition-colors"
          title={tool.label}
        >
          <tool.icon size={14} />
        </button>
      ))}
      {onAttach && (
        <>
          <div className="w-px h-4 bg-border mx-1" />
          <button
            onClick={onAttach}
            className="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded transition-colors"
            title="Attach"
          >
            <Paperclip size={14} />
          </button>
        </>
      )}
    </div>
  );
}
