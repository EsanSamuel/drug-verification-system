// MedVerify API Client

const BASE_URL = '/api/v1';

async function request(endpoint, options = {}) {
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const url = `${BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (response.status === 204) {
    return { success: true };
  }

  const contentType = response.headers.get('content-type');
  if (contentType && contentType.includes('image/')) {
    return response.blob();
  }

  let data;
  try {
    data = await response.json();
  } catch (err) {
    throw new Error(`Invalid server response (${response.status})`);
  }

  if (!response.ok) {
    const errorMsg = data?.error || data?.message || `Request failed with status ${response.status}`;
    throw new Error(errorMsg);
  }

  return data;
}

export const api = {
  // Public Verification
  verify: {
    check: (serial) => request(`/verify/${encodeURIComponent(serial)}`),
  },

  // Authentication
  auth: {
    login: (credentials) =>
      request('/auth/login', {
        method: 'POST',
        body: JSON.stringify(credentials),
      }),
    register: (userData) =>
      request('/auth/register', {
        method: 'POST',
        body: JSON.stringify(userData),
      }),
    me: () => request('/auth/me'),
  },

  // Manufacturers
  manufacturers: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/manufacturers${qs ? `?${qs}` : ''}`);
    },
    create: (data) =>
      request('/manufacturers', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },

  // Drugs
  drugs: {
    list: (params = {}) => {
      const cleanParams = Object.fromEntries(
        Object.entries(params).filter(([_, v]) => v !== '' && v !== null && v !== undefined)
      );
      const qs = new URLSearchParams(cleanParams).toString();
      return request(`/drugs${qs ? `?${qs}` : ''}`);
    },
    get: (id) => request(`/drugs/${id}`),
    create: (data) =>
      request('/drugs', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    update: (id, data) =>
      request(`/drugs/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
    delete: (id) =>
      request(`/drugs/${id}`, {
        method: 'DELETE',
      }),
    generateUnits: (drugId, quantity) =>
      request(`/drugs/${drugId}/units`, {
        method: 'POST',
        body: JSON.stringify({ quantity: Number(quantity) }),
      }),
    listUnits: (drugId, params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/drugs/${drugId}/units${qs ? `?${qs}` : ''}`);
    },
    getUnitQRUrl: (unitId, size = 300) => `${BASE_URL}/drug-units/${unitId}/qr?size=${size}`,
  },

  // Audit Logs
  logs: {
    list: (params = {}) => {
      const cleanParams = Object.fromEntries(
        Object.entries(params).filter(([_, v]) => v !== '' && v !== null && v !== undefined)
      );
      const qs = new URLSearchParams(cleanParams).toString();
      return request(`/verification-logs${qs ? `?${qs}` : ''}`);
    },
  },
};
