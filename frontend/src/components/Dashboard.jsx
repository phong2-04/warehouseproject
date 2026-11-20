import ProductTable from './ProductTable'

function Dashboard({
  products,
  categories,
  isAdmin,
  username,
  totalInventoryValue,
  categoryStats,
  stockLoading,
  totalImported,
  totalExported,
  maxStockMovement,
  stockTransactions,
  deletingId,
  onEdit,
  onDelete,
  formatMoney,
  onNavigate,
}) {
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
              onClick={() => onNavigate('products')}
            >
              Xem tất cả →
            </button>
          </div>
          <ProductTable
            list={products.slice(0, 5)}
            categories={categories}
            isAdmin={isAdmin}
            deletingId={deletingId}
            onEdit={onEdit}
            onDelete={onDelete}
            formatMoney={formatMoney}
          />
        </section>

        <section className="panel">
          <div className="panel-heading">
            <div>
              <h3>Sản phẩm cần nhập thêm</h3>
              <p>Các sản phẩm có số lượng tồn kho bằng hoặc thấp hơn mức tối thiểu.</p>
            </div>
            <button
              className="outline-button"
              onClick={() => onNavigate('stock')}
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
              onClick={() => onNavigate('products')}
            >
              Kiểm tra
            </button>
          </section>
        )}
      </>
    )
  }

export default Dashboard


