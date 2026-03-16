import { get, post, put, del } from './client'

export const serversApi = {
  list: () => get('/servers'),
  get: (id) => get(`/servers/${id}`),
  create: (data) => post('/servers', data),
  update: (id, data) => put(`/servers/${id}`, data),
  delete: (id) => del(`/servers/${id}`),
  testConnection: (id) => post(`/servers/${id}/test`),
  testConnectionDirect: (data) => post('/servers/test-direct', data)
}
