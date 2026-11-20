function Users({
  isAdmin,
  users,
  userForm,
  setUserForm,
  editingUserId,
  userSaving,
  userDeletingId,
  onCreate,
  onUpdate,
  onCancel,
  onRefresh,
  onEdit,
  onDelete,
}) {
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

          <button
            className="outline-button"
            type="button"
            onClick={onCancel}
          >
            Làm mới form
          </button>
        </div>

        <form
          onSubmit={editingUserId === null ? onCreate : onUpdate}
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
              onClick={onCancel}
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

          <button
            className="outline-button"
            type="button"
            onClick={onRefresh}
          >
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
                        type="button"
                        onClick={() => onEdit(user)}
                      >
                        Sửa
                      </button>

                      <button
                        className="outline-button"
                        type="button"
                        disabled={userDeletingId === user.id}
                        onClick={() => onDelete(user)}
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

export default Users
