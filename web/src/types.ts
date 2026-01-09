export interface WorkflowFormData {
  name: string;
  email: string;
  city: string;
  operator:
    | 'greater_than'
    | 'less_than'
    | 'equals'
    | 'greater_than_or_equal'
    | 'less_than_or_equal';
  threshold: number;
}

interface NodeData {
  label: string;
  description: string;
  type?: string;
  [key: string]: unknown;
}

export interface WorkflowNode {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: NodeData;
}

export interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
  type?: string;
  style?: {
    stroke: string;
    strokeWidth: number;
  };
  sourceHandle?: string;
  targetHandle?: string;
  animated?: boolean;
  labelStyle?: {
    fill: string;
    fontWeight: string;
  };
}

interface ExecutionStep {
  stepNumber: number;
  nodeType: string;
  status: 'success' | 'error';
  duration: number;
  output: {
    message: string;
    details?: Record<string, unknown>;
    emailContent?: {
      to: string;
      subject: string;
      body: string;
      timestamp?: string;
    };
    apiResponse?: {
      endpoint: string;
      method: string;
      statusCode: number;
      data: unknown;
    };
    conditionResult?: {
      expression: string;
      result: boolean;
      temperature: number;
      operator: string;
      threshold: number;
    };
    formData?: WorkflowFormData;
  };
  timestamp: string;
  error?: string;
}

export interface ExecutionResults {
  executionId: string;
  status: 'completed' | 'failed' | 'cancelled';
  startTime: string;
  endTime: string;
  totalDuration?: number;
  steps: ExecutionStep[];
  metadata?: {
    workflowVersion?: string;
    triggeredBy?: string;
    environment?: string;
  };
}
