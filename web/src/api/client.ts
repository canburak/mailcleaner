import axios from 'axios';
import type {
  Account,
  AccountCreate,
  Rule,
  RuleCreate,
  ConnectionStatus,
  PreviewResult,
  Folder
} from './types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_BASE}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Accounts API
export const accountsApi = {
  list: () => api.get<Account[]>('/accounts').then(r => r.data),

  get: (id: number) => api.get<Account>(`/accounts/${id}`).then(r => r.data),

  create: (data: AccountCreate) =>
    api.post<Account>('/accounts', data).then(r => r.data),

  update: (id: number, data: Partial<AccountCreate>) =>
    api.put<Account>(`/accounts/${id}`, data).then(r => r.data),

  delete: (id: number) => api.delete(`/accounts/${id}`),

  test: (id: number) =>
    api.post<ConnectionStatus>(`/accounts/${id}/test`).then(r => r.data),

  testDirect: (data: AccountCreate) =>
    api.post<ConnectionStatus>('/accounts/test', data).then(r => r.data),

  getFolders: (id: number) =>
    api.get<Folder[]>(`/accounts/${id}/folders`).then(r => r.data),

  createFolder: (id: number, name: string) =>
    api.post(`/accounts/${id}/folders`, { name }).then(r => r.data),
};

// Rules API
export const rulesApi = {
  list: (accountId: number) =>
    api.get<Rule[]>(`/accounts/${accountId}/rules`).then(r => r.data),

  get: (id: number) =>
    api.get<Rule>(`/rules/${id}`).then(r => r.data),

  create: (accountId: number, data: RuleCreate) =>
    api.post<Rule>(`/accounts/${accountId}/rules`, data).then(r => r.data),

  update: (id: number, data: Partial<RuleCreate>) =>
    api.put<Rule>(`/rules/${id}`, data).then(r => r.data),

  delete: (id: number) => api.delete(`/rules/${id}`),
};

// Preview API
export const previewApi = {
  preview: (accountId: number, folder = 'INBOX', limit = 100) =>
    api.get<PreviewResult>(`/accounts/${accountId}/preview`, {
      params: { folder, limit }
    }).then(r => r.data),

  apply: (accountId: number, folder = 'INBOX', dryRun = false) =>
    api.post<PreviewResult>(`/accounts/${accountId}/apply`, null, {
      params: { folder, dry_run: dryRun }
    }).then(r => r.data),
};

// WebSocket for live preview
export function createPreviewWebSocket(): WebSocket {
  const wsUrl = API_BASE.replace('http', 'ws') + '/ws/preview';
  return new WebSocket(wsUrl);
}

export default api;
