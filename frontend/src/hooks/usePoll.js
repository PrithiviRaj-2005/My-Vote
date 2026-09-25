import { useState, useEffect, useCallback, useRef } from 'react';
import api from '../services/api';
import { createPollWebSocket } from '../services/websocket';

export function usePoll(shareCode) {
  const [poll, setPoll] = useState(null);
  const [results, setResults] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef(null);

  // Recalculates percentages given an array of options
  const recalculateOptions = (options) => {
    const total = options.reduce((sum, opt) => sum + (opt.votes || 0), 0);
    return options.map((opt) => ({
      ...opt,
      percentage: total > 0 ? ((opt.votes || 0) / total) * 100 : 0,
    }));
  };

  // Fetch initial poll and results
  const fetchPollData = useCallback(async () => {
    if (!shareCode) return;
    try {
      setLoading(true);
      setError(null);
      const data = await api.polls.getPublicPoll(shareCode);
      setPoll(data.poll);
      if (data.results) {
        setResults(data.results);
      }
    } catch (err) {
      setError(err.message || 'Failed to load poll');
    } finally {
      setLoading(false);
    }
  }, [shareCode]);

  useEffect(() => {
    fetchPollData();
  }, [fetchPollData]);

  // Connect Gorilla WebSocket
  useEffect(() => {
    if (!shareCode) return;

    const ws = createPollWebSocket(
      shareCode,
      (msg) => {
        // 1. If full results are broadcasted
        if (msg.results) {
          setResults(msg.results);
          if (msg.results.isActive !== undefined) {
            setPoll((prev) => (prev ? { ...prev, isActive: msg.results.isActive } : prev));
          }
          return;
        }

        // 2. If status change event
        if (msg.event === 'poll_closed') {
          setPoll((prev) => (prev ? { ...prev, isActive: false } : prev));
          setResults((prev) => (prev ? { ...prev, isActive: false } : prev));
          return;
        }

        // 3. If raw vote increment: { optionId: "...", increment: 1 }
        if (msg.optionId && msg.increment) {
          setResults((prev) => {
            if (!prev || !prev.options) return prev;

            const updatedOptions = prev.options.map((opt) => {
              if (opt.id === msg.optionId) {
                return { ...opt, votes: (opt.votes || 0) + msg.increment };
              }
              return opt;
            });

            const newTotal = (prev.totalVotes || 0) + msg.increment;
            const recalculated = recalculateOptions(updatedOptions);

            return {
              ...prev,
              totalVotes: newTotal,
              options: recalculated,
            };
          });
        }
      },
      () => setIsConnected(false),
      () => setIsConnected(false)
    );

    wsRef.current = ws;
    setIsConnected(true);

    return () => {
      if (wsRef.current) {
        wsRef.current.disconnect();
      }
    };
  }, [shareCode]);

  // Submit vote function
  const submitVote = async (optionId) => {
    try {
      const response = await api.polls.vote(shareCode, optionId);
      if (response.results) {
        setResults(response.results);
      }
      return { success: true, message: response.message };
    } catch (err) {
      return { success: false, error: err.message };
    }
  };

  return {
    poll,
    results,
    loading,
    error,
    isConnected,
    submitVote,
    refetch: fetchPollData,
  };
}
