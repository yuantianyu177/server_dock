import { get, post } from './client'

export const applicationsApi = {
  // Admin
  list: (status) => get('/applications', status ? { status } : undefined),
  get: (id) => get(`/applications/${id}`),
  action: (id, action, adminNotes) =>
    post(`/applications/${id}/action`, { action, admin_notes: adminNotes }),

  // Public
  publicServers: () => get('/applications/public/servers'),
  publicImages: (serverId) => get(`/applications/public/server/${serverId}/images`),
  apply: (data) => post('/applications/public/apply', data)
}
