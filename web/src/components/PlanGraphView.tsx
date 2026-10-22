import React, { useMemo, useState } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  MarkerType,
  BackgroundVariant,
} from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { LayoutGrid, ArrowDown, ArrowRight } from 'lucide-react';
import type { ResourceNode, DependencyEdge } from '../types';
import { CustomResourceNode } from './CustomResourceNode';

export type GraphLayout = 'TB' | 'LR' | 'clusters';

interface PlanGraphViewProps {
  nodes: ResourceNode[];
  edges: DependencyEdge[];
  selectedNode: ResourceNode | null;
  onSelectNode: (node: ResourceNode) => void;
  actionFilter: string;
  searchQuery: string;
}

const nodeTypes = {
  customResource: CustomResourceNode,
};

export const PlanGraphView: React.FC<PlanGraphViewProps> = ({
  nodes,
  edges,
  selectedNode,
  onSelectNode,
  actionFilter,
  searchQuery,
}) => {
  const [layoutMode, setLayoutMode] = useState<GraphLayout>('TB');

  // Filter nodes based on action and search
  const filteredNodes = useMemo(() => {
    return nodes.filter(n => {
      if (actionFilter !== 'all' && n.action !== actionFilter) {
        return false;
      }
      if (searchQuery.trim() !== '') {
        const query = searchQuery.toLowerCase();
        const matchesAddress = n.address.toLowerCase().includes(query);
        const matchesType = n.type.toLowerCase().includes(query);
        return matchesAddress || matchesType;
      }
      return true;
    });
  }, [nodes, actionFilter, searchQuery]);

  const filteredNodeAddresses = useMemo(() => {
    return new Set(filteredNodes.map(n => n.address));
  }, [filteredNodes]);

  // Set of downstream nodes impacted by currently selected node
  const activeDownstreamSet = useMemo(() => {
    if (!selectedNode || !selectedNode.downstream_ids) {
      return new Set<string>();
    }
    return new Set(selectedNode.downstream_ids);
  }, [selectedNode]);

  // Compute Layout coordinates based on mode
  const { flowNodes, flowEdges } = useMemo(() => {
    const NODE_WIDTH = 300;
    const NODE_HEIGHT = 140;
    const HORIZONTAL_GAP = 60;
    const VERTICAL_GAP = 180;

    const computedNodes: Node[] = [];

    if (layoutMode === 'clusters') {
      // Group by module
      const moduleMap: Record<string, ResourceNode[]> = {};
      filteredNodes.forEach(n => {
        const mod = n.module && n.module.trim() !== '' ? n.module : 'root_module';
        if (!moduleMap[mod]) moduleMap[mod] = [];
        moduleMap[mod].push(n);
      });

      const moduleKeys = Object.keys(moduleMap);
      const MODULE_COLS = Math.ceil(Math.sqrt(moduleKeys.length)) || 1;
      const CLUSTER_WIDTH = 700;
      const CLUSTER_HEIGHT = 500;

      moduleKeys.forEach((modKey, modIndex) => {
        const modCol = modIndex % MODULE_COLS;
        const modRow = Math.floor(modIndex / MODULE_COLS);
        const clusterOriginX = modCol * CLUSTER_WIDTH;
        const clusterOriginY = modRow * CLUSTER_HEIGHT;

        const modNodes = moduleMap[modKey];
        const INNER_COLS = 2;

        modNodes.forEach((node, nodeIdx) => {
          const innerCol = nodeIdx % INNER_COLS;
          const innerRow = Math.floor(nodeIdx / INNER_COLS);
          const x = clusterOriginX + innerCol * (NODE_WIDTH + 40);
          const y = clusterOriginY + innerRow * (NODE_HEIGHT + 50);

          const isSelected = selectedNode?.address === node.address;
          const isDownstream = activeDownstreamSet.has(node.address);
          const isHighlighted = selectedNode ? (isSelected || isDownstream) : true;

          computedNodes.push({
            id: node.address,
            type: 'customResource',
            position: { x, y },
            data: {
              node,
              isSelected,
              isDownstream,
              isHighlighted,
              onSelect: onSelectNode,
            },
          });
        });
      });
    } else {
      // Level-based layered DAG (TB or LR)
      const inDegree: Record<string, number> = {};
      const adj: Record<string, string[]> = {};

      filteredNodes.forEach(n => {
        inDegree[n.address] = 0;
        adj[n.address] = [];
      });

      edges.forEach(e => {
        if (filteredNodeAddresses.has(e.source) && filteredNodeAddresses.has(e.target)) {
          adj[e.source].push(e.target);
          inDegree[e.target] = (inDegree[e.target] || 0) + 1;
        }
      });

      // Assign levels (ranks) using BFS
      const levels: Record<string, number> = {};
      const queue: string[] = [];

      Object.keys(inDegree).forEach(addr => {
        if (inDegree[addr] === 0) {
          levels[addr] = 0;
          queue.push(addr);
        }
      });

      while (queue.length > 0) {
        const curr = queue.shift()!;
        const currLevel = levels[curr] || 0;

        adj[curr]?.forEach(neighbor => {
          levels[neighbor] = Math.max(levels[neighbor] || 0, currLevel + 1);
          inDegree[neighbor]--;
          if (inDegree[neighbor] === 0) {
            queue.push(neighbor);
          }
        });
      }

      filteredNodes.forEach(n => {
        if (levels[n.address] === undefined) {
          levels[n.address] = 0;
        }
      });

      // Group nodes by level
      const levelGroups: Record<number, ResourceNode[]> = {};
      filteredNodes.forEach(n => {
        const lvl = levels[n.address] || 0;
        if (!levelGroups[lvl]) {
          levelGroups[lvl] = [];
        }
        levelGroups[lvl].push(n);
      });

      Object.keys(levelGroups).forEach(lvlStr => {
        const lvl = parseInt(lvlStr, 10);
        const group = levelGroups[lvl];

        if (layoutMode === 'TB') {
          const totalWidth = group.length * NODE_WIDTH + (group.length - 1) * HORIZONTAL_GAP;
          const startX = -totalWidth / 2;

          group.forEach((node, index) => {
            const x = startX + index * (NODE_WIDTH + HORIZONTAL_GAP);
            const y = lvl * VERTICAL_GAP;

            const isSelected = selectedNode?.address === node.address;
            const isDownstream = activeDownstreamSet.has(node.address);
            const isHighlighted = selectedNode ? (isSelected || isDownstream) : true;

            computedNodes.push({
              id: node.address,
              type: 'customResource',
              position: { x, y },
              data: {
                node,
                isSelected,
                isDownstream,
                isHighlighted,
                onSelect: onSelectNode,
              },
            });
          });
        } else {
          // LR (Left to Right)
          const totalHeight = group.length * NODE_HEIGHT + (group.length - 1) * 60;
          const startY = -totalHeight / 2;

          group.forEach((node, index) => {
            const x = lvl * (NODE_WIDTH + 120);
            const y = startY + index * (NODE_HEIGHT + 60);

            const isSelected = selectedNode?.address === node.address;
            const isDownstream = activeDownstreamSet.has(node.address);
            const isHighlighted = selectedNode ? (isSelected || isDownstream) : true;

            computedNodes.push({
              id: node.address,
              type: 'customResource',
              position: { x, y },
              data: {
                node,
                isSelected,
                isDownstream,
                isHighlighted,
                onSelect: onSelectNode,
              },
            });
          });
        }
      });
    }

    // Compute edges with highlight states
    const computedEdges: Edge[] = [];
    edges.forEach(e => {
      if (filteredNodeAddresses.has(e.source) && filteredNodeAddresses.has(e.target)) {
        const isImpactedEdge = selectedNode && 
          (e.source === selectedNode.address || activeDownstreamSet.has(e.source)) &&
          activeDownstreamSet.has(e.target);

        computedEdges.push({
          id: e.id,
          source: e.source,
          target: e.target,
          animated: !!isImpactedEdge,
          style: {
            stroke: isImpactedEdge ? '#f59e0b' : '#475569',
            strokeWidth: isImpactedEdge ? 3 : 1.5,
            opacity: selectedNode ? (isImpactedEdge ? 1 : 0.25) : 0.6,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: isImpactedEdge ? '#f59e0b' : '#64748b',
            width: 16,
            height: 16,
          },
        });
      }
    });

    return { flowNodes: computedNodes, flowEdges: computedEdges };
  }, [filteredNodes, edges, filteredNodeAddresses, selectedNode, activeDownstreamSet, onSelectNode, layoutMode]);

  return (
    <div className="w-full h-full relative bg-slate-950">
      <ReactFlow
        nodes={flowNodes}
        edges={flowEdges}
        nodeTypes={nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.2}
        maxZoom={1.5}
        className="react-flow-dark"
      >
        <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="#334155" />
        <Controls className="!bg-slate-900 !border-slate-800 !text-slate-300 fill-slate-300 shadow-xl" />
        <MiniMap
          nodeColor={n => {
            const dataNode = n.data?.node as ResourceNode;
            if (dataNode?.action === 'replace') return '#f43f5e';
            if (dataNode?.action === 'delete') return '#ef4444';
            if (dataNode?.action === 'create') return '#10b981';
            if (dataNode?.action === 'update') return '#f59e0b';
            return '#64748b';
          }}
          className="!bg-slate-900 !border-slate-800 rounded-lg shadow-xl"
          maskColor="rgba(15, 23, 42, 0.7)"
        />
      </ReactFlow>

      {/* Floating Layout Selector Controls */}
      <div className="absolute top-4 right-4 z-10 flex items-center gap-1 p-1 rounded-lg bg-slate-900/90 border border-slate-800 shadow-xl backdrop-blur-md text-xs">
        <button
          onClick={() => setLayoutMode('TB')}
          className={`flex items-center gap-1 px-2 sm:px-2.5 py-1 rounded font-medium transition ${
            layoutMode === 'TB'
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-slate-400 hover:text-slate-200'
          }`}
          title="Vertical Top-to-Bottom DAG"
        >
          <ArrowDown className="w-3.5 h-3.5" />
          <span className="hidden sm:inline">Vertical</span>
          <span>(TB)</span>
        </button>

        <button
          onClick={() => setLayoutMode('LR')}
          className={`flex items-center gap-1 px-2 sm:px-2.5 py-1 rounded font-medium transition ${
            layoutMode === 'LR'
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-slate-400 hover:text-slate-200'
          }`}
          title="Horizontal Left-to-Right DAG"
        >
          <ArrowRight className="w-3.5 h-3.5" />
          <span className="hidden sm:inline">Horizontal</span>
          <span>(LR)</span>
        </button>

        <button
          onClick={() => setLayoutMode('clusters')}
          className={`flex items-center gap-1 px-2 sm:px-2.5 py-1 rounded font-medium transition ${
            layoutMode === 'clusters'
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-slate-400 hover:text-slate-200'
          }`}
          title="Group by Module Clusters"
        >
          <LayoutGrid className="w-3.5 h-3.5" />
          <span>Clusters</span>
        </button>
      </div>

      {/* Floating Cascade Active Notice */}
      {selectedNode && activeDownstreamSet.size > 0 && (
        <div className="absolute top-4 left-4 z-10 flex items-center gap-2 px-3 py-1.5 rounded-lg bg-amber-950/80 border border-amber-500/40 text-amber-200 text-xs shadow-lg backdrop-blur-md">
          <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping" />
          <span>Tracing blast radius cascade: <b>{activeDownstreamSet.size}</b> downstream dependent resources</span>
        </div>
      )}
    </div>
  );
};
