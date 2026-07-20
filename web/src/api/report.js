import request from '../utils/request'

export function getReportList(params) {
  return request.get('/report/list', { params })
}

export function getReportDetail(id) {
  return request.get(`/report/${id}`)
}
