import React, { useState } from 'react';
import Navbar from './components/common/Navbar';
import VerifyPortal from './components/verify/VerifyPortal';
import LoginModal from './components/auth/LoginModal';
import DashboardOverview from './components/dashboard/DashboardOverview';
import DrugsTable from './components/dashboard/DrugsTable';
import ManufacturersView from './components/dashboard/ManufacturersView';
import AuditLogsView from './components/dashboard/AuditLogsView';
import { useAuth } from './context/AuthContext';

const VIEW_TITLES = {
  verify: 'Drug Verification',
  dashboard: 'Operations Dashboard',
  drugs: 'Drug Batch Registry',
  manufacturers: 'Pharmaceutical Manufacturers',
  audit: 'Verification Audit Trail',
};

export default function App() {
  const { isAuthenticated, loading } = useAuth();
  const [currentView, setCurrentView] = useState('verify');
  const [loginOpen, setLoginOpen] = useState(false);

  const navigate = (view) => {
    // If navigating to a protected view, prompt login first
    if (!isAuthenticated && view !== 'verify') {
      setLoginOpen(true);
      return;
    }
    setCurrentView(view);
  };

  if (loading) {
    return (
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        height: '100vh', flexDirection: 'column', gap: '1rem',
      }}>
        <div style={{
          width: '48px', height: '48px', borderRadius: '50%',
          border: '3px solid var(--border-medium)', borderTopColor: 'var(--primary)',
          animation: 'spin 0.8s linear infinite',
        }} />
        <span style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>Loading MedVerify...</span>
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  const renderView = () => {
    switch (currentView) {
      case 'dashboard':
        return <DashboardOverview />;
      case 'drugs':
        return <DrugsTable />;
      case 'manufacturers':
        return <ManufacturersView />;
      case 'audit':
        return <AuditLogsView />;
      case 'verify':
      default:
        return <VerifyPortal />;
    }
  };

  return (
    <div className="app-container">
      <Navbar
        currentView={currentView}
        onNavigate={navigate}
        onLoginClick={() => setLoginOpen(true)}
      />

      <main className="main-content">
        {/* Page Header (for authenticated sections) */}
        {currentView !== 'verify' && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h1 style={{ fontSize: '1.65rem', fontWeight: 800 }}>
              {VIEW_TITLES[currentView] || 'Dashboard'}
            </h1>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginTop: '0.25rem' }}>
              {currentView === 'dashboard' && 'System overview and quick start actions.'}
              {currentView === 'drugs' && 'Manage registered pharmaceutical drug batches, generate serialized units and QR codes.'}
              {currentView === 'manufacturers' && 'View and register pharmaceutical production companies.'}
              {currentView === 'audit' && 'Monitor all public verification scans and detect counterfeit patterns.'}
            </p>
          </div>
        )}

        {renderView()}
      </main>

      {/* Footer */}
      <footer style={{
        textAlign: 'center', padding: '1.5rem',
        borderTop: '1px solid var(--border-subtle)',
        fontSize: '0.75rem', color: 'var(--text-dim)',
      }}>
        <span>MedVerify — Computerized Drug Verification & Management System</span>
        <span style={{ margin: '0 0.5rem' }}>·</span>
        <span>Final Year Project © {new Date().getFullYear()}</span>
      </footer>

      {/* Login Modal */}
      <LoginModal isOpen={loginOpen} onClose={() => setLoginOpen(false)} />
    </div>
  );
}
