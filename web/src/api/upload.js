import request from '../utils/request'

export function uploadFile(formData) {
  return request.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function getFileInfo(fileId) {
  return request.get(`/upload/${fileId}`)
}
