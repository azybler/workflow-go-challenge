import { useDeferredValue, useState } from 'react';

import '@xyflow/react/dist/style.css';

import { Box, Button, Flex, Text } from '@radix-ui/themes';

import type { WorkflowFormData } from './types';

import { WORKFLOW_EDGES, WORKFLOW_NODES } from './constants';

import { useExecuteWorkflow } from './hooks/useExecuteWorkflow';
import { useWorkflow } from './hooks/useWorkflow';

import { ExecutionResultsComponent } from './components/ExecutionResults';
import { UserInputForm } from './components/UserInputForm';
import { WorkflowDiagram } from './components/WorkflowDiagram';

const WORKFLOW_ID = '550e8400-e29b-41d4-a716-446655440000';

function App() {
  const {
    nodes,
    edges,
    setNodes,
    setEdges,
    loading: graphLoading,
    error: graphError,
  } = useWorkflow(WORKFLOW_ID);
  const {
    execute,
    results: executionResults,
    loading: isExecuting,
    resetExecuteResult,
  } = useExecuteWorkflow(WORKFLOW_ID);

  // Defer the heavy graph updates
  const deferredNodes = useDeferredValue(nodes);
  const deferredEdges = useDeferredValue(edges);

  const [formData, setFormData] = useState<WorkflowFormData | null>(null);

  

  const handleExecute = async (data: WorkflowFormData) => {
    setFormData(data);
    await execute(data);
  };

  const onReset = () => {
    resetExecuteResult();
    setFormData(null);
    setNodes(WORKFLOW_NODES);
    setEdges(WORKFLOW_EDGES);
  };

  return (
    <Box style={{ height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <Box
        p="4"
        style={{ borderBottom: '1px solid var(--gray-6)', backgroundColor: 'var(--gray-2)' }}
      >
        <Flex justify="between" align="center">
          <Text size="6" weight="bold" style={{ display: 'block' }}>
            🌤️ Weather Alert Workflow Engine
          </Text>

          <Flex gap="2">
            <Button variant="soft" onClick={onReset}>
              Reset
            </Button>
          </Flex>
        </Flex>
      </Box>

      <Box style={{ flex: 1, display: 'flex', minHeight: 0 }}>
        {/* Left: Workflow Diagram */}
        <Box style={{ flex: 1, minHeight: 0 }} p="4">
          <Text size="4" weight="medium" mb="3">
            Workflow Diagram
          </Text>

          {graphLoading ? (
            <Text>Loading workflow…</Text>
          ) : graphError ? (
            <Text color="red">Error loading workflow: {graphError}</Text>
          ) : (
            <WorkflowDiagram nodes={deferredNodes} edges={deferredEdges} onNodesChange={setNodes} />
          )}
        </Box>

        <Box
          style={{
            borderLeft: '1px solid var(--gray-6)',
            backgroundColor: 'var(--gray-1)',
            width: '400px',
            height: 'calc(100vh - 80px)',
            overflow: 'hidden',
          }}
        >
          {!executionResults ? (
            <UserInputForm onExecute={handleExecute} isExecuting={isExecuting} />
          ) : (
            <ExecutionResultsComponent results={executionResults} formData={formData} />
          )}
        </Box>
      </Box>
    </Box>
  );
}

export default App;
