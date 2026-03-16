import { get, post, put, del } from './client'

export const imagesApi = {
  // DB image records
  list: (serverId) => get('/images', serverId ? { server_id: serverId } : undefined),
  get: (id) => get(`/images/${id}`),
  create: (data) => post('/images', data),
  update: (id, data) => put(`/images/${id}`, data),
  delete: (id) => del(`/images/${id}`),

  // Remote images on server
  listRemote: (serverId) => get(`/servers/${serverId}/images`),
  pull: (serverId, image, tag) =>
    post(`/servers/${serverId}/images/pull`, { image: tag ? `${image}:${tag}` : image }),
  deleteRemote: (serverId, imageId) => del(`/servers/${serverId}/images/${imageId}`)
}
