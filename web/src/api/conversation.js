import request from '../utils/request'

export function getConversationList(params) {
  return request.get('/conversation/list', { params })
}

export function getConversationDetail(id) {
  return request.get(`/conversation/${id}`)
}

export function deleteConversation(id) {
  return request.delete(`/conversation/${id}`)
}
