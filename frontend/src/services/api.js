// My Vote API Client Service
const rawBaseUrl = import.meta.env.VITE_API_URL || '';
// Strip trailing slash if present to avoid '//api/...' double slash issues
const BASE_URL = rawBaseUrl.trim().replace(/\/+$/, '') || (import.meta.env.DEV ? 'http://localhost:8080' : '');

if (import.meta.env.PROD && (!import.meta.env.VITE_API_URL || import.meta.env.VITE_API_URL.includes('localhost'))) {
  console.warn(
    '[My Vote] Notice: VITE_API_URL is unset or using localhost in production build.\n' +
    'To connect your Vercel deployment to Render, set VITE_API_URL in your Vercel Project Environment Variables to your Render backend URL (e.g. https://myvote-backend.onrender.com).'
  );
}

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

  let response;
  try {
    response = await fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers,
    });
  } catch (err) {
    // Catch browser "Failed to fetch" (network/CORS/offline)
    const targetUrl = BASE_URL || window.location.origin;
    console.error(`[My Vote API] Network failure attempting to reach ${targetUrl}${endpoint}:`, err);
    throw new Error(
      `Unable to connect to backend server at ${targetUrl}. Please ensure the backend is running and CORS is configured.`
    );
  }

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
