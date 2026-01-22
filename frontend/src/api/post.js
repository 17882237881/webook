const BASE_URL = ''

// 获取存储的 Token
function getToken() {
    return localStorage.getItem('token')
}

// API 请求封装
async function request(url, options = {}) {
    const token = getToken()
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    }

    // 如果有 Token，添加 Authorization 头
    if (token) {
        headers['Authorization'] = `Bearer ${token}`
    }

    const response = await fetch(BASE_URL + url, {
        ...options,
        headers
    })
    return response.json()
}

// 保存草稿
export function savePost(id, title, content) {
    return request('/posts', {
        method: 'POST',
        body: JSON.stringify({ id, title, content })
    })
}

// 发布帖子
export function publishPost(id, title, content) {
    return request('/posts/publish', {
        method: 'POST',
        body: JSON.stringify({ id, title, content })
    })
}

// 获取草稿详情（作者用）
export function getDraft(id) {
    return request(`/posts/draft/${id}`)
}

// 获取已发布帖子详情（读者用）
export function getPublishedPost(id) {
    return request(`/posts/${id}`)
}

// 获取作者的帖子列表
export function getMyPosts(page = 1, pageSize = 10) {
    return request(`/posts/author?page=${page}&pageSize=${pageSize}`)
}

// 获取已发布帖子列表（公开）
export function getPublishedPosts(page = 1, pageSize = 10) {
    return request(`/posts?page=${page}&pageSize=${pageSize}`)
}

// 删除帖子
export function deletePost(id) {
    return request(`/posts/${id}`, {
        method: 'DELETE'
    })
}
