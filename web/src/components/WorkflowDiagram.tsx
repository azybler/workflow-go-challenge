import { type Dispatch, type SetStateAction } from 'react';

import {
  Background,
  BackgroundVariant,
  Controls,
  type NodeChange,
  ReactFlow,
  ReactFlowProvider,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { Card } from '@radix-ui/themes';

import type { WorkflowEdge, WorkflowNode as WorkflowNodeType } from '../types';

import { WorkflowNode } from './WorkflowNode';

const nodeTypes = {
  start: WorkflowNode,
  form: WorkflowNode,
  integration: WorkflowNode,
  condition: WorkflowNode,
  email: WorkflowNode,
  end: WorkflowNode,
};

interface WorkflowDiagramProps {
  nodes: WorkflowNodeType[];
  edges: WorkflowEdge[];
  onNodesChange: Dispatch<SetStateAction<WorkflowNodeType[]>>;
}

export const WorkflowDiagram = ({ nodes, edges, onNodesChange }: WorkflowDiagramProps) => {
  const handleNodeChange = (changes: NodeChange<WorkflowNodeType>[]) => {
    onNodesChange(nds => {
      return nds.map(node => {
        const change = changes.find(c => 'id' in c && c.id === node.id);
        if (change && change.type === 'position' && 'position' in change && change.position) {
          return { ...node, position: change.position };
        }
        return node;
      });
    });
  };

  return (
    <Card style={{ height: 'calc(100vh - 200px)' }}>
      <ReactFlowProvider>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={handleNodeChange}
          nodeTypes={nodeTypes}
          nodesDraggable={true}
          nodesConnectable={false}
          elementsSelectable={true}
          fitView
          fitViewOptions={{
            padding: 0.08,
            minZoom: 0.4,
            maxZoom: 1.5,
            includeHiddenNodes: false,
          }}
          defaultViewport={{
            x: 0,
            y: 0,
            zoom: 0.2,
          }}
          minZoom={0.2}
          maxZoom={2}
        >
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
          <Controls showInteractive={false} />
        </ReactFlow>
      </ReactFlowProvider>
    </Card>
  );
};
