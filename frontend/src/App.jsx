
import { useEffect, useState } from 'react'
import { apiRequest, login, logout } from './api/client'
import './App.css'
import Login from './components/Login'
import Sidebar from './components/Sidebar'
import Topbar from './components/Topbar'
import Dashboard from './components/Dashboard'
import Categories from './components/Categories'
import Users from './components/Users'
import Stock from './components/Stock'
import Products from './components/Products'

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

  async function handleLogout() {
    const refreshToken = localStorage.getItem('refresh_token')

    try {
      if (refreshToken) {
        await logout(refreshToken)
      }
    } catch (err) {
      console.error('Khong the thu hoi refresh token:', err)
    } finally {
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
  }

  function makeProductPayload(product, includeQuantity = true) {
    const payload = {
      code: product.code.trim(),
      name: product.name.trim(),
      category_id: product.category_id
        ? Number(product.category_id)
        : null,
      price: Number(product.price),
      min_stock: Number(product.min_stock),
    }

    if (includeQuantity) {
      payload.quantity = Number(product.quantity)
    }

    return payload
  }

  async function handleCreateProduct(e) {
    e.preventDefault()
    setError('')
    setSaving(true)

    try {
      await apiRequest('/products', {
        method: 'POST',
        body: JSON.stringify(makeProductPayload(newProduct, false)),
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
      version: product.version,
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
        body: JSON.stringify({
          ...makeProductPayload(editProduct, false),
          version: editProduct.version,
        }),
      })

      setEditingProductId(null)
      setEditProduct({ ...emptyProduct })
      await loadProducts()
    } catch (err) {
      if (err.message.includes('San pham da duoc cap nhat boi nguoi khac')) {
        setEditingProductId(null)
        setEditProduct({ ...emptyProduct })
        await loadProducts()
        setError('Sản phẩm đã được người khác cập nhật. Danh sách đã được tải lại, hãy chọn sản phẩm và sửa lại.')
      } else {
        setError(err.message)
      }
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

  function renderCategories() {
    return (
      <Categories
        categories={categories}
        isAdmin={isAdmin}
        categoryName={categoryName}
        setCategoryName={setCategoryName}
        categoryDescription={categoryDescription}
        setCategoryDescription={setCategoryDescription}
        editingCategory={editingCategory}
        setEditingCategory={setEditingCategory}
        saving={saving}
        categoryLoading={categoryLoading}
        onSave={handleSaveCategory}
        onRefresh={loadCategories}
        onEdit={startEditCategory}
        onDelete={handleDeleteCategory}
      />
    )
  }

  function renderUsers() {
    return (
      <Users
        isAdmin={isAdmin}
        users={users}
        userForm={userForm}
        setUserForm={setUserForm}
        editingUserId={editingUserId}
        userSaving={userSaving}
        userDeletingId={userDeletingId}
        onCreate={handleCreateUser}
        onUpdate={handleUpdateUser}
        onCancel={cancelUserForm}
        onRefresh={loadUsers}
        onEdit={startEditUser}
        onDelete={handleDeleteUser}
      />
    )
  }

  function renderStock() {
    return (
      <Stock
        products={products}
        stockType={stockType}
        setStockType={setStockType}
        stockForm={stockForm}
        setStockForm={setStockForm}
        stockSaving={stockSaving}
        stockLoading={stockLoading}
        stockTransactions={stockTransactions}
        onSubmit={handleStockSubmit}
        onRefresh={loadStockTransactions}
      />
    )
  }

  function renderContent() {
    if (activePage === 'dashboard') {
      return (
        <Dashboard
          products={products}
          categories={categories}
          isAdmin={isAdmin}
          username={username}
          totalInventoryValue={totalInventoryValue}
          categoryStats={categoryStats}
          stockLoading={stockLoading}
          totalImported={totalImported}
          totalExported={totalExported}
          maxStockMovement={maxStockMovement}
          stockTransactions={stockTransactions}
          deletingId={deletingId}
          onEdit={startEditProduct}
          onDelete={handleDeleteProduct}
          formatMoney={money}
          onNavigate={setActivePage}
        />
      )
    }

    if (activePage === 'products') {
      return (
        <Products
          products={products}
          categories={categories}
          isAdmin={isAdmin}
          loading={loading}
          saving={saving}
          deletingId={deletingId}
          showProductForm={showProductForm}
          setShowProductForm={setShowProductForm}
          newProduct={newProduct}
          setNewProduct={setNewProduct}
          editingProductId={editingProductId}
          setEditingProductId={setEditingProductId}
          editProduct={editProduct}
          setEditProduct={setEditProduct}
          onCreate={handleCreateProduct}
          onUpdate={handleUpdateProduct}
          onEdit={startEditProduct}
          onDelete={handleDeleteProduct}
          onRefresh={loadProducts}
          productSearch={productSearch}
          setProductSearch={setProductSearch}
          productCategoryFilter={productCategoryFilter}
          setProductCategoryFilter={setProductCategoryFilter}
          productStockFilter={productStockFilter}
          setProductStockFilter={setProductStockFilter}
          filteredProducts={filteredProducts}
          paginatedProducts={paginatedProducts}
          currentProductPage={currentProductPage}
          totalProductPages={totalProductPages}
          setProductPage={setProductPage}
          formatMoney={money}
        />
      )
    }

    if (activePage === 'categories') return renderCategories()
    if (activePage === 'users') return renderUsers()
    if (activePage === 'stock') return renderStock()

    return null
  }

  if (!token) {
    return (
      <Login
        username={username}
        setUsername={setUsername}
        password={password}
        setPassword={setPassword}
        error={error}
        onSubmit={handleLogin}
      />
    )
  }

  return (
    <div className="app-layout">
      <Sidebar
        menuItems={menuItems}
        activePage={activePage}
        setActivePage={setActivePage}
        setError={setError}
        isAdmin={isAdmin}
      />

      <main className="main-area">
        <Topbar
            currentMenu={currentMenu}
            username={username}
            isAdmin={isAdmin}
            onLogout={handleLogout}
          />

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







