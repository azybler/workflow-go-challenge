import React from 'react';

import { CheckIcon, ClockIcon, Cross2Icon } from '@radix-ui/react-icons';
import { Badge, Box, Card, Code, Flex, ScrollArea, Separator, Text } from '@radix-ui/themes';

import type { ExecutionResults, WorkflowFormData } from '../types';

// Generic execution result types to handle multiple formats
interface GenericStep {
  nodeId?: string;
  stepNumber?: string;
  nodeType?: string;
  type?: string;
  label?: string;
  status: string;
  output?: Record<string, unknown>;
  duration?: number;
  error?: string;
}

interface GenericExecutionResults {
  runId?: string;
  executionId?: string;
  executedAt?: string;
  status: string;
  steps: GenericStep[];
}

interface ExecutionResultsProps {
  results: ExecutionResults | GenericExecutionResults | null;
  formData: WorkflowFormData | null;
}

export const ExecutionResultsComponent: React.FC<ExecutionResultsProps> = ({
  results,
  formData,
}) => {
  if (!results || !formData) return null;

  const getOperatorSymbol = (operator: string) => {
    const symbols = {
      greater_than: '>',
      less_than: '<',
      equals: '=',
      greater_than_or_equal: '≥',
      less_than_or_equal: '≤',
    };
    return symbols[operator as keyof typeof symbols] || '>';
  };

  const getStatusIcon = (status: string) => {
    return status === 'completed' ? (
      <CheckIcon style={{ color: 'var(--green-9)' }} />
    ) : (
      <Cross2Icon style={{ color: 'var(--red-9)' }} />
    );
  };

  const getStatusColor = (status: string) => {
    return status === 'completed' ? 'green' : 'red';
  };

  const getOverallStatusColor = (status: string) => {
    if (status === 'completed') return 'green';
    if (status === 'cancelled') return 'orange';
    return 'red';
  };

  const getOverallStatusIcon = (status: string) => {
    if (status === 'completed') {
      return <CheckIcon style={{ color: 'var(--green-9)' }} />;
    }
    return <Cross2Icon style={{ color: 'var(--red-9)' }} />;
  };

  return (
    <Box p="4" style={{ height: '100%', overflow: 'hidden' }}>
      <ScrollArea style={{ height: '100%' }}>
        <Text size="5" weight="bold" mb="4" style={{ display: 'block' }}>
          📊 Execution Results
        </Text>

        {/* Execution Status */}
        <Card mb="4">
          <Box p="4">
            <Flex align="center" gap="2" mb="2">
              {getOverallStatusIcon(results.status)}
              <Badge color={getOverallStatusColor(results.status)} variant="soft" size="2">
                {results.status}
              </Badge>
            </Flex>
            <Text size="2" color="gray">
              {'executedAt' in results && results.executedAt
                ? `Executed: ${new Date(results.executedAt).toLocaleString()}`
                : 'runId' in results && results.runId
                  ? `Run ID: ${results.runId}`
                  : `Execution ID: ${'executionId' in results ? results.executionId : 'N/A'}`}
            </Text>
          </Box>
        </Card>

        {/* Configuration Summary */}
        <Card mb="4">
          <Box p="4">
            <Text size="4" weight="medium" mb="3" style={{ display: 'block' }}>
              ⚙️ Your Configuration
            </Text>

            <Flex direction="column" gap="2">
              <Flex justify="between">
                <Text size="2" color="gray">
                  User:
                </Text>
                <Text size="2" weight="medium">
                  {formData.name}
                </Text>
              </Flex>
              <Flex justify="between">
                <Text size="2" color="gray">
                  Email:
                </Text>
                <Text size="2" weight="medium">
                  {formData.email}
                </Text>
              </Flex>
              <Flex justify="between">
                <Text size="2" color="gray">
                  City:
                </Text>
                <Text size="2" weight="medium">
                  {formData.city}
                </Text>
              </Flex>
              <Flex justify="between">
                <Text size="2" color="gray">
                  Condition:
                </Text>
                <Code size="2">
                  temperature {getOperatorSymbol(formData.operator)} {formData.threshold}°C
                </Code>
              </Flex>
            </Flex>
          </Box>
        </Card>

        {/* Execution Steps */}
        <Card>
          <Box p="4">
            <Text size="4" weight="medium" mb="3" style={{ display: 'block' }}>
              🔄 Execution Steps
            </Text>

            <Flex direction="column" gap="3">
              {results.steps.map((step, index) => (
                <Box key={index}>
                  <Card variant="surface">
                    <Box p="3">
                      <Flex align="center" justify="between" mb="2">
                        <Flex align="center" gap="2">
                          {getStatusIcon(step.status)}
                          <Text size="3" weight="medium">
                            {('type' in step && step.type) ||
                              ('nodeId' in step && step.nodeId) ||
                              ('stepNumber' in step && step.stepNumber) ||
                              'Unknown Step'}
                          </Text>
                          <Badge color={getStatusColor(step.status)} variant="soft" size="1">
                            {step.status}
                          </Badge>
                        </Flex>
                        {step.duration && (
                          <Flex align="center" gap="1">
                            <ClockIcon
                              style={{ width: '12px', height: '12px', color: 'var(--gray-9)' }}
                            />
                            <Text size="1" color="gray">
                              {step.duration}ms
                            </Text>
                          </Flex>
                        )}
                      </Flex>

                      <Text size="2" style={{ display: 'block', marginBottom: '8px' }}>
                        {'output' in step &&
                        step.output &&
                        'message' in step.output &&
                        typeof step.output.message === 'string'
                          ? step.output.message
                          : 'Step completed'}
                      </Text>

                      {step.status === 'error' && 'error' in step && step.error && (
                        <Box mb="2">
                          <Text size="2" color="red" style={{ display: 'block' }}>
                            Error: {String(step.error)}
                          </Text>
                        </Box>
                      )}

                      {/* Generic output display */}
                      {((step.output && Object.keys(step.output).length > 0) ||
                        ('nodeId' in step && step.nodeId) ||
                        ('type' in step && step.type) ||
                        ('label' in step && step.label)) && (
                        <Box mt="2">
                          <Text size="2" color="gray" mb="1" style={{ display: 'block' }}>
                            Node Data:
                          </Text>
                          <Code
                            size="1"
                            style={{
                              display: 'block',
                              whiteSpace: 'pre-wrap',
                              backgroundColor: 'var(--gray-2)',
                              padding: '8px',
                              borderRadius: 'var(--radius-2)',
                            }}
                          >
                            {JSON.stringify(
                              {
                                ...('nodeId' in step && step.nodeId && { nodeId: step.nodeId }),
                                ...('type' in step && step.type && { type: step.type }),
                                ...('label' in step && step.label && { label: step.label }),
                                ...('nodeType' in step &&
                                  step.nodeType && { nodeType: step.nodeType }),
                                status: step.status,
                                ...('duration' in step &&
                                  step.duration && { duration: step.duration }),
                                ...(step.output && { output: step.output }),
                              },
                              null,
                              2
                            )}
                          </Code>
                        </Box>
                      )}
                    </Box>
                  </Card>

                  {index < results.steps.length - 1 && <Separator my="2" />}
                </Box>
              ))}
            </Flex>
          </Box>
        </Card>
      </ScrollArea>
    </Box>
  );
};
