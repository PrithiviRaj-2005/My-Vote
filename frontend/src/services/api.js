// PulseVote API Client Service
const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Helper for HTTP requests
async function request(endpoint, options = {}) {
  const token = localStorage.getItem('pulsevote_token');

  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  const isJson = response.headers.get('content-type')?.includes('application/json');
  const data = isJson ? await response.json() : null;

  if (!response.ok) {
    const errorMsg = data?.error || `HTTP error! Status: ${response.status}`;
    throw new Error(errorMsg);
  }

  return data;
}

export const api = {
  // Authentication
  auth: {
    signup: (data) => request('/api/auth/signup', { method: 'POST', body: JSON.stringify(data) }),
    login: (data) => request('/api/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  },

  // Polls
  polls: {
    create: (data) => request('/api/polls', { method: 'POST', body: JSON.stringify(data) }),
    getUserPolls: () => request('/api/polls'),
    getPublicPoll: (shareCode) => request(`/api/polls/${shareCode}`),
    getResults: (shareCode) => request(`/api/polls/${shareCode}/results`),
    vote: (shareCode, optionId) =>
      request(`/api/polls/${shareCode}/vote`, {
        method: 'POST',
        body: JSON.stringify({ optionId }),
      }),
    update: (id, data) => request(`/api/polls/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    close: (id) => request(`/api/polls/${id}/close`, { method: 'POST' }),
    delete: (id) => request(`/api/polls/${id}`, { method: 'DELETE' }),
  },

  // Health
  health: {
    check: () => request('/api/health'),
  },
};

export default api;
