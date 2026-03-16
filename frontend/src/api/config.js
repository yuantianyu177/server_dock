import { get, put, post } from './client'

export const configApi = {
  list: () => get('/config'),
  all: () => get('/config/all'),
  update: (key, value) => put(`/config/${key}`, { value }),
  testEmail: () => post('/config/test-email')
}
