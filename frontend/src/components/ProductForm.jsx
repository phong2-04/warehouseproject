function ProductForm({
  product,
  setProduct,
  onSubmit,
  title,
  onCancel,
  showQuantity = false,
  saving = false,
}) {
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
              {product.categories?.map((category) => (
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

          {showQuantity && (
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
          )}

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

export default ProductForm