import { type ReactNode } from 'react'

interface Props {
  title: string
  onClose: () => void
  children: ReactNode
}

export default function Modal({ title, onClose, children }: Props) {
  return (
    <div style={overlay}>
      <div style={box}>
        <div style={header}>
          <h3 style={{ margin: 0 }}>{title}</h3>
          <button onClick={onClose} style={closeBtn}>&#x2715;</button>
        </div>
        <div style={content}>
          {children}
        </div>
      </div>
    </div>
  )
}

const overlay: React.CSSProperties = {
  position: 'fixed', inset: 0,
  background: 'rgba(0,0,0,.45)',
  display: 'flex', alignItems: 'center', justifyContent: 'center',
  zIndex: 1000,
}
const box: React.CSSProperties = {
  background: '#fff', borderRadius: 10,
  padding: '24px 28px', minWidth: 360, maxWidth: 560, width: '100%',
  maxHeight: '90vh',
  display: 'flex', flexDirection: 'column',
  boxShadow: '0 8px 32px rgba(0,0,0,.18)',
}
const header: React.CSSProperties = {
  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
  marginBottom: 20, flexShrink: 0,
}
const closeBtn: React.CSSProperties = {
  background: 'none', border: 'none', fontSize: 18,
  cursor: 'pointer', color: '#888',
}
const content: React.CSSProperties = {
  flex: 1,
  overflowY: 'auto',
  paddingRight: 4,
}
