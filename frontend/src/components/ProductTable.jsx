function ProductTable({
  list,
  categories,
  isAdmin,
  deletingId,
  onEdit,
  onDelete,
  formatMoney,
}) {
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
                <td>{formatMoney(product.price)}</td>
                <td>{product.quantity}</td>
                <td>
                  <span
                    className={`status ${
                      Number(product.quantity) === 0
                        ? 'out'
                        : Number(product.quantity) <= Number(product.min_stock)
                          ? 'low'
                          : 'available'
                    }`}
                  >
                    {Number(product.quantity) === 0
                      ? 'Hết hàng'
                      : Number(product.quantity) <= Number(product.min_stock)
                        ? 'Cần nhập thêm'
                        : 'Còn hàng'}
                  </span>
                </td>
                {isAdmin && (
                  <td>
                    <div className="heading-actions">
                      <button
                        className="outline-button"
                        onClick={() => onEdit(product)}
                      >
                        Sửa
                      </button>
                      <button
                        className="outline-button"
                        disabled={deletingId === product.id}
                        onClick={() => onDelete(product)}
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

export default ProductTable