import { useEffect, useState } from 'react';

import type { WorkflowEdge, WorkflowNode } from '../types';

import { WORKFLOW_EDGES, WORKFLOW_NODES } from '../constants';

interface WorkflowResponse {
  id: string;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
}

export function useWorkflow(id: string) {
  const [nodes, setNodes] = useState<WorkflowNode[]>([]);
  const [edges, setEdges] = useState<WorkflowEdge[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    fetch(`/api/v1/workflows/${id}`)
      .then(res => {
        if (!res.ok) throw new Error(`Failed to load workflow (${res.status})`);
        return res.json() as Promise<WorkflowResponse>;
      })
      .then(({ nodes, edges }) => {
        setNodes(nodes);
        setEdges(edges);
        setError(null);
      })
      .catch((err: Error) => {
        console.warn('API not available, falling back to default workflow:', err.message);
        // Fall back to default workflow structure when API is not available
        setNodes(WORKFLOW_NODES);
        setEdges(WORKFLOW_EDGES);
        setError(null); // Clear error since we have fallback data
      })
      .finally(() => setLoading(false));
  }, [id]);

  return { nodes, edges, setNodes, setEdges, loading, error };
}
