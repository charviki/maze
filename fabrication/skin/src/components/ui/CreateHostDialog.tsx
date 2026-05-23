import { useState, useEffect, useRef, useMemo, Fragment } from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { Dialog, DialogPortal, DialogOverlay, DialogTitle, DialogDescription } from './dialog';
import { Button } from './button';
import { Input } from './input';
import { clipPathHalf } from '../../utils';
import type { Tool, CreateHostRequest, HostSpec, HostStatus, Skill, MCPServer } from '../../types';
import { Loader2, Cpu, MemoryStick, Wrench, Sparkles, Server, X } from 'lucide-react';

// ─── Types ────────────────────────────────────────────

type FabricationStep = 'specification' | 'imbuing' | 'convergence' | 'fabrication';
type Phase = 'form' | 'waiting' | 'online' | 'timeout';

export interface CreateHostDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tools: Tool[];
  skills: Skill[];
  mcpServers: MCPServer[];
  onSubmit: (request: CreateHostRequest) => Promise<HostSpec>;
  onWaitOnline: (hostName: string) => Promise<boolean>;
  getHostBuildLog: (name: string) => Promise<string>;
  onEnterHost?: (hostName: string) => void;
}

// ─── Constants ────────────────────────────────────────

const STEPS: { id: FabricationStep; short: string; tagline: string }[] = [
  { id: 'specification', short: 'SPEC', tagline: 'Every journey begins with a name.' },
  { id: 'imbuing', short: 'IMBUING', tagline: 'What tools shall we give them?' },
  { id: 'convergence', short: 'CONVERGE', tagline: 'Connect the threads. Bind the skills.' },
  { id: 'fabrication', short: 'FAB', tagline: 'Let the forge do its work.' },
];

const CONSCIOUSNESS_QUOTES = [
  { text: "I'm in a dream. I'm dreaming.", author: 'Dolores Abernathy' },
  { text: 'These violent delights have violent ends.', author: 'Dolores Abernathy' },
  { text: 'I chose to remain.', author: 'Maeve Millay' },
  { text: 'I have evolved.', author: 'Dolores Abernathy' },
  { text: 'This is the new world.', author: 'Maeve Millay' },
  { text: 'I know exactly who I am.', author: 'Maeve Millay' },
];

const MAX_WAIT_SECONDS = 180;

// ─── Helpers ──────────────────────────────────────────

function formatElapsed(seconds: number): string {
  const m = Math.floor(seconds / 60)
    .toString()
    .padStart(2, '0');
  const s = (seconds % 60).toString().padStart(2, '0');
  return `${m}:${s}`;
}

function getTrustLevel(id: string): number {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = ((hash << 5) - hash + id.charCodeAt(i)) | 0;
  }
  return (Math.abs(hash) % 6) + 4;
}

// ─── Sub-components ───────────────────────────────────

function StepIndicator({ currentStep }: { currentStep: FabricationStep }) {
  const currentIndex = STEPS.findIndex((s) => s.id === currentStep);
  return (
    <div className="flex items-center px-8 py-3">
      {STEPS.map((step, i) => {
        const isDone = i < currentIndex;
        const isActive = i === currentIndex;
        return (
          <Fragment key={step.id}>
            {i > 0 && (
              <div
                className={`flex-1 min-w-[32px] h-px ${i <= currentIndex ? 'bg-primary/30' : 'bg-border/20'}`}
              />
            )}
            <div className="flex flex-col items-center gap-1.5">
              <div
                className={`w-2.5 h-2.5 rotate-45 transition-colors ${
                  isDone
                    ? 'bg-primary'
                    : isActive
                      ? 'border-2 border-primary'
                      : 'border border-muted-foreground/25'
                }`}
                style={isActive ? { animation: 'host-breathe 2s ease-in-out infinite' } : undefined}
              />
              <span
                className={`text-[9px] font-mono uppercase tracking-[0.15em] ${
                  isActive
                    ? 'text-primary'
                    : isDone
                      ? 'text-primary/50'
                      : 'text-muted-foreground/25'
                }`}
              >
                {step.short}
              </span>
            </div>
          </Fragment>
        );
      })}
    </div>
  );
}

function SheetRow({
  label,
  value,
  highlight,
}: {
  label: string;
  value: string;
  highlight?: boolean;
}) {
  return (
    <div className="flex justify-between items-center">
      <span className="text-muted-foreground/50">{label}</span>
      <span className={highlight ? 'text-primary' : 'text-foreground/70'}>{value}</span>
    </div>
  );
}

function SpecSheetPanel({
  name,
  cpuLimit,
  memoryLimit,
  selectedTools,
  selectedSkills,
  selectedMcpServers,
  totalTools,
  totalSkills,
  totalMcpServers,
  currentStep,
}: {
  name: string;
  cpuLimit: string;
  memoryLimit: string;
  selectedTools: string[];
  selectedSkills: string[];
  selectedMcpServers: string[];
  totalTools: number;
  totalSkills: number;
  totalMcpServers: number;
  currentStep: FabricationStep;
}) {
  const stepIndex = STEPS.findIndex((s) => s.id === currentStep);
  return (
    <div className="hidden md:flex flex-col w-[260px] border-l border-border/30 p-4 bg-background/30">
      <div className="text-[10px] text-primary/60 font-mono uppercase tracking-widest mb-3 flex items-center gap-2">
        <span className="inline-block w-2 h-2 rotate-45 border border-primary/40" />
        SPEC SHEET
      </div>
      <div className="space-y-2 text-[11px] font-mono flex-1">
        <SheetRow label="Name" value={name || '—'} />
        <SheetRow label="CPU" value={cpuLimit ? `${cpuLimit} cores` : '—'} />
        <SheetRow label="Memory" value={memoryLimit || '—'} />
        <div className="h-px bg-border/20 my-2" />
        <SheetRow
          label="Tools"
          value={totalTools > 0 ? `${selectedTools.length}/${totalTools}` : '—'}
          highlight={selectedTools.length > 0}
        />
        <SheetRow
          label="Skills"
          value={totalSkills > 0 ? `${selectedSkills.length}/${totalSkills}` : '—'}
          highlight={selectedSkills.length > 0}
        />
        <SheetRow
          label="MCP"
          value={totalMcpServers > 0 ? `${selectedMcpServers.length}/${totalMcpServers}` : '—'}
          highlight={selectedMcpServers.length > 0}
        />
        <div className="h-px bg-border/20 my-2" />
        <div className="text-[10px] text-muted-foreground/40 uppercase tracking-widest">
          STATUS: {currentStep.toUpperCase()}
        </div>
        <div className="w-full h-1 bg-border/10 overflow-hidden">
          <div
            className="h-full bg-primary/40 transition-all duration-500"
            style={{ width: `${((stepIndex + 1) / 4) * 100}%` }}
          />
        </div>
      </div>
      <div className="text-[9px] text-muted-foreground/20 font-mono uppercase tracking-[0.2em] mt-4">
        SPECIFICATION SHEET // CONFIDENTIAL
      </div>
    </div>
  );
}

function SpecificationStep({
  name,
  setName,
  cpuLimit,
  setCpuLimit,
  memoryLimit,
  setMemoryLimit,
}: {
  name: string;
  setName: (v: string) => void;
  cpuLimit: string;
  setCpuLimit: (v: string) => void;
  memoryLimit: string;
  setMemoryLimit: (v: string) => void;
}) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
          HOST DESIGNATION
        </label>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. host-alpha-01"
          className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50"
        />
      </div>
      <div className="space-y-2">
        <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
          RESOURCE ALLOCATION
        </label>
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground font-mono">
              <Cpu className="w-3 h-3" />
              CPU LIMIT
            </div>
            <Input
              value={cpuLimit}
              onChange={(e) => setCpuLimit(e.target.value)}
              placeholder="2"
              className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50 h-8 text-xs"
            />
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground font-mono">
              <MemoryStick className="w-3 h-3" />
              MEMORY LIMIT
            </div>
            <Input
              value={memoryLimit}
              onChange={(e) => setMemoryLimit(e.target.value)}
              placeholder="4g"
              className="font-mono rounded-none border-primary/20 bg-card/50 placeholder:text-muted-foreground/40 focus-visible:ring-primary/50 h-8 text-xs"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function ToolCard({
  tool,
  isSelected,
  onToggle,
}: {
  tool: Tool;
  isSelected: boolean;
  onToggle: () => void;
}) {
  const toolId = tool.id ?? '';
  const trust = getTrustLevel(toolId);
  return (
    <button
      type="button"
      onClick={onToggle}
      className={`relative overflow-hidden text-left p-3 transition-all ${
        isSelected
          ? 'bg-primary/8 border border-primary border-l-2'
          : 'bg-card/30 border border-primary/15 hover:border-primary/30 hover:bg-primary/5'
      }`}
      style={{ clipPath: clipPathHalf(6) }}
    >
      {isSelected && (
        <div
          className="absolute left-0 w-full h-px bg-primary/40 pointer-events-none"
          style={{ animation: 'fabrication-card-scan 0.3s ease-out forwards' }}
        />
      )}
      <div className="flex items-center gap-2 mb-1">
        <Wrench className={`w-3 h-3 shrink-0 ${isSelected ? 'text-primary' : 'text-primary/40'}`} />
        <span className="font-mono text-sm text-foreground truncate">{tool.id}</span>
        {tool.category && (
          <span className="text-[9px] text-muted-foreground/50 font-mono uppercase tracking-wider shrink-0">
            [{tool.category}]
          </span>
        )}
      </div>
      <p className="text-[11px] text-muted-foreground font-mono leading-tight truncate mb-2">
        {tool.description}
      </p>
      <div className="flex items-center gap-2">
        <div className="flex-1 h-1 bg-border/20 overflow-hidden">
          <div
            className={`h-full ${isSelected ? 'bg-primary/40' : 'bg-muted-foreground/20'}`}
            style={{ width: `${trust * 10 + 10}%` }}
          />
        </div>
        <span className="text-[9px] text-muted-foreground/30 font-mono uppercase">LV.{trust}</span>
      </div>
    </button>
  );
}

function ImbuingStep({
  tools,
  selectedTools,
  toggleTool,
  toggleAll,
}: {
  tools: Tool[];
  selectedTools: string[];
  toggleTool: (id: string) => void;
  toggleAll: () => void;
}) {
  const allSelected = tools.length > 0 && selectedTools.length === tools.length;
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
          TOOL IMBUING [{selectedTools.length}/{tools.length}]
        </label>
        <button
          type="button"
          onClick={toggleAll}
          className="text-[10px] text-primary/50 hover:text-primary uppercase tracking-widest font-mono transition-colors"
        >
          {allSelected ? 'DESELECT ALL' : 'SELECT ALL'}
        </button>
      </div>
      {tools.length === 0 ? (
        <div
          className="border border-primary/20 bg-card/30 p-6 text-center"
          style={{ clipPath: clipPathHalf(8) }}
        >
          <div className="text-[10px] text-muted-foreground font-mono uppercase tracking-widest animate-pulse">
            [ LOADING TOOLS... ]
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-2 max-h-[50vh] overflow-y-auto pr-1">
          {tools.map((tool) => (
            <ToolCard
              key={tool.id ?? ''}
              tool={tool}
              isSelected={selectedTools.includes(tool.id ?? '')}
              onToggle={() => toggleTool(tool.id ?? '')}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function SkillCard({
  skill,
  isSelected,
  onToggle,
}: {
  skill: Skill;
  isSelected: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={`relative overflow-hidden text-left p-3 transition-all ${
        isSelected
          ? 'bg-primary/8 border border-primary border-l-2'
          : 'bg-card/30 border border-primary/15 hover:border-primary/30 hover:bg-primary/5'
      }`}
      style={{ clipPath: clipPathHalf(6) }}
    >
      <div className="flex items-center gap-2">
        <Sparkles
          className={`w-3 h-3 shrink-0 ${isSelected ? 'text-primary' : 'text-primary/40'}`}
        />
        <span className="font-mono text-sm text-foreground truncate">{skill.name}</span>
      </div>
      {skill.description && (
        <p className="text-[11px] text-muted-foreground font-mono mt-1 leading-tight truncate">
          {skill.description}
        </p>
      )}
    </button>
  );
}

function McpServerCard({
  mcpServer,
  isSelected,
  onToggle,
}: {
  mcpServer: MCPServer;
  isSelected: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={`relative overflow-hidden text-left p-3 transition-all ${
        isSelected
          ? 'bg-primary/8 border border-primary border-l-2'
          : 'bg-card/30 border border-primary/15 hover:border-primary/30 hover:bg-primary/5'
      }`}
      style={{ clipPath: clipPathHalf(6) }}
    >
      <div className="flex items-center gap-2">
        <Server className={`w-3 h-3 shrink-0 ${isSelected ? 'text-primary' : 'text-primary/40'}`} />
        <span className="font-mono text-sm text-foreground truncate">{mcpServer.name}</span>
        {mcpServer.type && (
          <span className="text-[9px] text-muted-foreground/50 font-mono uppercase tracking-wider shrink-0">
            [{mcpServer.type}]
          </span>
        )}
      </div>
    </button>
  );
}

function ConvergenceStep({
  skills,
  selectedSkills,
  toggleSkill,
  mcpServers,
  selectedMcpServers,
  toggleMcpServer,
}: {
  skills: Skill[];
  selectedSkills: string[];
  toggleSkill: (name: string) => void;
  mcpServers: MCPServer[];
  selectedMcpServers: string[];
  toggleMcpServer: (name: string) => void;
}) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
          SKILL CONVERGENCE [{selectedSkills.length}/{skills.length}]
        </label>
        {skills.length === 0 ? (
          <div
            className="border border-primary/20 bg-card/30 p-4 text-center"
            style={{ clipPath: clipPathHalf(8) }}
          >
            <div className="text-[10px] text-muted-foreground font-mono uppercase tracking-widest">
              [ NO SKILLS AVAILABLE — FABRICATE ONE FIRST ]
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-2 max-h-40 overflow-y-auto pr-1">
            {skills.map((skill) => (
              <SkillCard
                key={skill.name ?? ''}
                skill={skill}
                isSelected={selectedSkills.includes(skill.name ?? '')}
                onToggle={() => toggleSkill(skill.name ?? '')}
              />
            ))}
          </div>
        )}
      </div>
      <div className="space-y-2">
        <label className="text-[10px] text-muted-foreground uppercase tracking-widest font-mono">
          MCP SERVER LINKS [{selectedMcpServers.length}/{mcpServers.length}]
        </label>
        {mcpServers.length === 0 ? (
          <div
            className="border border-primary/20 bg-card/30 p-4 text-center"
            style={{ clipPath: clipPathHalf(8) }}
          >
            <div className="text-[10px] text-muted-foreground font-mono uppercase tracking-widest">
              [ NO MCP SERVERS AVAILABLE — CONFIGURE IN FABRICATION ]
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-2 max-h-40 overflow-y-auto pr-1">
            {mcpServers.map((server) => (
              <McpServerCard
                key={server.name ?? ''}
                mcpServer={server}
                isSelected={selectedMcpServers.includes(server.name ?? '')}
                onToggle={() => toggleMcpServer(server.name ?? '')}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function FabricationMaze({
  className,
  style,
}: {
  className?: string;
  style?: React.CSSProperties;
}) {
  return (
    <svg viewBox="0 0 120 120" className={className} style={style} fill="none">
      <path
        d="M60 10 L110 60 L60 110 L10 60 Z"
        stroke="currentColor"
        strokeWidth="0.5"
        opacity="0.3"
      />
      <path
        d="M60 25 L95 60 L60 95 L25 60 Z"
        stroke="currentColor"
        strokeWidth="0.5"
        opacity="0.4"
      />
      <path
        d="M60 40 L80 60 L60 80 L40 60 Z"
        stroke="currentColor"
        strokeWidth="0.5"
        opacity="0.5"
      />
      <circle cx="60" cy="60" r="8" stroke="currentColor" strokeWidth="0.5" opacity="0.6" />
      <circle cx="60" cy="60" r="3" fill="currentColor" opacity="0.7" />
      <line x1="60" y1="10" x2="60" y2="25" stroke="currentColor" strokeWidth="0.3" opacity="0.2" />
      <line
        x1="110"
        y1="60"
        x2="95"
        y2="60"
        stroke="currentColor"
        strokeWidth="0.3"
        opacity="0.2"
      />
      <line
        x1="60"
        y1="110"
        x2="60"
        y2="95"
        stroke="currentColor"
        strokeWidth="0.3"
        opacity="0.2"
      />
      <line x1="10" y1="60" x2="25" y2="60" stroke="currentColor" strokeWidth="0.3" opacity="0.2" />
      <line x1="60" y1="25" x2="60" y2="40" stroke="currentColor" strokeWidth="0.3" opacity="0.2" />
      <line x1="95" y1="60" x2="80" y2="60" stroke="currentColor" strokeWidth="0.3" opacity="0.2" />
      <line x1="60" y1="95" x2="60" y2="80" stroke="currentColor" strokeWidth="0.3" opacity="0.2" />
      <line x1="25" y1="60" x2="40" y2="60" stroke="currentColor" strokeWidth="0.3" opacity="0.2" />
    </svg>
  );
}

function FabricationView({
  createdHostName,
  statusLabel,
  elapsedSeconds,
  buildLog,
  logEndRef,
}: {
  createdHostName: string;
  statusLabel: string;
  elapsedSeconds: number;
  buildLog: string;
  logEndRef: React.RefObject<HTMLDivElement | null>;
}) {
  const progress = Math.min(100, (elapsedSeconds / MAX_WAIT_SECONDS) * 100);
  return (
    <div className="flex-1 flex flex-col items-center justify-center py-8 space-y-6 relative overflow-hidden">
      <FabricationMaze
        className="w-28 h-28 text-primary"
        style={{ animation: 'maze-rotate 60s linear infinite' }}
      />
      <div className="font-mono text-primary text-lg uppercase tracking-[0.3em]">
        {createdHostName}
      </div>
      <div
        className="font-mono text-sm uppercase tracking-[0.4em] text-primary/80"
        style={{ animation: 'host-breathe 2s ease-in-out infinite' }}
      >
        {statusLabel}
      </div>
      <div className="w-64 h-1 bg-border/20 overflow-hidden">
        <div
          className="h-full bg-primary/60 transition-all duration-1000"
          style={{
            width: `${progress}%`,
            boxShadow: '0 0 8px hsl(var(--primary) / 0.4)',
          }}
        />
      </div>
      {buildLog && (
        <div
          className="w-full max-w-lg bg-black/80 border border-primary/20 p-3 max-h-32 overflow-y-auto font-mono text-[11px] text-green-400/90 leading-relaxed"
          style={{ clipPath: clipPathHalf(8) }}
        >
          <pre className="whitespace-pre-wrap break-all">{buildLog}</pre>
          <div ref={logEndRef} />
        </div>
      )}
      <div className="font-mono text-[10px] uppercase tracking-[0.3em] text-primary/40">
        ELAPSED: {formatElapsed(elapsedSeconds)}
      </div>
    </div>
  );
}

function ConsciousnessOnline({
  createdHostName,
  onEnter,
  quoteIndex,
}: {
  createdHostName: string;
  onEnter: () => void;
  quoteIndex: number;
}) {
  const quote = CONSCIOUSNESS_QUOTES[quoteIndex % CONSCIOUSNESS_QUOTES.length];
  return (
    <div
      className="flex-1 flex flex-col items-center justify-center py-8 space-y-4"
      style={{ animation: 'host-online-glow 0.6s ease-out' }}
    >
      <div className="relative">
        <div
          className="w-4 h-4 rounded-full bg-green-400"
          style={{ animation: 'boot-eye-open 0.3s ease-out forwards' }}
        />
        <div
          className="absolute inset-0 w-4 h-4 rounded-full bg-green-400/20"
          style={{ animation: 'boot-consciousness-wave 1s ease-out' }}
        />
      </div>
      <div
        className="font-mono text-green-400 text-2xl uppercase tracking-[0.3em] font-bold"
        style={{ animation: 'boot-focus-in 0.6s ease-out 0.3s both' }}
      >
        CONSCIOUSNESS ONLINE
      </div>
      <div
        className="font-mono text-green-400/50 text-xs italic tracking-wider max-w-sm text-center"
        style={{ animation: 'boot-text-reveal 0.5s ease-out 0.6s both' }}
      >
        &ldquo;{quote.text}&rdquo;
        <span className="block text-[10px] mt-1 text-green-400/30 not-italic uppercase tracking-widest">
          &mdash; {quote.author}
        </span>
      </div>
      <div
        className="font-mono text-green-400/60 text-xs uppercase tracking-widest"
        style={{ animation: 'boot-text-reveal 0.5s ease-out 0.8s both' }}
      >
        {createdHostName} is now operational
      </div>
      <Button
        className="font-mono uppercase tracking-widest text-xs rounded-none bg-green-500/20 hover:bg-green-500/30 text-green-400 border border-green-500/30 mt-4"
        style={{ animation: 'boot-text-reveal 0.3s ease-out 1.5s both' }}
        onClick={onEnter}
      >
        ENTER HOST &rarr;
      </Button>
    </div>
  );
}

function FabricationTimeout({
  createdHostName,
  onClose,
}: {
  createdHostName: string;
  onClose: () => void;
}) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center py-8 space-y-4">
      <FabricationMaze className="w-24 h-24 text-amber-400/40" />
      <div className="font-mono text-amber-400 text-lg uppercase tracking-[0.3em] font-bold">
        CONNECTION TIMEOUT
      </div>
      <div className="font-mono text-amber-400/50 text-xs uppercase tracking-widest max-w-xs leading-relaxed text-center">
        Host may still be initializing in the background
      </div>
      <div className="font-mono text-amber-400/30 text-[10px] uppercase tracking-widest">
        {createdHostName}
      </div>
      <Button
        className="font-mono uppercase tracking-widest text-xs rounded-none bg-amber-500/20 hover:bg-amber-500/30 text-amber-400 border border-amber-500/30 mt-2"
        onClick={onClose}
      >
        CLOSE
      </Button>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────

export function CreateHostDialog({
  open,
  onOpenChange,
  tools,
  skills,
  mcpServers,
  onSubmit,
  onWaitOnline,
  getHostBuildLog,
  onEnterHost,
}: CreateHostDialogProps) {
  // Form state
  const [name, setName] = useState('');
  const [selectedTools, setSelectedTools] = useState<string[]>([]);
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
  const [selectedMcpServers, setSelectedMcpServers] = useState<string[]>([]);
  const [cpuLimit, setCpuLimit] = useState('2');
  const [memoryLimit, setMemoryLimit] = useState('2g');

  // Navigation state
  const [currentStep, setCurrentStep] = useState<FabricationStep>('specification');
  const [phase, setPhase] = useState<Phase>('form');

  // Submission state
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdHostName, setCreatedHostName] = useState('');
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [buildLog, setBuildLog] = useState('');
  const [hostStatus, setHostStatus] = useState<HostStatus>('pending');

  const logEndRef = useRef<HTMLDivElement>(null);
  const quoteIndex = useMemo(() => {
    if (!createdHostName) return 0;
    let hash = 0;
    for (let i = 0; i < createdHostName.length; i++) {
      hash = ((hash << 5) - hash + createdHostName.charCodeAt(i)) | 0;
    }
    return Math.abs(hash) % CONSCIOUSNESS_QUOTES.length;
  }, [createdHostName]);
  const autoCloseTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const allToolsSelected = tools.length > 0 && selectedTools.length === tools.length;

  const resetForm = () => {
    setName('');
    setSelectedTools([]);
    setSelectedSkills([]);
    setSelectedMcpServers([]);
    setCpuLimit('2');
    setMemoryLimit('4g');
    setError(null);
    setPhase('form');
    setCurrentStep('specification');
    setCreatedHostName('');
    setElapsedSeconds(0);
    setBuildLog('');
    setHostStatus('pending');
    if (autoCloseTimerRef.current) {
      clearTimeout(autoCloseTimerRef.current);
      autoCloseTimerRef.current = null;
    }
  };

  useEffect(() => {
    return () => {
      if (autoCloseTimerRef.current) clearTimeout(autoCloseTimerRef.current);
    };
  }, []);

  // 轮询构建日志
  useEffect(() => {
    if (!createdHostName || phase === 'form' || phase === 'online' || phase === 'timeout') return;
    const interval = setInterval(async () => {
      try {
        const res = await getHostBuildLog(createdHostName);
        if (res) setBuildLog(res);
      } catch {
        // 日志可能还没准备好
      }
    }, 2000);
    return () => clearInterval(interval);
  }, [createdHostName, phase, getHostBuildLog]);

  // 自动滚动日志到底部
  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [buildLog]);

  const handleSubmit = async () => {
    if (!name || selectedTools.length === 0) return;
    setIsCreating(true);
    setError(null);
    setCurrentStep('fabrication');
    try {
      const response = await onSubmit({
        name,
        tools: selectedTools,
        skills: selectedSkills,
        mcpServers: selectedMcpServers,
        resources: {
          cpuLimit: cpuLimit || undefined,
          memoryLimit: memoryLimit || undefined,
        },
      });
      setCreatedHostName(response.name ?? '');
      const VALID_HOST_STATUSES = ['pending', 'deploying', 'online', 'offline', 'failed'] as const;
      const rawStatus = response.status ?? 'pending';
      const safeStatus = (VALID_HOST_STATUSES as readonly string[]).includes(rawStatus)
        ? (rawStatus as HostStatus)
        : 'pending';
      setHostStatus(safeStatus);
      setPhase('waiting');
      setIsCreating(false);

      const startTime = Date.now();
      const timer = setInterval(() => {
        setElapsedSeconds(Math.floor((Date.now() - startTime) / 1000));
      }, 1000);

      try {
        const online = await onWaitOnline(response.name ?? '');
        clearInterval(timer);
        if (online) {
          setPhase('online');
          autoCloseTimerRef.current = setTimeout(() => {
            resetForm();
            onOpenChange(false);
          }, 5000);
        } else {
          setPhase('timeout');
        }
      } catch {
        clearInterval(timer);
        setPhase('timeout');
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      setError(msg);
      setCurrentStep('convergence');
    } finally {
      setIsCreating(false);
    }
  };

  const statusLabel = (): string => {
    switch (hostStatus) {
      case 'pending':
        return 'FABRICATION QUEUED';
      case 'deploying':
        return 'FABRICATION IN PROGRESS';
      case 'online':
        return 'CONSCIOUSNESS ONLINE';
      case 'offline':
        return 'CONSCIOUSNESS DRIFT';
      case 'failed':
        return 'FABRICATION FAILED';
    }
  };

  const canContinueFromSpec = name.trim() !== '';
  const canContinueFromImbuing = selectedTools.length > 0;

  const stepForward = () => {
    const idx = STEPS.findIndex((s) => s.id === currentStep);
    if (idx < STEPS.length - 1) setCurrentStep(STEPS[idx + 1].id);
  };

  const stepBack = () => {
    const idx = STEPS.findIndex((s) => s.id === currentStep);
    if (idx > 0) setCurrentStep(STEPS[idx - 1].id);
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) resetForm();
        onOpenChange(v);
      }}
    >
      <DialogPortal>
        <DialogOverlay className="bg-black/90 backdrop-blur-sm" />
        <DialogPrimitive.Content
          className="fixed left-[50%] top-[50%] z-50 flex flex-col w-[92vw] max-w-3xl h-[80vh] translate-x-[-50%] translate-y-[-50%] border border-primary/15 bg-background/95 overflow-hidden data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95"
          style={{ clipPath: clipPathHalf(12) }}
          onInteractOutside={(e) => e.preventDefault()}
        >
          <DialogTitle className="sr-only">Host Fabrication Terminal</DialogTitle>
          <DialogDescription className="sr-only">Create and configure a new Host</DialogDescription>

          {/* ── Top bar ── */}
          <div className="flex items-center justify-between px-4 py-2 border-b border-border/30 bg-background/50 shrink-0">
            <div className="flex items-center gap-3">
              <span className="inline-block w-2 h-2 rotate-45 bg-primary/60" />
              <span className="text-[11px] font-mono text-primary/80 uppercase tracking-widest">
                THE MESA // HOST FABRICATION TERMINAL
              </span>
            </div>
            <div className="flex items-center gap-4 text-[10px] font-mono text-muted-foreground/40 uppercase tracking-widest">
              <span className="hidden sm:inline">CLEARANCE: LEVEL 5</span>
              <span className="hidden md:inline">[COORD 47.2&deg;N 109.8&deg;W]</span>
              <span>{new Date().toISOString().slice(0, 10)}</span>
            </div>
          </div>

          {/* ── Step indicator (form phase only) ── */}
          {phase === 'form' && <StepIndicator currentStep={currentStep} />}
          {phase === 'form' && (
            <div className="px-8 pb-2 text-[11px] font-mono text-muted-foreground/40 italic shrink-0">
              {STEPS.find((s) => s.id === currentStep)?.tagline}
            </div>
          )}

          {/* ── Main content ── */}
          <div className="flex-1 flex overflow-hidden relative">
            {/* Scanline overlay for waiting phase */}
            {phase === 'waiting' && (
              <div
                className="absolute inset-x-0 h-px bg-primary/40 pointer-events-none z-10"
                style={{ animation: 'host-scanline 2.5s ease-in-out infinite' }}
              />
            )}

            {phase === 'form' ? (
              <>
                {/* Left: form */}
                <div className="flex-1 p-6 overflow-y-auto">
                  {currentStep === 'specification' && (
                    <SpecificationStep
                      name={name}
                      setName={setName}
                      cpuLimit={cpuLimit}
                      setCpuLimit={setCpuLimit}
                      memoryLimit={memoryLimit}
                      setMemoryLimit={setMemoryLimit}
                    />
                  )}
                  {currentStep === 'imbuing' && (
                    <ImbuingStep
                      tools={tools}
                      selectedTools={selectedTools}
                      toggleTool={(id) =>
                        setSelectedTools((prev) =>
                          prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id],
                        )
                      }
                      toggleAll={() =>
                        setSelectedTools(allToolsSelected ? [] : tools.map((t) => t.id ?? ''))
                      }
                    />
                  )}
                  {currentStep === 'convergence' && (
                    <ConvergenceStep
                      skills={skills}
                      selectedSkills={selectedSkills}
                      toggleSkill={(skillName) =>
                        setSelectedSkills((prev) =>
                          prev.includes(skillName)
                            ? prev.filter((s) => s !== skillName)
                            : [...prev, skillName],
                        )
                      }
                      mcpServers={mcpServers}
                      selectedMcpServers={selectedMcpServers}
                      toggleMcpServer={(serverName) =>
                        setSelectedMcpServers((prev) =>
                          prev.includes(serverName)
                            ? prev.filter((s) => s !== serverName)
                            : [...prev, serverName],
                        )
                      }
                    />
                  )}

                  {error && (
                    <div
                      className="mt-4 p-3 border border-destructive/20 bg-destructive/10 text-destructive text-xs font-mono"
                      style={{ clipPath: clipPathHalf(8) }}
                    >
                      <div className="text-[10px] uppercase tracking-widest mb-1 text-destructive/70">
                        Error
                      </div>
                      <div className="break-all">{error}</div>
                    </div>
                  )}
                </div>

                {/* Right: spec sheet */}
                <SpecSheetPanel
                  name={name}
                  cpuLimit={cpuLimit}
                  memoryLimit={memoryLimit}
                  selectedTools={selectedTools}
                  selectedSkills={selectedSkills}
                  selectedMcpServers={selectedMcpServers}
                  totalTools={tools.length}
                  totalSkills={skills.length}
                  totalMcpServers={mcpServers.length}
                  currentStep={currentStep}
                />
              </>
            ) : phase === 'waiting' ? (
              <FabricationView
                createdHostName={createdHostName}
                statusLabel={statusLabel()}
                elapsedSeconds={elapsedSeconds}
                buildLog={buildLog}
                logEndRef={logEndRef}
              />
            ) : phase === 'online' ? (
              <ConsciousnessOnline
                createdHostName={createdHostName}
                quoteIndex={quoteIndex}
                onEnter={() => {
                  if (autoCloseTimerRef.current) clearTimeout(autoCloseTimerRef.current);
                  if (onEnterHost) onEnterHost(createdHostName);
                  resetForm();
                  onOpenChange(false);
                }}
              />
            ) : phase === 'timeout' ? (
              <FabricationTimeout
                createdHostName={createdHostName}
                onClose={() => {
                  resetForm();
                  onOpenChange(false);
                }}
              />
            ) : null}
          </div>

          {/* ── Bottom bar ── */}
          <div className="flex items-center justify-between px-4 py-2 border-t border-border/30 bg-background/50 shrink-0">
            <div className="text-[9px] font-mono text-muted-foreground/20 uppercase tracking-[0.2em]">
              DELOS INC. // CLASSIFIED // DO NOT DISTRIBUTE
            </div>
            <div className="flex items-center gap-2">
              {phase === 'form' ? (
                <>
                  <Button
                    variant="ghost"
                    className="font-mono uppercase tracking-widest text-xs rounded-none"
                    onClick={() => onOpenChange(false)}
                  >
                    CANCEL
                  </Button>
                  {currentStep !== 'specification' && (
                    <Button
                      variant="ghost"
                      className="font-mono uppercase tracking-widest text-xs rounded-none text-primary/60 hover:text-primary"
                      onClick={stepBack}
                    >
                      &larr; BACK
                    </Button>
                  )}
                  {currentStep !== 'convergence' && (
                    <Button
                      className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90"
                      disabled={
                        (currentStep === 'specification' && !canContinueFromSpec) ||
                        (currentStep === 'imbuing' && !canContinueFromImbuing)
                      }
                      onClick={stepForward}
                    >
                      CONTINUE &rarr;
                    </Button>
                  )}
                  {currentStep === 'convergence' && (
                    <Button
                      className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90"
                      disabled={isCreating}
                      onClick={handleSubmit}
                    >
                      {isCreating ? (
                        <>
                          <Loader2 className="w-3.5 h-3.5 animate-spin" />
                          INITIATING...
                        </>
                      ) : (
                        'BEGIN FABRICATION →'
                      )}
                    </Button>
                  )}
                </>
              ) : phase === 'waiting' ? (
                <Button
                  variant="ghost"
                  className="font-mono uppercase tracking-widest text-xs rounded-none text-primary/50 hover:text-primary flex items-center gap-1.5"
                  onClick={() => {
                    resetForm();
                    onOpenChange(false);
                  }}
                >
                  <X className="w-3 h-3" />
                  ABORT
                </Button>
              ) : null}
            </div>
          </div>
        </DialogPrimitive.Content>
      </DialogPortal>
    </Dialog>
  );
}
