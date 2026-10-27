
import { useEffect, useState } from 'react'
import { apiRequest, login } from './api/client'
import './App.css'

const menuItems = [
  { id: 'dashboard', icon: '▦', label: 'Tổng quan' },
  { id: 'products', icon: '📦', label: 'Sản phẩm' },
  { id: 'categories', icon: '▤', label: 'Danh mục' },
  { id: 'users', icon: '♙', label: 'Người dùng' },
  { id: 'stock', icon: '⇅', label: 'Nhập / Xuất kho' },
]

const emptyProduct = {
  code: '',
  name: '',
  price: '0',
  quantity: '0',
  min_stock: '0',
  category_id: '',
}

const emptyUser = {
  username: '',
  password: '',
  full_name: '',
  role: 'staff',
}

const money = (value) =>
  Number(value || 0).toLocaleString('vi-VN') + ' đ'

function App() {
  const [token, setToken] = useState(() =>
    localStorage.getItem('access_token'),
  )
  const [username, setUsername] = useState('test')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState('')
  const [activePage, setActivePage] = useState('dashboard')

  const [products, setProducts] = useState([])
  const [categories, setCategories] = useState([])
  const [users, setUsers] = useState([])
  // Tim kiem, loc va phan trang san pham
  const [productSearch, setProductSearch] = useState('')
  const [productCategoryFilter, setProductCategoryFilter] = useState('all')
  const [productStockFilter, setProductStockFilter] = useState('all')
  const [productPage, setProductPage] = useState(1)
  const productsPerPage = 8

  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [deletingId, setDeletingId] = useState(null)

  const [showProductForm, setShowProductForm] = useState(false)
  const [newProduct, setNewProduct] = useState({ ...emptyProduct })
  const [editingProductId, setEditingProductId] = useState(null)
  const [editProduct, setEditProduct] = useState({ ...emptyProduct })

  const [categoryName, setCategoryName] = useState('')
  const [categoryDescription, setCategoryDescription] = useState('')
  const [editingCategory, setEditingCategory] = useState(null)
  const [categoryLoading, setCategoryLoading] = useState(false)

  // Quản lý người dùng
  const [userForm, setUserForm] = useState({ ...emptyUser })
  const [editingUserId, setEditingUserId] = useState(null)
  const [userSaving, setUserSaving] = useState(false)
  const [userDeletingId, setUserDeletingId] = useState(null)
const [stockForm, setStockForm] = useState({
    product_id: '',
    quantity: '1',
    note: '',
  })
  const [stockType, setStockType] = useState('import')
  const [stockTransactions, setStockTransactions] = useState([])
  const [stockLoading, setStockLoading] = useState(false)
  const [stockSaving, setStockSaving] = useState(false)

  const totalInventoryValue = products.reduce(
    (total, product) =>
      total + Number(product.quantity || 0) * Number(product.price || 0),
    0,
  )

  const categoryStats = categories.map((category) => {
    const categoryProducts = products.filter(
      (product) => String(product.category_id) === String(category.id),
    )
    return {
      ...category,
      productCount: categoryProducts.length,
      quantity: categoryProducts.reduce(
        (total, product) => total + Number(product.quantity || 0),
        0,
      ),
    }
  })

  const totalImported = stockTransactions.reduce(
    (total, item) =>
      item.transaction_type === 'import'
        ? total + Number(item.quantity || 0)
        : total,
    0,
  )

  const totalExported = stockTransactions.reduce(
    (total, item) =>
      item.transaction_type === 'export'
        ? total + Number(item.quantity || 0)
        : total,
    0,
  )

  const maxStockMovement = Math.max(totalImported, totalExported, 1)
  const currentMenu = menuItems.find((item) => item.id === activePage)
  const isAdmin = role === 'admin'
  const filteredProducts = products.filter((product) => {
    const search = productSearch.trim().toLowerCase()
    const matchesSearch =
      String(product.code || '').toLowerCase().includes(search) ||
      String(product.name || '').toLowerCase().includes(search)

    const matchesCategory =
      productCategoryFilter === 'all' ||
      String(product.category_id ?? '') === productCategoryFilter

    const quantity = Number(product.quantity || 0)
    const minStock = Number(product.min_stock || 0)

    const matchesStock =
      productStockFilter === 'all' ||
      (productStockFilter === 'out' && quantity <= 0) ||
      (productStockFilter === 'low' && quantity > 0 && quantity <= minStock) ||
      (productStockFilter === 'safe' && quantity > minStock)

    return matchesSearch && matchesCategory && matchesStock
  })

  const totalProductPages = Math.max(
    1,
    Math.ceil(filteredProducts.length / productsPerPage),
  )
  const currentProductPage = Math.min(productPage, totalProductPages)
  const paginatedProducts = filteredProducts.slice(
    (currentProductPage - 1) * productsPerPage,
    currentProductPage * productsPerPage,
  )

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

  async function loadCategories() {
    setCategoryLoading(true)

    try {
      const data = await apiRequest('/categories')
      setCategories(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setCategoryLoading(false)
    }
  }

  async function loadStockTransactions() {
    setStockLoading(true)
    try {
      const data = await apiRequest('/stock/transactions')
      setStockTransactions(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setStockLoading(false)
    }
  }
  async function loadUsers() {
    if (!isAdmin) {
      setUsers([])
      return
    }

    try {
      const data = await apiRequest('/users')
      setUsers(Array.isArray(data) ? data : data?.users || [])
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    if (!token) return

    let cancelled = false

    apiRequest('/profile')
      .then((data) => {
        if (!cancelled) {
          setRole(data.role || '')
          setUsername(data.username || 'test')
        }
      })
      .catch((err) => {
        if (!cancelled) setError(err.message)
      })

    loadProducts()
    loadCategories()

    return () => {
      cancelled = true
    }
  }, [token])

  useEffect(() => {
    if (!token) return

    if (activePage === 'dashboard') loadStockTransactions()
    if (activePage === 'products') loadProducts()
    if (activePage === 'categories') loadCategories()
    if (activePage === 'users' && isAdmin) loadUsers()
    if (activePage === 'stock') {
      loadProducts()
      loadStockTransactions()
    }
  }, [activePage, token, isAdmin])

  async function handleStockSubmit(e) {
    e.preventDefault()
    setError('')

    const productId = Number(stockForm.product_id)
    const quantity = Number(stockForm.quantity)

    if (!productId || quantity <= 0 || !Number.isInteger(quantity)) {
      setError('Vui lòng chọn sản phẩm và nhập số lượng nguyên lớn hơn 0.')
      return
    }

    setStockSaving(true)
    try {
      await apiRequest(`/stock/${stockType}`, {
        method: 'POST',
        body: JSON.stringify({
          product_id: productId,
          quantity,
          note: stockForm.note,
        }),
      })

      setStockForm({ product_id: '', quantity: '1', note: '' })
      await loadProducts()
      await loadStockTransactions()
    } catch (err) {
      setError(err.message)
    } finally {
      setStockSaving(false)
    }
  }
  async function handleLogin(e) {
    e.preventDefault()
    setError('')

    try {
      const data = await login(username, password)
      setToken(data.access_token)
      setActivePage('dashboard')
      setPassword('')
    } catch (err) {
      setError(err.message)
    }
  }

  function handleLogout() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')

    setToken(null)
    setRole('')
    setProducts([])
    setCategories([])
    setUsers([])
    setPassword('')
    setError('')
    setActivePage('dashboard')
    setUserForm({ ...emptyUser })
    setEditingUserId(null)
  }

  function makeProductPayload(product) {
    return {
      code: product.code.trim(),
      name: product.name.trim(),
      category_id: product.category_id
        ? Number(product.category_id)
        : null,
      price: Number(product.price),
      quantity: Number(product.quantity),
      min_stock: Number(product.min_stock),
    }
  }

  async function handleCreateProduct(e) {
    e.preventDefault()
    setError('')
    setSaving(true)

    try {
      await apiRequest('/products', {
        method: 'POST',
        body: JSON.stringify(makeProductPayload(newProduct)),
      })

      setNewProduct({ ...emptyProduct })
      setShowProductForm(false)
      await loadProducts()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  function startEditProduct(product) {
    setEditingProductId(product.id)
    setEditProduct({
      code: product.code || '',
      name: product.name || '',
      price: String(product.price ?? 0),
      quantity: String(product.quantity ?? 0),
      min_stock: String(product.min_stock ?? 0),
      category_id: product.category_id
        ? String(product.category_id)
        : '',
    })
    setShowProductForm(false)
    setError('')
  }

  async function handleUpdateProduct(e) {
    e.preventDefault()
    if (editingProductId === null) return

    setSaving(true)
    setError('')

    try {
      await apiRequest(`/products/${editingProductId}`, {
        method: 'PUT',
        body: JSON.stringify(makeProductPayload(editProduct)),
      })

      setEditingProductId(null)
      setEditProduct({ ...emptyProduct })
      await loadProducts()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteProduct(product) {
    if (
      !window.confirm(
        `Bạn có chắc muốn xóa sản phẩm "${product.name}" không?`,
      )
    ) {
      return
    }

    setDeletingId(product.id)
    setError('')

    try {
      await apiRequest(`/products/${product.id}`, {
        method: 'DELETE',
      })

      if (editingProductId === product.id) {
        setEditingProductId(null)
      }

      await loadProducts()
    } catch (err) {
      setError(err.message)
    } finally {
      setDeletingId(null)
    }
  }

  async function handleSaveCategory(e) {
    e.preventDefault()
    setSaving(true)
    setError('')

    const payload = {
      name: categoryName.trim(),
      description: categoryDescription.trim() || null,
    }

    try {
      if (editingCategory) {
        await apiRequest(`/categories/${editingCategory.id}`, {
          method: 'PUT',
          body: JSON.stringify(payload),
        })
      } else {
        await apiRequest('/categories/create', {
          method: 'POST',
          body: JSON.stringify(payload),
        })
      }

      setCategoryName('')
      setCategoryDescription('')
      setEditingCategory(null)
      await loadCategories()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  function startEditCategory(category) {
    setEditingCategory(category)
    setCategoryName(category.name || '')
    setCategoryDescription(category.description || '')
    setError('')
  }

  async function handleDeleteCategory(category) {
    if (
      !window.confirm(
        `Bạn có chắc muốn xóa danh mục "${category.name}" không?`,
      )
    ) {
      return
    }

    setError('')

    try {
      await apiRequest(`/categories/${category.id}`, {
        method: 'DELETE',
      })

      if (editingCategory?.id === category.id) {
        setEditingCategory(null)
        setCategoryName('')
        setCategoryDescription('')
      }

      await loadCategories()
    } catch (err) {
      setError(err.message)
    }
  }

  // Thêm người dùng
  async function handleCreateUser(e) {
    e.preventDefault()
    setUserSaving(true)
    setError('')

    try {
      const payload = {
        username: userForm.username.trim(),
        password: userForm.password,
        full_name: userForm.full_name.trim(),
        role: userForm.role,
      }

      await apiRequest('/users', {
        method: 'POST',
        body: JSON.stringify(payload),
      })

      setUserForm({ ...emptyUser })
      await loadUsers()
    } catch (err) {
      setError(err.message)
    } finally {
      setUserSaving(false)
    }
  }

  // Chọn người dùng để sửa
  function startEditUser(user) {
    setEditingUserId(user.id)
    setUserForm({
      username: user.username || '',
      password: '',
      full_name: user.full_name || '',
      role: user.role || 'staff',
    })
    setError('')
  }

  // Cập nhật người dùng
  async function handleUpdateUser(e) {
    e.preventDefault()
    if (editingUserId === null) return

    setUserSaving(true)
    setError('')

    try {
      // API cập nhật hiện tại nhận username, full_name và role.
      const payload = {
        username: userForm.username.trim(),
        full_name: userForm.full_name.trim(),
        role: userForm.role,
      }

      await apiRequest(`/users/${editingUserId}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      })

      setEditingUserId(null)
      setUserForm({ ...emptyUser })
      await loadUsers()
    } catch (err) {
      setError(err.message)
    } finally {
      setUserSaving(false)
    }
  }

  // Hủy thao tác thêm hoặc sửa
  function cancelUserForm() {
    setEditingUserId(null)
    setUserForm({ ...emptyUser })
    setError('')
  }

  // Xóa người dùng
  async function handleDeleteUser(user) {
    if (
      !window.confirm(
        `Bạn có chắc muốn xóa tài khoản "${user.username}" không?`,
      )
    ) {
      return
    }

    setUserDeletingId(user.id)
    setError('')

    try {
      await apiRequest(`/users/${user.id}`, {
        method: 'DELETE',
      })

      if (editingUserId === user.id) {
        cancelUserForm()
      }

      await loadUsers()
    } catch (err) {
      setError(err.message)
    } finally {
      setUserDeletingId(null)
    }
  }

  function renderProductTable(list) {
    return (
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>MÃ SP</th>
              <th>TÊN SẢN PHẨM</th>
              <th>DANH MỤC</th>
              <th>ĐƠN GIÁ</th>
              <th>TỒN KHO</th>
              <th>TRẠNG THÁI</th>
              {isAdmin && <th>THAO TÁC</th>}
            </tr>
          </thead>
          <tbody>
            {list.map((product) => {
              const category = categories.find(
                (item) => Number(item.id) === Number(product.category_id),
              )

              return (
                <tr key={product.id}>
                  <td className="product-code">{product.code}</td>
                  <td className="product-name">{product.name}</td>
                  <td>{category?.name || '—'}</td>
                  <td>{money(product.price)}</td>
                  <td>{product.quantity}</td>
                  <td>
                    <span
                      className={`status ${
                        product.quantity <= product.min_stock
                          ? 'low'
                          : 'available'
                      }`}
                    >
                      {product.quantity <= product.min_stock
                        ? 'Sắp hết hàng'
                        : 'Còn hàng'}
                    </span>
                  </td>
                  {isAdmin && (
                    <td>
                      <div className="heading-actions">
                        <button
                          className="outline-button"
                          onClick={() => startEditProduct(product)}
                        >
                          Sửa
                        </button>
                        <button
                          className="outline-button"
                          disabled={deletingId === product.id}
                          onClick={() => handleDeleteProduct(product)}
                        >
                          {deletingId === product.id ? 'Đang xóa...' : 'Xóa'}
                        </button>
                      </div>
                    </td>
                  )}
                </tr>
              )
            })}
            {list.length === 0 && (
              <tr>
                <td colSpan={isAdmin ? 7 : 6}>Chưa có sản phẩm nào.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    )
  }

  function renderProductForm(product, setProduct, onSubmit, title, onCancel) {
    return (
      <section className="panel">
        <div className="panel-heading">
          <div>
            <h3>{title}</h3>
            <p>Nhập thông tin sản phẩm.</p>
          </div>
        </div>

        <form onSubmit={onSubmit}>
          <div className="form-grid">
            <div className="form-field">
              <label>Mã sản phẩm</label>
              <input
                value={product.code}
                onChange={(e) =>
                  setProduct({ ...product, code: e.target.value })
                }
                required
              />
            </div>
            <div className="form-field">
              <label>Tên sản phẩm</label>
              <input
                value={product.name}
                onChange={(e) =>
                  setProduct({ ...product, name: e.target.value })
                }
                required
              />
            </div>
            <div className="form-field">
              <label>Danh mục</label>
              <select
                value={product.category_id || ''}
                onChange={(e) =>
                  setProduct({ ...product, category_id: e.target.value })
                }
              >
                <option value="">Chưa phân loại</option>
                {categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-field">
              <label>Đơn giá (VNĐ)</label>
              <input
                type="number"
                min="0"
                step="any"
                value={product.price}
                onChange={(e) =>
                  setProduct({ ...product, price: e.target.value })
                }
                required
              />
            </div>
            <div className="form-field">
              <label>Số lượng tồn</label>
              <input
                type="number"
                min="0"
                step="1"
                value={product.quantity}
                onChange={(e) =>
                  setProduct({ ...product, quantity: e.target.value })
                }
                required
              />
            </div>
            <div className="form-field">
              <label>Tồn kho tối thiểu</label>
              <input
                type="number"
                min="0"
                step="1"
                value={product.min_stock}
                onChange={(e) =>
                  setProduct({ ...product, min_stock: e.target.value })
                }
                required
              />
            </div>
          </div>
          <div className="form-actions">
            <button
              className="outline-button"
              type="button"
              onClick={onCancel}
            >
              Hủy
            </button>
            <button
              className="primary-button"
              type="submit"
              disabled={saving}
            >
              {saving ? 'Đang lưu...' : 'Lưu'}
            </button>
          </div>
        </form>
      </section>
    )
  }

  function renderDashboard() {
    const lowStock = products.filter(
      (p) => Number(p.quantity) <= Number(p.min_stock),
    )

    return (
      <>
        <div className="welcome">
          <div>
            <h2>Chào mừng trở lại! 👋</h2>
            <p>Đây là tình hình kho hàng hiện tại.</p>
          </div>
          <span className="today">● Hệ thống đang hoạt động</span>
        </div>

        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-icon purple">📦</div>
            <p>Tổng sản phẩm</p>
            <h3>{products.length}</h3>
            <span>Sản phẩm đang quản lý</span>
          </div>
          <div className="stat-card">
            <div className="stat-icon blue">▤</div>
            <p>Danh mục</p>
            <h3>{categories.length}</h3>
            <span>Nhóm sản phẩm</span>
          </div>
          <div className="stat-card">
            <div className="stat-icon green">▣</div>
            <p>Tổng số lượng tồn kho</p>
            <h3>
              {products.reduce(
                (total, product) => total + Number(product.quantity || 0),
                0,
              ).toLocaleString('vi-VN')}
            </h3>
            <span>Tổng số đơn vị hàng hóa</span>
          </div>
          <div className="stat-card">
            <div className="stat-icon orange">⚠</div>
            <p>Sắp hết hàng</p>
            <h3>{lowStock.length}</h3>
            <span>Cần kiểm tra tồn kho</span>
          </div>
          <div className="stat-card">
            <div className="stat-icon green">♙</div>
            <p>Vai trò hiện tại</p>
            <h3>{isAdmin ? 'Admin' : 'Staff'}</h3>
            <span>{username}</span>
          </div>
          <div className="stat-card">
            <div className="stat-icon purple">₫</div>
            <p>Giá trị tồn kho</p>
            <h3>{totalInventoryValue.toLocaleString('vi-VN')} ₫</h3>
            <span>Tổng số lượng × giá sản phẩm</span>
          </div>        </div>

        <div className="dashboard-extra-grid">
          <section className="panel">
            <div className="panel-heading">
              <div>
                <h3>Thống kê theo danh mục</h3>
                <p>Số sản phẩm và tổng số lượng theo từng danh mục.</p>
              </div>
            </div>
            {categoryStats.length > 0 ? (
              <div className="category-stat-list">
                {categoryStats.map((category) => (
                  <div className="category-stat-row" key={category.id}>
                    <div className="category-stat-info">
                      <strong>{category.name}</strong>
                      <span>{category.productCount} sản phẩm · {category.quantity.toLocaleString('vi-VN')} đơn vị</span>
                    </div>
                    <div className="category-bar-track">
                      <div
                        className="category-bar"
                        style={{
                          width: `${Math.max(
                            0,
                            ...categoryStats.map((item) => item.productCount),
                          ) > 0
                            ? (category.productCount /
                                Math.max(...categoryStats.map((item) => item.productCount))) * 100
                            : 0}%`,
                        }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p>Chưa có danh mục để thống kê.</p>
            )}
          </section>

          <section className="panel">
            <div className="panel-heading">
              <div>
                <h3>Thống kê nhập – xuất kho</h3>
                <p>Tổng số lượng trong lịch sử giao dịch đã ghi nhận.</p>
              </div>
            </div>
            {stockLoading ? (
              <p>Đang tải dữ liệu giao dịch...</p>
            ) : (
              <div className="movement-chart">
                <div className="movement-summary">
                  <div>
                    <span className="movement-dot import-dot" />
                    <span>Tổng nhập</span>
                    <strong>{totalImported.toLocaleString('vi-VN')}</strong>
                  </div>
                  <div>
                    <span className="movement-dot export-dot" />
                    <span>Tổng xuất</span>
                    <strong>{totalExported.toLocaleString('vi-VN')}</strong>
                  </div>
                </div>
                <div className="movement-row">
                  <span>Nhập kho</span>
                  <div className="movement-bar-track">
                    <div
                      className="movement-bar import-bar"
                      style={{ width: `${(totalImported / maxStockMovement) * 100}%` }}
                    />
                  </div>
                  <strong>{totalImported.toLocaleString('vi-VN')}</strong>
                </div>
                <div className="movement-row">
                  <span>Xuất kho</span>
                  <div className="movement-bar-track">
                    <div
                      className="movement-bar export-bar"
                      style={{ width: `${(totalExported / maxStockMovement) * 100}%` }}
                    />
                  </div>
                  <strong>{totalExported.toLocaleString('vi-VN')}</strong>
                </div>
                <p className="chart-footnote">
                  Dựa trên {stockTransactions.length} giao dịch trong lịch sử kho.
                </p>
              </div>
            )}
          </section>
        </div>
        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Sản phẩm gần đây</h3>
              <p>Dữ liệu lấy trực tiếp từ API.</p>
            </div>
            <button
              className="outline-button"
              onClick={() => setActivePage('products')}
            >
              Xem tất cả →
            </button>
          </div>
          {renderProductTable(products.slice(0, 5))}
        </section>

        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Sản phẩm cần nhập thêm</h3>
              <p>Các sản phẩm có số lượng tồn kho bằng hoặc thấp hơn mức tối thiểu.</p>
            </div>
            <button
              className="outline-button"
              onClick={() => setActivePage('stock')}
            >
              Nhập kho →
            </button>
          </div>

          {lowStock.length > 0 ? (
            <div className="table-wrapper">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Mã sản phẩm</th>
                    <th>Tên sản phẩm</th>
                    <th>Tồn kho</th>
                    <th>Tồn tối thiểu</th>
                    <th>Trạng thái</th>
                  </tr>
                </thead>
                <tbody>
                  {lowStock.map((product) => (
                    <tr key={product.id}>
                      <td>{product.code}</td>
                      <td>{product.name}</td>
                      <td>{product.quantity}</td>
                      <td>{product.min_stock}</td>
                      <td>
                        {Number(product.quantity) <= 0
                          ? 'Hết hàng'
                          : 'Cần nhập thêm'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p>Tất cả sản phẩm hiện có mức tồn kho an toàn.</p>
          )}
        </section>
        {lowStock.length > 0 && (
          <section className="notice-panel">
            <div className="notice-symbol">!</div>
            <div>
              <h4>Cần chú ý hàng tồn kho</h4>
              <p>
                Có {lowStock.length} sản phẩm đang ở mức tồn kho tối thiểu
                hoặc thấp hơn.
              </p>
            </div>
            <button
              className="outline-button"
              onClick={() => setActivePage('products')}
            >
              Kiểm tra
            </button>
          </section>
        )}
      </>
    )
  }

  function renderCategories() {
    return (
      <>
        {isAdmin && (
          <section className="panel">
            <div className="panel-heading">
              <div>
                <h3>{editingCategory ? 'Sửa danh mục' : 'Thêm danh mục mới'}</h3>
                <p>Quản lý nhóm sản phẩm trong kho.</p>
              </div>
            </div>
            <form onSubmit={handleSaveCategory}>
              <div className="form-grid">
                <div className="form-field">
                  <label>Tên danh mục</label>
                  <input
                    value={categoryName}
                    onChange={(e) => setCategoryName(e.target.value)}
                    placeholder="Ví dụ: Thiết bị văn phòng"
                    required
                  />
                </div>
                <div className="form-field">
                  <label>Mô tả</label>
                  <input
                    value={categoryDescription}
                    onChange={(e) => setCategoryDescription(e.target.value)}
                    placeholder="Mô tả (không bắt buộc)"
                  />
                </div>
              </div>
              <div className="form-actions">
                {editingCategory && (
                  <button
                    className="outline-button"
                    type="button"
                    onClick={() => {
                      setEditingCategory(null)
                      setCategoryName('')
                      setCategoryDescription('')
                    }}
                  >
                    Hủy sửa
                  </button>
                )}
                <button
                  className="primary-button"
                  type="submit"
                  disabled={saving}
                >
                  {saving
                    ? 'Đang lưu...'
                    : editingCategory
                      ? 'Lưu thay đổi'
                      : '+ Thêm danh mục'}
                </button>
              </div>
            </form>
          </section>
        )}

        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Danh sách danh mục</h3>
              <p>Tổng số: {categories.length} danh mục</p>
            </div>
            <button className="outline-button" onClick={loadCategories}>
              ↻ Làm mới
            </button>
          </div>
          {categoryLoading ? (
            <p>Đang tải danh mục...</p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>TÊN DANH MỤC</th>
                    <th>MÔ TẢ</th>
                    <th>NGÀY TẠO</th>
                    {isAdmin && <th>THAO TÁC</th>}
                  </tr>
                </thead>
                <tbody>
                  {categories.map((category) => (
                    <tr key={category.id}>
                      <td>{category.id}</td>
                      <td className="product-name">{category.name}</td>
                      <td>{category.description || '—'}</td>
                      <td>
                        {category.created_at
                          ? new Date(category.created_at).toLocaleDateString(
                              'vi-VN',
                            )
                          : '—'}
                      </td>
                      {isAdmin && (
                        <td>
                          <div className="heading-actions">
                            <button
                              className="outline-button"
                              onClick={() => startEditCategory(category)}
                            >
                              Sửa
                            </button>
                            <button
                              className="outline-button"
                              onClick={() => handleDeleteCategory(category)}
                            >
                              Xóa
                            </button>
                          </div>
                        </td>
                      )}
                    </tr>
                  ))}
                  {categories.length === 0 && (
                    <tr>
                      <td colSpan={isAdmin ? 5 : 4}>
                        Chưa có danh mục nào.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </>
    )
  }

  function renderUsers() {
    if (!isAdmin) {
      return (
        <section className="panel">
          <p>Chỉ Admin mới được quản lý người dùng.</p>
        </section>
      )
    }

    return (
      <>
        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>
                {editingUserId === null
                  ? 'Thêm người dùng'
                  : `Sửa người dùng #${editingUserId}`}
              </h3>
              <p>Tạo tài khoản và phân quyền truy cập hệ thống.</p>
            </div>
            <button className="outline-button" onClick={cancelUserForm}>
              Làm mới form
            </button>
          </div>

          <form
            onSubmit={
              editingUserId === null
                ? handleCreateUser
                : handleUpdateUser
            }
          >
            <div className="form-grid">
              <div className="form-field">
                <label>Tên đăng nhập</label>
                <input
                  value={userForm.username}
                  onChange={(e) =>
                    setUserForm({ ...userForm, username: e.target.value })
                  }
                  placeholder="Nhập tên đăng nhập"
                  required
                />
              </div>

              {editingUserId === null && (
                <div className="form-field">
                  <label>Mật khẩu</label>
                  <input
                    type="password"
                    value={userForm.password}
                    onChange={(e) =>
                      setUserForm({ ...userForm, password: e.target.value })
                    }
                    placeholder="Nhập mật khẩu"
                    required
                  />
                </div>
              )}

              <div className="form-field">
                <label>Họ và tên</label>
                <input
                  value={userForm.full_name}
                  onChange={(e) =>
                    setUserForm({ ...userForm, full_name: e.target.value })
                  }
                  placeholder="Nhập họ và tên"
                  required
                />
              </div>

              <div className="form-field">
                <label>Vai trò</label>
                <select
                  value={userForm.role}
                  onChange={(e) =>
                    setUserForm({ ...userForm, role: e.target.value })
                  }
                >
                  <option value="staff">Staff</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
            </div>

            <div className="form-actions">
              <button
                type="button"
                className="outline-button"
                onClick={cancelUserForm}
              >
                Hủy
              </button>
              <button
                type="submit"
                className="primary-button"
                disabled={userSaving}
              >
                {userSaving
                  ? 'Đang lưu...'
                  : editingUserId === null
                    ? 'Thêm người dùng'
                    : 'Lưu thay đổi'}
              </button>
            </div>
          </form>
        </section>

        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Danh sách người dùng</h3>
              <p>Tổng số: {users.length} tài khoản</p>
            </div>
            <button className="outline-button" onClick={loadUsers}>
              ↻ Làm mới
            </button>
          </div>

          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>TÊN ĐĂNG NHẬP</th>
                  <th>HỌ VÀ TÊN</th>
                  <th>VAI TRÒ</th>
                  <th>THAO TÁC</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id}>
                    <td>{user.id}</td>
                    <td className="product-code">{user.username}</td>
                    <td>{user.full_name || '—'}</td>
                    <td>
                      <span
                        className={`status ${
                          user.role === 'admin' ? 'available' : 'low'
                        }`}
                      >
                        {user.role}
                      </span>
                    </td>
                    <td>
                      <div className="heading-actions">
                        <button
                          className="outline-button"
                          onClick={() => startEditUser(user)}
                        >
                          Sửa
                        </button>
                        <button
                          className="outline-button"
                          disabled={userDeletingId === user.id}
                          onClick={() => handleDeleteUser(user)}
                        >
                          {userDeletingId === user.id
                            ? 'Đang xóa...'
                            : 'Xóa'}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {users.length === 0 && (
                  <tr>
                    <td colSpan={5}>
                      Chưa có dữ liệu người dùng.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>
      </>
    )
  }

  function renderStock() {
    return (
      <>
        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Nhập / Xuất kho</h3>
              <p>Ghi nhận số lượng hàng hóa nhập vào hoặc xuất ra.</p>
            </div>
            <button
              className="outline-button"
              onClick={loadStockTransactions}
              disabled={stockLoading}
            >
              Làm mới lịch sử
            </button>
          </div>

          <form onSubmit={handleStockSubmit}>
            <div className="form-grid">
              <div className="form-field">
                <label>Loại giao dịch</label>
                <select
                  value={stockType}
                  onChange={(e) => setStockType(e.target.value)}
                >
                  <option value="import">Nhập kho</option>
                  <option value="export">Xuất kho</option>
                </select>
              </div>

              <div className="form-field">
                <label>Sản phẩm</label>
                <select
                  value={stockForm.product_id}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, product_id: e.target.value })
                  }
                  required
                >
                  <option value="">-- Chọn sản phẩm --</option>
                  {products.map((product) => (
                    <option key={product.id} value={product.id}>
                      {product.code} - {product.name} (Tồn: {product.quantity})
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-field">
                <label>Số lượng</label>
                <input
                  type="number"
                  min="1"
                  step="1"
                  value={stockForm.quantity}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, quantity: e.target.value })
                  }
                  required
                />
              </div>

              <div className="form-field">
                <label>Ghi chú</label>
                <input
                  value={stockForm.note}
                  onChange={(e) =>
                    setStockForm({ ...stockForm, note: e.target.value })
                  }
                  placeholder="Ví dụ: Nhập hàng mới"
                />
              </div>
            </div>

            <div className="form-actions">
              <button
                type="submit"
                className="primary-button"
                disabled={stockSaving}
              >
                {stockSaving
                  ? 'Đang xử lý...'
                  : stockType === 'import'
                    ? 'Xác nhận nhập kho'
                    : 'Xác nhận xuất kho'}
              </button>
            </div>
          </form>
        </section>

        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Lịch sử giao dịch kho</h3>
              <p>Tổng số: {stockTransactions.length} giao dịch</p>
            </div>
          </div>

          {stockLoading ? (
            <p>Đang tải lịch sử...</p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>SẢN PHẨM</th>
                    <th>LOẠI</th>
                    <th>SỐ LƯỢNG</th>
                    <th>NGƯỜI THỰC HIỆN</th>
                    <th>GHI CHÚ</th>
                    <th>THỜI GIAN</th>
                  </tr>
                </thead>
                <tbody>
                  {stockTransactions.map((item) => {
                    const product = products.find(
                      (p) => p.id === item.product_id,
                    )
                    return (
                      <tr key={item.id}>
                        <td>{item.id}</td>
                        <td>
                          {product
                            ? `${product.code} - ${product.name}`
                            : `ID ${item.product_id}`}
                        </td>
                        <td>
                          {item.transaction_type === 'import' ||
                          item.transaction_type === 'in'
                            ? 'Nhập kho'
                            : item.transaction_type === 'export' ||
                                item.transaction_type === 'out'
                              ? 'Xuất kho'
                              : item.transaction_type || '—'}
                        </td>
                        <td>{item.quantity}</td>
                        <td>{item.user_id ?? '—'}</td>
                        <td>{item.note || '—'}</td>
                        <td>
                          {item.created_at
                            ? new Date(item.created_at).toLocaleString('vi-VN')
                            : '—'}
                        </td>
                      </tr>
                    )
                  })}
                  {stockTransactions.length === 0 && (
                    <tr>
                      <td colSpan={7}>Chưa có giao dịch kho.</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </>
    )
  }
  function renderContent() {
    if (activePage === 'dashboard') return renderDashboard()

    if (activePage === 'products') {
      return (
        <>
          <section className="panel">
            <div className="panel-heading">
              <div>
                <h3>Danh sách sản phẩm</h3>
                <p>Quản lý thông tin và số lượng hàng hóa.</p>
              </div>
              <div className="heading-actions">
                <button className="outline-button" onClick={loadProducts}>
                  ↻ Làm mới
                </button>
                {isAdmin && (
                  <button
                    className="primary-button"
                    onClick={() => {
                      setShowProductForm(!showProductForm)
                      setEditingProductId(null)
                    }}
                  >
                    {showProductForm ? 'Đóng biểu mẫu' : '+ Thêm sản phẩm'}
                  </button>
                )}
              </div>
            </div>
          </section>

          {showProductForm &&
            isAdmin &&
            renderProductForm(
              newProduct,
              setNewProduct,
              handleCreateProduct,
              'Thêm sản phẩm mới',
              () => setShowProductForm(false),
            )}

          {editingProductId !== null &&
            isAdmin &&
            renderProductForm(
              editProduct,
              setEditProduct,
              handleUpdateProduct,
              'Sửa sản phẩm',
              () => setEditingProductId(null),
            )}

          <section className="panel">
            <div className="panel-heading">
              <div>
                <h3>Sản phẩm trong kho</h3>
                <p>Tổng số: {products.length} sản phẩm</p>
              </div>
            </div>
            <div className="form-grid">
              <div className="form-field">
                <label>Tìm theo mã hoặc tên sản phẩm</label>
                <input
                  value={productSearch}
                  onChange={(e) => {
                    setProductSearch(e.target.value)
                    setProductPage(1)
                  }}
                  placeholder="Nhập mã hoặc tên sản phẩm..."
                />
              </div>

              <div className="form-field">
                <label>Lọc theo danh mục</label>
                <select
                  value={productCategoryFilter}
                  onChange={(e) => {
                    setProductCategoryFilter(e.target.value)
                    setProductPage(1)
                  }}
                >
                  <option value="all">Tất cả danh mục</option>
                  {categories.map((category) => (
                    <option key={category.id} value={String(category.id)}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-field">
                <label>Trạng thái tồn kho</label>
                <select
                  value={productStockFilter}
                  onChange={(e) => {
                    setProductStockFilter(e.target.value)
                    setProductPage(1)
                  }}
                >
                  <option value="all">Tất cả trạng thái</option>
                  <option value="out">Hết hàng</option>
                  <option value="low">Cần nhập thêm</option>
                  <option value="safe">Tồn kho an toàn</option>
                </select>
              </div>
            </div>

            <p>
              Tìm thấy {filteredProducts.length} sản phẩm.
            </p>

            {loading ? (
              <p>Đang tải sản phẩm...</p>
            ) : (
              renderProductTable(paginatedProducts)
            )}

            <div className="heading-actions">
              <button
                className="outline-button"
                disabled={currentProductPage <= 1}
                onClick={() =>
                  setProductPage((page) => Math.max(1, page - 1))
                }
              >
                ← Trang trước
              </button>
              <span>
                Trang {currentProductPage} / {totalProductPages}
              </span>
              <button
                className="outline-button"
                disabled={currentProductPage >= totalProductPages}
                onClick={() =>
                  setProductPage((page) =>
                    Math.min(totalProductPages, page + 1)
                  )
                }
              >
                Trang sau →
              </button>
            </div>
          </section>
        </>
      )
    }

    if (activePage === 'categories') return renderCategories()
    if (activePage === 'users') return renderUsers()
    if (activePage === 'stock') return renderStock()

    return null
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
          <div className="brand-icon">W</div>
          <div>
            <h1>WareHouse</h1>
            <span>MANAGEMENT SYSTEM</span>
          </div>
        </div>

        <p className="menu-label">MENU CHÍNH</p>
        <nav className="side-menu">
          {menuItems.map((item) => (
            <button
              key={item.id}
              className={`menu-item ${
                activePage === item.id ? 'active' : ''
              }`}
              type="button"
              onClick={() => {
                setActivePage(item.id)
                setError('')
              }}
            >
              <span className="menu-icon">{item.icon}</span>
              {item.label}
              {activePage === item.id && (
                <span className="active-dot" />
              )}
            </button>
          ))}
        </nav>

        <div className="sidebar-bottom">
          <div className="help-icon">?</div>
          <div>
            <strong>Warehouse System</strong>
            <p>{isAdmin ? 'Quản trị viên' : 'Nhân viên'}</p>
          </div>
        </div>
      </aside>

      <main className="main-area">
        <header className="topbar">
          <div className="breadcrumb">
            <span>Trang chủ</span>
            <span>/</span>
            <strong>{currentMenu?.label}</strong>
          </div>
          <div className="user-area">
            <div className="user-avatar">
              {(username || 'U').charAt(0).toUpperCase()}
            </div>
            <div className="user-info">
              <strong>{username}</strong>
              <span>{isAdmin ? 'Quản trị viên' : 'Nhân viên'}</span>
            </div>
            <button className="outline-button" onClick={handleLogout}>
              Đăng xuất
            </button>
          </div>
        </header>

        <div className="page-content">
          <div className="page-title">
            <div>
              <h2>{currentMenu?.label}</h2>
              <p>Hệ thống quản lý kho hàng</p>
            </div>
          </div>

          {error && <p className="error-message">{error}</p>}
          {renderContent()}
          <footer>Warehouse Management System · Go + React</footer>
        </div>
      </main>
    </div>
  )
}

export default App



