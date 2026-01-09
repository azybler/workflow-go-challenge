import React from 'react';

import { Handle, type NodeProps, Position } from '@xyflow/react';

import { Badge, Box, Text } from '@radix-ui/themes';

import type { WorkflowNode as NodeType } from '../types';

const nodeConfigs = {
  start: {
    icon: '🚀',
    color: 'green' as const,
    variant: 'soft' as const,
    bg: 'var(--green-2)',
    borderColor: 'var(--green-6)',
    selectedBorderColor: 'var(--green-8)',
  },
  form: {
    icon: '📝',
    color: 'blue' as const,
    variant: 'soft' as const,
    bg: 'var(--blue-2)',
    borderColor: 'var(--blue-6)',
    selectedBorderColor: 'var(--blue-8)',
  },
  integration: {
    icon: '🌤️',
    color: 'orange' as const,
    variant: 'soft' as const,
    bg: 'var(--orange-2)',
    borderColor: 'var(--orange-6)',
    selectedBorderColor: 'var(--orange-8)',
  },
  condition: {
    icon: '🔍',
    color: 'purple' as const,
    variant: 'soft' as const,
    bg: 'var(--purple-2)',
    borderColor: 'var(--purple-6)',
    selectedBorderColor: 'var(--purple-8)',
  },
  email: {
    icon: '📧',
    color: 'red' as const,
    variant: 'soft' as const,
    bg: 'var(--red-2)',
    borderColor: 'var(--red-6)',
    selectedBorderColor: 'var(--red-8)',
  },
  end: {
    icon: '✅',
    color: 'gray' as const,
    variant: 'soft' as const,
    bg: 'var(--gray-2)',
    borderColor: 'var(--gray-6)',
    selectedBorderColor: 'var(--gray-8)',
  },
};

const getNodeConfig = (type: string) => {
  if (type in nodeConfigs) {
    return nodeConfigs[type as keyof typeof nodeConfigs];
  }
  return nodeConfigs.start;
};

export const WorkflowNode: React.FC<NodeProps> = ({ data, type, selected }) => {
  const config = getNodeConfig(type);
  const nodeData = data as NodeType['data'];

  return (
    <Box
      style={{
        maxWidth: '200px',
        minWidth: '50px',
        position: 'relative',
        backgroundColor: config.bg,
        border: selected
          ? `2px solid ${config.selectedBorderColor}`
          : `1px solid ${config.borderColor}`,
        borderRadius: 'var(--radius-3)',
        boxShadow: selected ? `0 0 0 3px var(--${config.color}-4)` : '0 1px 3px rgba(0, 0, 0, 0.1)',
        transition: 'all 0.2s ease',
      }}
      p="2"
    >
      {type !== 'start' && (
        <Handle
          type="target"
          position={Position.Left}
          style={{
            background: `var(--${config.color}-9)`,
            border: 'none',
            width: '8px',
            height: '8px',
          }}
        />
      )}

      <Box style={{ textAlign: 'center' }}>
        <Text size="6" style={{ display: 'block', marginBottom: '4px' }}>
          {config.icon}
        </Text>

        <Text size="3" weight="bold" style={{ display: 'block', marginBottom: '4px' }}>
          {nodeData.label}
        </Text>

        <Text size="2" color="gray" style={{ display: 'block' }}>
          {nodeData.description}
        </Text>

        <Badge color={config.color} variant={config.variant} size="1" style={{ marginTop: '6px' }}>
          {type}
        </Badge>
      </Box>

      {type !== 'end' && (
        <Handle
          type="source"
          position={Position.Right}
          id={type === 'condition' ? 'true' : undefined}
          style={{
            background: `var(--${config.color}-9)`,
            border: 'none',
            width: '8px',
            height: '8px',
          }}
        />
      )}

      {type === 'condition' && (
        <Handle
          type="source"
          position={Position.Bottom}
          id="false"
          style={{
            background: `var(--${config.color}-9)`,
            border: 'none',
            width: '8px',
            height: '8px',
          }}
        />
      )}
    </Box>
  );
};
