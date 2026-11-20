function Categories({
  categories,
  isAdmin,
  categoryName,
  setCategoryName,
  categoryDescription,
  setCategoryDescription,
  editingCategory,
  setEditingCategory,
  saving,
  categoryLoading,
  onSave,
  onRefresh,
  onEdit,
  onDelete,
}) {
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

          <form onSubmit={onSave}>
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

          <button
            className="outline-button"
            onClick={onRefresh}
            type="button"
          >
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
                            onClick={() => onEdit(category)}
                            type="button"
                          >
                            Sửa
                          </button>

                          <button
                            className="outline-button"
                            onClick={() => onDelete(category)}
                            type="button"
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

export default Categories
