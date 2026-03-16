import { get, post } from './client'

export const authApi = {
  login: (username, password) => post('/auth/login', { username, password }),
  me: () => get('/auth/me'),
  changePassword: (oldPassword, newPassword) =>
    post('/auth/change-password', { old_password: oldPassword, new_password: newPassword })
}
