function Login({
  username,
  setUsername,
  password,
  setPassword,
  error,
  onSubmit,
}) {
  return (
    <div className="login-page">
      <form className="login-card" onSubmit={onSubmit}>
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

export default Login