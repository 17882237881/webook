const BASE_URL = 'http://localhost:8080'

// API 请求封装
async function request(url, options = {}) {
    const response = await fetch(BASE_URL + url, {
        ...options,
        credentials: 'include', // 携带 cookie
        headers: {
            'Content-Type': 'application/json',
            ...options.headers
        }
    })
    return response.json()
}

// 注册
export function signup(email, password, confirmPassword) {
    return request('/users', {
        method: 'POST',
        body: JSON.stringify({ email, password, confirmPassword })
    })
}

// 登录
export function login(email, password) {
    return request('/users/login', {
        method: 'POST',
        body: JSON.stringify({ email, password })
    })
}

// 获取用户信息
export function getProfile(id) {
    return request(`/users/${id}`)
}

// 修改密码
export function updatePassword(id, oldPassword, newPassword) {
    return request(`/users/${id}/password`, {
        method: 'PUT',
        body: JSON.stringify({ oldPassword, newPassword })
    })
}
