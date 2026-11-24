const API_URL = 'http://localhost:8080/api'

let refreshPromise = null

function clearSession() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')

  window.dispatchEvent(new Event('auth:session-expired'))
}

export async function login(username, password) {
  const response = await fetch(`${API_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })

  const data = await response.json().catch(() => ({}))

  if (!response.ok) {
    throw new Error(data.error || 'Đăng nhập thất bại')
  }

  localStorage.setItem('access_token', data.access_token)

  if (data.refresh_token) {
    localStorage.setItem('refresh_token', data.refresh_token)
  } else {
    localStorage.removeItem('refresh_token')
  }

  return data
}

async function refreshAccessToken() {
  if (refreshPromise) {
    return refreshPromise
  }

  refreshPromise = (async () => {
    const refreshToken = localStorage.getItem('refresh_token')

    if (!refreshToken) {
      throw new Error('Không tìm thấy Refresh Token')
    }

    const response = await fetch(`${API_URL}/refresh-token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || 'Không thể làm mới phiên đăng nhập')
    }

    if (!data.access_token || !data.refresh_token) {
      throw new Error('Phản hồi Refresh Token không hợp lệ')
    }

    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)

    return data.access_token
  })()
    .catch((error) => {
      clearSession()
      throw new Error('Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.')
    })
    .finally(() => {
      refreshPromise = null
    })

  return refreshPromise
}

async function sendRequest(path, options, token) {
  return fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })
}

async function parseResponse(response) {
  if (response.status === 204) return null

  const data = await response.json().catch(() => ({}))

  if (!response.ok) {
    throw new Error(data.error || `Lỗi API (${response.status})`)
  }

  return data
}

export async function apiRequest(path, options = {}) {
  const token = localStorage.getItem('access_token')

  let response = await sendRequest(path, options, token)

  if (response.status === 401 && token) {
    const currentToken = localStorage.getItem('access_token')

    const newToken =
      currentToken && currentToken !== token
        ? currentToken
        : await refreshAccessToken()

    response = await sendRequest(path, options, newToken)
  }

  return parseResponse(response)
}

export async function logout(refreshToken) {
  const response = await fetch(`${API_URL}/logout`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  const data = await response.json().catch(() => ({}))

  if (!response.ok) {
    throw new Error(data.error || 'Đăng xuất thất bại')
  }

  return data
}