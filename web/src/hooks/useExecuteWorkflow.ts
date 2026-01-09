import { useState } from 'react';

import type { ExecutionResults, WorkflowFormData } from '../types';

interface ExecuteError {
  message: string;
}

export function useExecuteWorkflow(id: string) {
  const [results, setResults] = useState<ExecutionResults | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function execute(formData: WorkflowFormData) {
    setLoading(true);
    setError(null);
    setResults(null);

    try {
      const res = await fetch(`/api/v1/workflows/${id}/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          formData,
          condition: { operator: formData.operator, threshold: formData.threshold },
        }),
      });
      if (!res.ok) {
        const errBody = (await res.json()) as ExecuteError;
        throw new Error(errBody.message || `Execute failed (${res.status})`);
      }
      const data = (await res.json()) as ExecutionResults;
      setResults(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unknown error occurred');
      }
    } finally {
      setLoading(false);
    }
  }

  return { execute, results, loading, error, resetExecuteResult: () => setResults(null) };
}
