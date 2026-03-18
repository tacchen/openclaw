import axios from 'axios'
import router from '../router'

const api = axios.create({
  baseURL: '/api'
})

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      // 清除本地缓存的 token，刷新页面
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.reload()
    }
    return Promise.reject(error)
  }
)

// 标记文章已读
export const markArticleRead = (articleId) => 
  api.patch(`/articles/${articleId}/read`)

// 批量标记已读
export const markAllRead = (feedId, category) => 
  api.post('/articles/mark-all-read', { feed_id: feedId, category })

// 获取未读数量
export const getUnreadCount = () => 
  api.get('/articles/unread-count')

export default api

// AI Summary
export const generateSummary = (articleId) =>
  api.post(`/articles/${articleId}/summary`)

export const getSummary = (articleId) =>
  api.get(`/articles/${articleId}/summary`)
