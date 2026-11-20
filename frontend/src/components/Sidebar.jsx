function Sidebar({ menuItems, activePage, setActivePage, setError, isAdmin }) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <div className="brand-icon">W</div>
        <div>
          <h1>WareHouse</h1>
          <span>MANAGEMENT SYSTEM</span>
        </div>
      </div>

      <p className="menu-label">MENU CHÍNH</p>
      <nav className="side-menu">
        {menuItems.map((item) => (
          <button
            key={item.id}
            className={`menu-item ${
              activePage === item.id ? 'active' : ''
            }`}
            type="button"
            onClick={() => {
              setActivePage(item.id)
              setError('')
            }}
          >
            <span className="menu-icon">{item.icon}</span>
            {item.label}
            {activePage === item.id && (
              <span className="active-dot" />
            )}
          </button>
        ))}
      </nav>

      <div className="sidebar-bottom">
        <div className="help-icon">?</div>
        <div>
          <strong>Warehouse System</strong>
          <p>{isAdmin ? 'Quản trị viên' : 'Nhân viên'}</p>
        </div>
      </div>
    </aside>
  )
}

export default Sidebar