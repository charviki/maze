import { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, Loader2, Wrench, FileText, ChevronDown, Brain } from 'lucide-react';
import {
  oracle,
  type ChatMessage,
  type ToolCallInfo,
  type ToolResultInfo,
  type DocContentInfo,
} from '@/api';
import { Skeleton } from '@maze/fabrication';

interface ChatPanelProps {
  className?: string;
}

interface DisplayMessage extends ChatMessage {
  id: string;
  thinking?: string;
  toolUse?: ToolCallInfo;
  toolResult?: ToolResultInfo;
  docContent?: DocContentInfo;
}

export default function ChatPanel({ className = '' }: ChatPanelProps) {
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    const prompt = input.trim();
    if (!prompt || loading) return;

    const userMsg: DisplayMessage = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: prompt,
    };

    const assistantMsg: DisplayMessage = {
      id: `assistant-${Date.now()}`,
      role: 'assistant',
      content: '',
    };

    setMessages((prev) => [...prev, userMsg, assistantMsg]);
    setInput('');
    setLoading(true);

    let aborted = false;

    await oracle.chat(prompt, {
      onThinking: (content) => {
        if (aborted) return;
        setMessages((prev) =>
          prev.map((m) => (m.id === assistantMsg.id ? { ...m, thinking: content } : m)),
        );
      },
      onText: (text) => {
        if (aborted) return;
        setMessages((prev) =>
          prev.map((m) => (m.id === assistantMsg.id ? { ...m, content: m.content + text } : m)),
        );
      },
      onToolUse: (data) => {
        if (aborted) return;
        setMessages((prev) =>
          prev.map((m) => (m.id === assistantMsg.id ? { ...m, toolUse: data } : m)),
        );
      },
      onToolResult: (data) => {
        if (aborted) return;
        setMessages((prev) =>
          prev.map((m) => (m.id === assistantMsg.id ? { ...m, toolResult: data } : m)),
        );
      },
      onDocContent: (data) => {
        if (aborted) return;
        setMessages((prev) =>
          prev.map((m) => (m.id === assistantMsg.id ? { ...m, docContent: data } : m)),
        );
      },
      onDone: () => {
        setLoading(false);
      },
      onError: (error) => {
        aborted = true;
        setMessages((prev) =>
          prev.map((m) => (m.id === assistantMsg.id ? { ...m, content: `Error: ${error}` } : m)),
        );
        setLoading(false);
      },
    });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void handleSend();
    }
  };

  return (
    <div className={`flex flex-col h-full ${className}`}>
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border">
        <Bot size={14} className="text-primary" />
        <span className="text-foreground font-mono text-sm">ORACLE</span>
        {loading && <Loader2 size={12} className="text-primary animate-spin ml-auto" />}
      </div>

      <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full text-muted-foreground font-mono text-sm gap-2">
            <Bot size={32} className="text-primary/30" />
            <p>Ask the Oracle about your knowledge base</p>
          </div>
        )}
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            {msg.role === 'assistant' && (
              <div className="w-6 h-6 rounded bg-primary/10 flex items-center justify-center shrink-0 mt-1">
                <Bot size={12} className="text-primary" />
              </div>
            )}
            <div
              className={`max-w-[80%] rounded px-3 py-2 text-sm font-mono ${
                msg.role === 'user'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-card border border-border text-foreground'
              }`}
            >
              {msg.thinking && (
                <div className="flex items-center gap-1 text-xs text-purple-400 mb-2">
                  <Brain size={10} />
                  <span className="italic">{msg.thinking}</span>
                </div>
              )}
              {msg.toolUse && (
                <div className="flex items-center gap-1 text-xs text-blue-400 mb-2 bg-blue-500/10 px-2 py-1 rounded">
                  <Wrench size={10} />
                  <span>{msg.toolUse.name}</span>
                  {msg.toolUse.input && (
                    <details className="ml-1">
                      <summary className="cursor-pointer">
                        <ChevronDown size={8} />
                      </summary>
                      <pre className="mt-1 whitespace-pre-wrap text-[10px]">
                        {msg.toolUse.input}
                      </pre>
                    </details>
                  )}
                </div>
              )}
              {msg.docContent && (
                <div className="flex items-center gap-1 text-xs text-emerald-400 mb-2 bg-emerald-500/10 px-2 py-1 rounded">
                  <FileText size={10} />
                  <span>{msg.docContent.title}</span>
                </div>
              )}
              {msg.content ? (
                <div className="whitespace-pre-wrap">{msg.content}</div>
              ) : loading && msg.role === 'assistant' ? (
                <Skeleton className="h-4 w-20" />
              ) : null}
            </div>
            {msg.role === 'user' && (
              <div className="w-6 h-6 rounded bg-blue-400/10 flex items-center justify-center shrink-0 mt-1">
                <User size={12} className="text-blue-400" />
              </div>
            )}
          </div>
        ))}
      </div>

      <div className="px-4 py-3 border-t border-border">
        <div className="flex items-center gap-2">
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask the Oracle..."
            disabled={loading}
            className="flex-1 bg-background border border-border rounded px-3 py-2 text-foreground font-mono text-sm placeholder:text-muted-foreground focus:outline-none focus:border-primary transition disabled:opacity-50"
          />
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            className="p-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send size={14} />
          </button>
        </div>
      </div>
    </div>
  );
}
