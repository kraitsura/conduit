import { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { SubTask } from '../types';

export const useTaskBreakdown = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [subtasks, setSubtasks] = useState<SubTask[]>([]);
  const [isConnected, setIsConnected] = useState(false);

  // Test backend connection on hook initialization
  useEffect(() => {
    testBackendConnection();
  }, []);

  const testBackendConnection = async () => {
    try {
      const response = await invoke<string>('test_backend');
      console.log('Backend test response:', response);
      setIsConnected(true);
      setError(null);
    } catch (err) {
      console.error('Backend test failed:', err);
      setIsConnected(false);
      setError('Failed to connect to backend. Please check if the application is running properly.');
    }
  };

  const getSubtasks = async (taskDescription: string) => {
    if (!isConnected) {
      await testBackendConnection();
      if (!isConnected) {
        throw new Error('Backend connection not established');
      }
    }

    setLoading(true);
    setError(null);
    try {
      console.log('Sending task to backend:', taskDescription);
      const result = await invoke<SubTask[]>('get_subtasks', { 
        task: { description: taskDescription }
      });
      console.log('Received subtasks from backend:', result);
      setSubtasks(result);
      return result;
    } catch (err) {
      console.error('Error getting subtasks:', err);
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(`Failed to get subtasks: ${errorMessage}`);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    subtasks,
    getSubtasks,
    testBackendConnection,
    isConnected,
  };
};