import { useEffect, useState } from 'react'
import { apiRequest, login } from './api/client'
import './App.css'

function App() {
  const [token, setToken] = useState(() => localStorage.getItem('access_token'))
  const [username, setUsername] = useState('test')
  const [password, setPassword] = useState('')
  const [products, setProducts] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function loadProducts() {
    setLoading(true)
    setError('')
    try {
      const data = await apiRequest('/products')
      setProducts(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (token) loadProducts()
  }, [token])

  async function handleLogin(e) {
    e.preventDefault()
    setError('')
    try {
      const data = await login(username, password)
      setToken(data.access_token)
    } catch (err) {
      setError(err.message)
    }
  }

  function handleLogout() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    setToken(null)
    setProducts([])
    setPassword('')
    setError('')
  }

  if (!token) {
    return (
      <div className="login-page">
        <form className="login-card" onSubmit={handleLogin}>
          <div className="brand-logo">W</div>
          <h1>Warehouse</h1>
          <p>Đăng nhập hệ thống quản lý kho</p>

          <label>Tên đăng nhập</label>
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Nhập tên đăng nhập"
            required
          />

          <label>Mật khẩu</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Nhập mật khẩu"
            required
          />

          {error && <p className="error-message">{error}</p>}
          <button className="primary-button full-width" type="submit">
            Đăng nhập
          </button>
        </form>
      </div>
    )
  }

  return (
    <div className="app-layout">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-logo">W</div>
          <div>
            <h2>WareHouse</h2>
            <small>MANAGEMENT SYSTEM</small>
          </div>
        </div>
        <p className="menu-label">MENU CHÍNH</p>
        <button className="menu-item active">📦　Sản phẩm</button>
        <div className="sidebar-bottom">
          <strong>Warehouse System</strong>
          <p>Hệ thống quản lý kho hàng</p>
        </div>
      </aside>

      <main className="main-area">
        <header className="topbar">
          <span>Trang chủ / <strong>Sản phẩm</strong></span>
          <button className="logout-button" onClick={handleLogout}>
            Đăng xuất
          </button>
        </header>

        <section className="page-content">
          <div className="page-heading">
            <div>
              <h1>Danh sách sản phẩm</h1>
              <p>Dữ liệu được lấy từ API Go và PostgreSQL.</p>
            </div>
            <button className="primary-button" onClick={loadProducts}>
              ↻ Làm mới
            </button>
          </div>

          {error && <p className="error-message">{error}</p>}
          {loading && <p>Đang tải dữ liệu...</p>}

          {!loading && !error && (
            <div className="panel">
              <div className="panel-heading">
                <div>
                  <h3>Sản phẩm trong kho</h3>
                  <p>Tổng số: {products.length} sản phẩm</p>
                </div>
              </div>

              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>MÃ SẢN PHẨM</th>
                      <th>TÊN SẢN PHẨM</th>
                      <th>GIÁ</th>
                      <th>SỐ LƯỢNG</th>
                      <th>TỒN KHO</th>
                    </tr>
                  </thead>
                  <tbody>
                    {products.map((p) => (
                      <tr key={p.id}>
                        <td className="code">{p.code}</td>
                        <td><strong>{p.name}</strong></td>
                        <td>{Number(p.price).toLocaleString('vi-VN')} đ</td>
                        <td>{p.quantity}</td>
                        <td>
                          <span className={`status ${p.quantity <= p.min_stock ? 'low' : 'available'}`}>
                            {p.quantity <= p.min_stock ? 'Sắp hết hàng' : 'Còn hàng'}
                          </span>
                        </td>
                      </tr>
                    ))}
                    {products.length === 0 && (
                      <tr><td colSpan="5">Chưa có sản phẩm nào.</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          <footer>Warehouse Management System · Go + React</footer>
        </section>
      </main>
    </div>
  )
}

export default App
