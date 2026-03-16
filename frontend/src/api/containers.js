import { get, post, del } from './client'

export const containersApi = {
  list: (serverId) => get(`/servers/${serverId}/containers`),
  create: (serverId, data) => post(`/servers/${serverId}/containers`, data),
  action: (serverId, name, action) =>
    post(`/servers/${serverId}/containers/${name}/action`, { action }),
  logs: (serverId, name, tail = 200) =>
    get(`/servers/${serverId}/containers/${name}/logs`, { tail }),
  exec: (serverId, command) => post(`/servers/${serverId}/exec`, { command }),

  // Volumes
  listVolumes: (serverId) => get(`/servers/${serverId}/volumes`),
  createVolume: (serverId, name) => post(`/servers/${serverId}/volumes`, { name }),
  deleteVolume: (serverId, name) => del(`/servers/${serverId}/volumes/${name}`)
}
