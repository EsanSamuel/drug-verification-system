import React, { useState } from 'react';
import { ShieldCheck, LogIn, LogOut, User, Menu, X } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

export default function Navbar({ currentView, onNavigate, onLoginClick }) {
  const { isAuthenticated, user, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const navigate = (view) => {
    onNavigate(view);
    setMenuOpen(false);
  };

  return (
    <nav className={`navbar${menuOpen ? ' is-open' : ''}`}>
      <div className="navbar-inner">
        {/* Logo / Brand */}
        <div
          onClick={() => navigate('verify')}
          className="navbar-brand"
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

        <button
          className="navbar-menu-toggle"
          type="button"
          aria-label={menuOpen ? 'Close navigation menu' : 'Open navigation menu'}
          aria-expanded={menuOpen}
          aria-controls="primary-navigation"
          onClick={() => setMenuOpen((open) => !open)}
        >
          {menuOpen ? <X size={21} /> : <Menu size={21} />}
        </button>

        {/* Navigation Tabs */}
        <div id="primary-navigation" className="navbar-links">
          <NavTab
            active={currentView === 'verify'}
            onClick={() => navigate('verify')}
            label="Verify Drug"
          />
          {isAuthenticated && (
            <>
              <NavTab
                active={currentView === 'dashboard'}
                onClick={() => navigate('dashboard')}
                label="Dashboard"
              />
              <NavTab
                active={currentView === 'drugs'}
                onClick={() => navigate('drugs')}
                label="Drugs"
              />
              <NavTab
                active={currentView === 'manufacturers'}
                onClick={() => navigate('manufacturers')}
                label="Manufacturers"
              />
              <NavTab
                active={currentView === 'audit'}
                onClick={() => navigate('audit')}
                label="Audit Logs"
              />
            </>
          )}
        </div>

        {/* Right Side: Auth */}
        <div className="navbar-auth">
          {isAuthenticated ? (
            <>
              {/* Online Indicator */}
              <div className="navbar-online">
                <span className="navbar-online-dot" />
                <span>ONLINE</span>
              </div>

              <div
                className="navbar-user"
                title={user?.full_name || user?.email}
              >
                <User size={14} color="var(--primary)" />
                <span className="navbar-user-label">
                  {user?.full_name?.split(' ')[0] || user?.email}
                </span>
              </div>

              <button onClick={() => { logout(); setMenuOpen(false); }} className="btn btn-ghost btn-sm navbar-logout" aria-label="Log out" title="Log out" style={{ color: 'var(--text-dim)' }}>
                <LogOut size={16} />
              </button>
            </>
          ) : (
            <button onClick={() => { setMenuOpen(false); onLoginClick(); }} className="btn btn-primary btn-sm">
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
      className={`navbar-tab${active ? ' is-active' : ''}`}
      aria-current={active ? 'page' : undefined}
    >
      {label}
    </button>
  );
}
