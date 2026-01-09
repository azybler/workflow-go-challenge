import type { WorkflowEdge, WorkflowNode } from './types';

export const WORKFLOW_NODES: WorkflowNode[] = [
  {
    id: 'start',
    type: 'start',
    position: { x: -160, y: 300 },
    data: {
      label: 'Start',
      description: 'Begin weather check workflow',
      metadata: {
        hasHandles: { source: true, target: false },
      },
    },
  },
  {
    id: 'form',
    type: 'form',
    position: { x: 152, y: 304 },
    data: {
      label: 'User Input',
      description: 'Process collected data - name, email, location',
      metadata: {
        hasHandles: { source: true, target: true },
        inputFields: ['name', 'email', 'city'],
        outputVariables: ['name', 'email', 'city'],
      },
    },
  },
  {
    id: 'weather-api',
    type: 'integration',
    position: { x: 460, y: 304 },
    data: {
      label: 'Weather API',
      description: 'Fetch current temperature for {{city}}',
      metadata: {
        hasHandles: { source: true, target: true },
        inputVariables: ['city'],
        apiEndpoint:
          'https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true',
        options: [
          {
            city: 'Sydney',
            lat: -33.8688,
            lon: 151.2093,
          },
          {
            city: 'Melbourne',
            lat: -37.8136,
            lon: 144.9631,
          },
          {
            city: 'Brisbane',
            lat: -27.4698,
            lon: 153.0251,
          },
          {
            city: 'Perth',
            lat: -31.9505,
            lon: 115.8605,
          },
          {
            city: 'Adelaide',
            lat: -34.9285,
            lon: 138.6007,
          },
        ],
        outputVariables: ['temperature'],
      },
    },
  },
  {
    id: 'condition',
    type: 'condition',
    position: { x: 794, y: 304 },
    data: {
      label: 'Check Condition',
      description: 'Evaluate temperature threshold',
      metadata: {
        hasHandles: { source: ['true', 'false'], target: true },
        conditionExpression: 'temperature {{operator}} {{threshold}}',
        outputVariables: ['conditionMet'],
      },
    },
  },
  {
    id: 'email',
    type: 'email',
    position: { x: 1096, y: 88 },
    data: {
      label: 'Send Alert',
      description: 'Email weather alert notification',
      metadata: {
        hasHandles: { source: true, target: true },
        inputVariables: ['name', 'city', 'temperature'],
        emailTemplate: {
          subject: 'Weather Alert',
          body: 'Weather alert for {{city}}! Temperature is {{temperature}}°C!',
        },
        outputVariables: ['emailSent'],
      },
    },
  },
  {
    id: 'end',
    type: 'end',
    position: { x: 1360, y: 302 },
    data: {
      label: 'Complete',
      description: 'Workflow execution finished',
      metadata: {
        hasHandles: { source: false, target: true },
      },
    },
  },
];

// Enhanced workflow edges with better styling
export const WORKFLOW_EDGES: WorkflowEdge[] = [
  {
    id: 'e1',
    source: 'start',
    target: 'form',
    type: 'smoothstep',
    animated: true,
    style: { stroke: '#10b981', strokeWidth: 3 },
    label: 'Initialize',
  },
  {
    id: 'e2',
    source: 'form',
    target: 'weather-api',
    type: 'smoothstep',
    animated: true,
    style: { stroke: '#3b82f6', strokeWidth: 3 },
    label: 'Submit Data',
  },
  {
    id: 'e3',
    source: 'weather-api',
    target: 'condition',
    type: 'smoothstep',
    animated: true,
    style: { stroke: '#f97316', strokeWidth: 3 },
    label: 'Temperature Data',
  },
  {
    id: 'e4',
    source: 'condition',
    target: 'email',
    type: 'smoothstep',
    sourceHandle: 'true',
    animated: true,
    style: { stroke: '#10b981', strokeWidth: 3 },
    label: '✓ Condition Met',
    labelStyle: { fill: '#10b981', fontWeight: 'bold' },
  },
  {
    id: 'e5',
    source: 'condition',
    target: 'end',
    type: 'smoothstep',
    sourceHandle: 'false',
    animated: true,
    style: { stroke: '#6b7280', strokeWidth: 3 },
    label: '✗ No Alert Needed',
    labelStyle: { fill: '#6b7280', fontWeight: 'bold' },
  },
  {
    id: 'e6',
    source: 'email',
    target: 'end',
    type: 'smoothstep',
    animated: true,
    style: { stroke: '#ef4444', strokeWidth: 2 },
    label: 'Alert Sent',
    labelStyle: { fill: '#ef4444', fontWeight: 'bold' },
  },
];
