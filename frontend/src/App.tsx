import { NavLink, Route, Routes } from 'react-router-dom'
import EmployeesPage from './pages/EmployeesPage'
import ClientsPage from './pages/ClientsPage'
import ThemesPage from './pages/ThemesPage'
import AppealsPage from './pages/AppealsPage'
import AppealDetailPage from './pages/AppealDetailPage'
import SlotsPage from './pages/SlotsPage'
import SubthemesPage from './pages/SubthemesPage'

const navItems = [
  { to: '/employees', label: '👤 Сотрудники' },
  { to: '/clients',   label: '🙋 Клиенты' },
  { to: '/themes',    label: '🗂️ Темы' },
  { to: '/subthemes', label: '🏷️ Подтемы' },
  { to: '/appeals',   label: '📋 Обращения' },
  { to: '/slots',     label: '🔲 Слоты' },
]

export default function App() {
  return (
    <div style={layout}>
      <aside style={sidebar}>
        <div style={logo}>
          <div style={logoTitle}>Система распределения обращений</div>
        </div>
        <nav style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
          {navItems.map(({ to, label }) => (
            <NavLink key={to} to={to} style={({ isActive }) => ({ ...navLink, ...(isActive ? navActive : {}) })}>
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <main style={main}>
        <Routes>
          <Route path="/" element={<EmployeesPage />} />
          <Route path="/employees" element={<EmployeesPage />} />
          <Route path="/clients"   element={<ClientsPage />} />
          <Route path="/themes"    element={<ThemesPage />} />
          <Route path="/appeals"       element={<AppealsPage />} />
          <Route path="/appeals/:id"   element={<AppealDetailPage />} />
          <Route path="/slots"     element={<SlotsPage />} />
          <Route path="/subthemes" element={<SubthemesPage />} />
        </Routes>
      </main>
    </div>
  )
}

const layout: React.CSSProperties = {
  display: 'flex', minHeight: '100vh', fontFamily: 'system-ui, sans-serif', background: '#f4f5fb',
}
const sidebar: React.CSSProperties = {
  width: 220, background: '#1e2140', color: '#fff',
  display: 'flex', flexDirection: 'column', padding: '24px 16px',
  position: 'sticky', top: 0, height: '100vh',
}
const logo: React.CSSProperties = {
  marginBottom: 36, paddingLeft: 8,
}
const logoTitle: React.CSSProperties = {
  fontSize: 14, fontWeight: 700, color: '#7c8cf8',
  lineHeight: 1.35, letterSpacing: 0.3,
}
const navLink: React.CSSProperties = {
  display: 'block', padding: '10px 14px', borderRadius: 8,
  color: '#aab0d0', textDecoration: 'none', fontSize: 14, fontWeight: 500,
  transition: 'background .15s, color .15s',
}
const navActive: React.CSSProperties = {
  background: '#2d3260', color: '#fff',
}
const main: React.CSSProperties = {
  flex: 1, padding: '32px 36px', overflowY: 'auto',
}
