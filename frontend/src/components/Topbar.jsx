function Topbar({ currentMenu, username, isAdmin, onLogout }) {
  return (
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
        <button className="outline-button" onClick={onLogout}>
          Đăng xuất
        </button>
      </div>
    </header>
  )
}

export default Topbar