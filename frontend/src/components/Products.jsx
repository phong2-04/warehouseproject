function Products({
  products,
  categories,
  isAdmin,
  loading,
  saving,
  deletingId,
  showProductForm,
  setShowProductForm,
  newProduct,
  setNewProduct,
  editingProductId,
  setEditingProductId,
  editProduct,
  setEditProduct,
  onCreate,
  onUpdate,
  onEdit,
  onDelete,
  onRefresh,
  productSearch,
  setProductSearch,
  productCategoryFilter,
  setProductCategoryFilter,
  productStockFilter,
  setProductStockFilter,
  filteredProducts,
  paginatedProducts,
  currentProductPage,
  totalProductPages,
  setProductPage,
  formatMoney,
}) {
  return (
    <>
      <section className="panel">
        <div className="panel-heading">
          <div>
            <h3>Danh sách sản phẩm</h3>
            <p>Quản lý thông tin và số lượng hàng hóa.</p>
          </div>
          <div className="heading-actions">
            <button className="outline-button" onClick={onRefresh}>
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

      {showProductForm && isAdmin && (
        <ProductForm
          product={{ ...newProduct, categories }}
          setProduct={setNewProduct}
          onSubmit={onCreate}
          title="Thêm sản phẩm mới"
          onCancel={() => setShowProductForm(false)}
          saving={saving}
        />
      )}

      {editingProductId !== null && isAdmin && (
        <ProductForm
          product={{ ...editProduct, categories }}
          setProduct={setEditProduct}
          onSubmit={onUpdate}
          title="Sửa sản phẩm"
          onCancel={() => setEditingProductId(null)}
          saving={saving}
        />
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

        <p>Tìm thấy {filteredProducts.length} sản phẩm.</p>

        {loading ? (
          <p>Đang tải sản phẩm...</p>
        ) : (
          <ProductTable
            list={paginatedProducts}
            categories={categories}
            isAdmin={isAdmin}
            deletingId={deletingId}
            onEdit={onEdit}
            onDelete={onDelete}
            formatMoney={formatMoney}
          />
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

import ProductForm from './ProductForm'
import ProductTable from './ProductTable'

export default Products
