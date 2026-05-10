import { useEffect, useRef } from 'react';
import { forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide } from 'd3-force';
import type { Link } from '@/api';

interface GraphNode {
  id: string;
  title: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
}

interface GraphLink {
  source: string | GraphNode;
  target: string | GraphNode;
  relationType: string;
}

interface KnowledgeGraphProps {
  links: Link[];
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
  const svgRef = useRef<SVGSVGElement>(null);
  const simulationRef = useRef<ReturnType<typeof forceSimulation<GraphNode>> | null>(null);

  useEffect(() => {
    if (!svgRef.current || graphLinksProp.length === 0) return;

    const nodeMap = new Map<string, GraphNode>();
    graphLinksProp.forEach((link) => {
      if (!nodeMap.has(link.sourceId)) {
        nodeMap.set(link.sourceId, {
          id: link.sourceId,
          title: link.sourceTitle || link.sourceId,
          x: width / 2 + (Math.random() - 0.5) * 100,
          y: height / 2 + (Math.random() - 0.5) * 100,
          vx: 0,
          vy: 0,
        });
      }
      if (!nodeMap.has(link.targetId)) {
        nodeMap.set(link.targetId, {
          id: link.targetId,
          title: link.targetTitle || link.targetId,
          x: width / 2 + (Math.random() - 0.5) * 100,
          y: height / 2 + (Math.random() - 0.5) * 100,
          vx: 0,
          vy: 0,
        });
      }
    });

    const nodes = Array.from(nodeMap.values());
    const graphLinks: GraphLink[] = graphLinksProp.map((link) => ({
      source: link.sourceId,
      target: link.targetId,
      relationType: link.relationType,
    }));

    if (simulationRef.current) {
      simulationRef.current.stop();
    }

    const simulation = forceSimulation<GraphNode>(nodes)
      .force(
        'link',
        forceLink<GraphNode, GraphLink>(graphLinks)
          .id((d) => d.id)
          .distance(80),
      )
      .force('charge', forceManyBody().strength(-120))
      .force('center', forceCenter(width / 2, height / 2))
      .force('collide', forceCollide(30));

    simulationRef.current = simulation;

    const svg = svgRef.current;
    const g = (svg.querySelector('g.graph-container') as SVGGElement) || svg;

    while (g.firstChild) g.removeChild(g.firstChild);

    const linkElements = graphLinks.map((link) => {
      const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
      line.setAttribute('stroke', linkColorMap[link.relationType] || '#60a5fa');
      line.setAttribute('stroke-opacity', '0.4');
      line.setAttribute('stroke-width', '1');
      g.appendChild(line);
      return line;
    });

    const nodeGroups = nodes.map((node) => {
      const group = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      group.style.cursor = 'pointer';

      const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
      const isCurrent = node.id === currentId;
      circle.setAttribute('r', isCurrent ? '6' : '4');
      circle.setAttribute(
        'fill',
        isCurrent ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground))',
      );
      circle.setAttribute('stroke', isCurrent ? 'hsl(var(--primary))' : 'none');
      circle.setAttribute('stroke-width', isCurrent ? '2' : '0');

      const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
      text.textContent = node.title.length > 12 ? node.title.substring(0, 12) + '...' : node.title;
      text.setAttribute('fill', 'hsl(var(--muted-foreground))');
      text.setAttribute('font-size', '8');
      text.setAttribute('font-family', 'monospace');
      text.setAttribute('text-anchor', 'middle');
      text.setAttribute('dy', '-10');

      group.appendChild(circle);
      group.appendChild(text);

      if (onNodeClick) {
        group.addEventListener('click', () => onNodeClick(node.id));
      }

      g.appendChild(group);
      return { group, node };
    });

    simulation.on('tick', () => {
      linkElements.forEach((line, i) => {
        const link = graphLinks[i];
        const source = link.source as GraphNode;
        const target = link.target as GraphNode;
        line.setAttribute('x1', String(source.x));
        line.setAttribute('y1', String(source.y));
        line.setAttribute('x2', String(target.x));
        line.setAttribute('y2', String(target.y));
      });

      nodeGroups.forEach(({ group, node }) => {
        group.setAttribute('transform', `translate(${node.x},${node.y})`);
      });
    });

    return () => {
      simulation.stop();
      simulation.on('tick', null);
    };
  }, [graphLinksProp, currentId, width, height, onNodeClick]);

  if (graphLinksProp.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground font-mono text-sm">
        No document links
      </div>
    );
  }

  return (
    <svg
      ref={svgRef}
      width={width}
      height={height}
      className="bg-background border border-border rounded"
    >
      <g className="graph-container" />
    </svg>
  );
}
