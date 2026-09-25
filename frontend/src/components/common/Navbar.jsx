import React from 'react';
import { ShieldCheck, LogIn, LogOut, User, Activity } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

export default function Navbar({ currentView, onNavigate, onLoginClick }) {
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <nav
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 100,
        background: 'rgba(9, 13, 22, 0.85)',
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        borderBottom: '1px solid var(--border-subtle)',
      }}
    >
      <div
        style={{
          maxWidth: '1280px',
          margin: '0 auto',
          padding: '0 1.5rem',
          height: '64px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        {/* Logo / Brand */}
        <div
          onClick={() => onNavigate('verify')}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.6rem',
            cursor: 'pointer',
            userSelect: 'none',
          }}
        >
          <div
            style={{
              width: '34px', height: '34px',
              borderRadius: 'var(--radius-md)',
              background: 'linear-gradient(135deg, var(--primary) 0%, var(--success) 100%)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              boxShadow: '0 2px 10px rgba(14, 165, 233, 0.35)',
            }}
          >
            <ShieldCheck size={19} color="#fff" />
          </div>
          <div>
            <span style={{ fontFamily: 'var(--font-display)', fontWeight: 800, fontSize: '1.15rem', letterSpacing: '-0.02em' }}>
              MedVerify
            </span>
            <span style={{ fontSize: '0.6rem', color: 'var(--text-dim)', marginLeft: '0.4rem', verticalAlign: 'super', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              BETA
            </span>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div style={{ display: 'flex', gap: '0.25rem' }}>
          <NavTab
            active={currentView === 'verify'}
            onClick={() => onNavigate('verify')}
            label="Verify Drug"
          />
          {isAuthenticated && (
            <>
              <NavTab
                active={currentView === 'dashboard'}
                onClick={() => onNavigate('dashboard')}
                label="Dashboard"
              />
              <NavTab
                active={currentView === 'drugs'}
                onClick={() => onNavigate('drugs')}
                label="Drugs"
              />
              <NavTab
                active={currentView === 'manufacturers'}
                onClick={() => onNavigate('manufacturers')}
                label="Manufacturers"
              />
              <NavTab
                active={currentView === 'audit'}
                onClick={() => onNavigate('audit')}
                label="Audit Logs"
              />
            </>
          )}
        </div>

        {/* Right Side: Auth */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          {isAuthenticated ? (
            <>
              {/* Online Indicator */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <div style={{
                  width: '8px', height: '8px', borderRadius: '50%',
                  background: 'var(--success)', boxShadow: '0 0 8px var(--success-glow)',
                }} />
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>ONLINE</span>
              </div>

              <div
                style={{
                  display: 'flex', alignItems: 'center', gap: '0.5rem',
                  padding: '0.35rem 0.75rem', borderRadius: 'var(--radius-full)',
                  background: 'rgba(255, 255, 255, 0.05)', border: '1px solid var(--border-subtle)',
                }}
              >
                <User size={14} color="var(--primary)" />
                <span style={{ fontSize: '0.8rem', fontWeight: 500, color: 'var(--text-muted)' }}>
                  {user?.full_name?.split(' ')[0] || user?.email}
                </span>
              </div>

              <button onClick={logout} className="btn btn-ghost btn-sm" style={{ color: 'var(--text-dim)' }}>
                <LogOut size={16} />
              </button>
            </>
          ) : (
            <button onClick={onLoginClick} className="btn btn-primary btn-sm">
              <LogIn size={15} />
              <span>Pharmacist Login</span>
            </button>
          )}
        </div>
      </div>
    </nav>
  );
}

function NavTab({ active, onClick, label }) {
  return (
    <button
      onClick={onClick}
      style={{
        padding: '0.45rem 0.85rem',
        fontSize: '0.825rem',
        fontWeight: active ? 600 : 500,
        color: active ? 'var(--text-main)' : 'var(--text-muted)',
        background: active ? 'rgba(14, 165, 233, 0.12)' : 'transparent',
        border: active ? '1px solid rgba(14, 165, 233, 0.25)' : '1px solid transparent',
        borderRadius: 'var(--radius-sm)',
        cursor: 'pointer',
        transition: 'all var(--transition-fast)',
        fontFamily: 'var(--font-sans)',
      }}
      onMouseEnter={(e) => {
        if (!active) {
          e.target.style.color = 'var(--text-main)';
          e.target.style.background = 'rgba(255, 255, 255, 0.04)';
        }
      }}
      onMouseLeave={(e) => {
        if (!active) {
          e.target.style.color = 'var(--text-muted)';
          e.target.style.background = 'transparent';
        }
      }}
    >
      {label}
    </button>
  );
}
