import { useEffect, useRef, useState } from 'react';
import { forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide } from 'd3-force';
import type { DocLink } from '@/api';

interface GraphNode {
  id: string;
  title: string;
  x: number;
  y: number;
}

interface GraphLinkRender {
  source: GraphNode;
  target: GraphNode;
  relationType: string;
}

interface KnowledgeGraphProps {
  links: DocLink[];
  currentId?: string;
  width?: number;
  height?: number;
  onNodeClick?: (id: string) => void;
}

const linkColorMap: Record<string, string> = {
  reference: '#60a5fa',
  depends_on: '#fbbf24',
  implements: '#34d399',
  relates_to: '#a78bfa',
};

export default function KnowledgeGraph({
  links: graphLinksProp,
  currentId,
  width = 600,
  height = 400,
  onNodeClick,
}: KnowledgeGraphProps) {
  const [nodes, setNodes] = useState<GraphNode[]>([]);
  const [graphLinks, setGraphLinks] = useState<GraphLinkRender[]>([]);
  const simulationRef = useRef<ReturnType<typeof forceSimulation<GraphNode>> | null>(null);

  useEffect(() => {
    if (graphLinksProp.length === 0) return;

    const validLinks = graphLinksProp.filter((l) => l.sourceId && l.targetId);
    if (validLinks.length === 0) return;

    const nodeMap = new Map<string, GraphNode>();
    validLinks.forEach((link) => {
      const srcId = link.sourceId!;
      const tgtId = link.targetId!;
      if (!nodeMap.has(srcId)) {
        nodeMap.set(srcId, {
          id: srcId,
          title: link.sourceTitle || srcId,
          x: width / 2 + (Math.random() - 0.5) * 100,
          y: height / 2 + (Math.random() - 0.5) * 100,
        });
      }
      if (!nodeMap.has(tgtId)) {
        nodeMap.set(tgtId, {
          id: tgtId,
          title: link.targetTitle || tgtId,
          x: width / 2 + (Math.random() - 0.5) * 100,
          y: height / 2 + (Math.random() - 0.5) * 100,
        });
      }
    });

    const simNodes = Array.from(nodeMap.values());
    interface SimLink {
      source: string | GraphNode;
      target: string | GraphNode;
      relationType: string;
    }
    const simLinks: SimLink[] = validLinks.map((link) => ({
      source: link.sourceId!,
      target: link.targetId!,
      relationType: link.relationType || 'relates_to',
    }));

    if (simulationRef.current) {
      simulationRef.current.stop();
    }

    const simulation = forceSimulation<GraphNode>(simNodes)
      .force(
        'link',
        forceLink<GraphNode, SimLink>(simLinks)
          .id((d) => d.id)
          .distance(80),
      )
      .force('charge', forceManyBody().strength(-120))
      .force('center', forceCenter(width / 2, height / 2))
      .force('collide', forceCollide(30));

    simulationRef.current = simulation;

    simulation.on('tick', () => {
      const renderedLinks: GraphLinkRender[] = simLinks.map((l) => ({
        source: l.source as GraphNode,
        target: l.target as GraphNode,
        relationType: l.relationType,
      }));
      setNodes([...simNodes]);
      setGraphLinks(renderedLinks);
    });

    return () => {
      simulation.stop();
      simulation.on('tick', null);
    };
  }, [graphLinksProp, currentId, width, height]);

  if (graphLinksProp.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground font-mono text-sm">
        No document links
      </div>
    );
  }

  return (
    <svg width={width} height={height} className="bg-background border border-border rounded">
      {graphLinks.map((link, i) => (
        <line
          key={i}
          x1={link.source.x}
          y1={link.source.y}
          x2={link.target.x}
          y2={link.target.y}
          stroke={linkColorMap[link.relationType] || '#60a5fa'}
          strokeOpacity={0.4}
          strokeWidth={1}
        />
      ))}
      {nodes.map((node) => {
        const isCurrent = node.id === currentId;
        const displayTitle = node.title.length > 12 ? node.title.slice(0, 12) + '...' : node.title;
        return (
          <g
            key={node.id}
            transform={`translate(${node.x},${node.y})`}
            style={{ cursor: 'pointer' }}
            onClick={() => onNodeClick?.(node.id)}
          >
            <circle
              r={isCurrent ? 6 : 4}
              fill={isCurrent ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground))'}
              stroke={isCurrent ? 'hsl(var(--primary))' : 'none'}
              strokeWidth={isCurrent ? 2 : 0}
            />
            <text
              fill="hsl(var(--muted-foreground))"
              fontSize={8}
              fontFamily="monospace"
              textAnchor="middle"
              dy={-10}
            >
              {displayTitle}
            </text>
          </g>
        );
      })}
    </svg>
  );
}
