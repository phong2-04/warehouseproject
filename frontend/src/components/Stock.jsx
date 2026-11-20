function Stock({
  products,
  stockType,
  setStockType,
  stockForm,
  setStockForm,
  stockSaving,
  stockLoading,
  stockTransactions,
  onSubmit,
  onRefresh,
}) {
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
            type="button"
            onClick={onRefresh}
            disabled={stockLoading}
          >
            Làm mới lịch sử
          </button>
        </div>

        <form onSubmit={onSubmit}>
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

export default Stock
