import React, { useState, useEffect } from 'react';
import { Pill, Building2, QrCode, FileSearch, TrendingUp, AlertTriangle, ShieldCheck, Loader2 } from 'lucide-react';
import { api } from '../../api/client';

function StatCard({ icon, label, value, color, bgColor }) {
  return (
    <div
      className="glass-panel"
      style={{
        padding: '1.25rem',
        display: 'flex',
        alignItems: 'center',
        gap: '1rem',
        transition: 'border-color var(--transition-fast), transform var(--transition-fast)',
      }}
    >
      <div
        style={{
          width: '48px', height: '48px', borderRadius: 'var(--radius-md)',
          background: bgColor || 'rgba(14, 165, 233, 0.12)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
        }}
      >
        {icon}
      </div>
      <div>
        <div style={{ fontSize: '1.6rem', fontWeight: 800, color: color || 'var(--text-main)', fontFamily: 'var(--font-display)' }}>
          {value !== null ? value : <Loader2 size={18} className="spin" />}
        </div>
        <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500 }}>{label}</div>
      </div>
    </div>
  );
}

export default function DashboardOverview() {
  const [stats, setStats] = useState({
    drugs: null,
    manufacturers: null,
    verifications: null,
    counterfeits: null,
  });

  useEffect(() => {
    async function loadStats() {
      try {
        const [drugsRes, mfgRes, logsRes] = await Promise.allSettled([
          api.drugs.list({ page: 1, limit: 1 }),
          api.manufacturers.list({ page: 1, limit: 1 }),
          api.logs.list({ page: 1, limit: 1 }),
        ]);

        const newStats = { ...stats };

        if (drugsRes.status === 'fulfilled' && drugsRes.value.pagination) {
          newStats.drugs = drugsRes.value.pagination.total;
        } else {
          newStats.drugs = 0;
        }

        if (mfgRes.status === 'fulfilled' && mfgRes.value.pagination) {
          newStats.manufacturers = mfgRes.value.pagination.total;
        } else if (mfgRes.status === 'fulfilled' && mfgRes.value.data) {
          newStats.manufacturers = mfgRes.value.data.length;
        } else {
          newStats.manufacturers = 0;
        }

        if (logsRes.status === 'fulfilled' && logsRes.value.pagination) {
          newStats.verifications = logsRes.value.pagination.total;
        } else {
          newStats.verifications = 0;
        }

        // Fetch counterfeit attempts
        try {
          const cfRes = await api.logs.list({ page: 1, limit: 1, result: 'not_found' });
          newStats.counterfeits = cfRes.pagination ? cfRes.pagination.total : 0;
        } catch {
          newStats.counterfeits = 0;
        }

        setStats(newStats);
      } catch (err) {
        console.error('Failed to load dashboard stats:', err);
      }
    }

    loadStats();
  }, []);

  return (
    <div>
      {/* Stat Cards */}
      <div className="stats-grid">
        <StatCard
          icon={<Pill size={22} color="var(--primary)" />}
          label="Registered Drug Batches"
          value={stats.drugs}
          color="var(--primary)"
          bgColor="rgba(14, 165, 233, 0.12)"
        />
        <StatCard
          icon={<Building2 size={22} color="var(--success)" />}
          label="Manufacturers"
          value={stats.manufacturers}
          color="var(--success)"
          bgColor="var(--success-bg)"
        />
        <StatCard
          icon={<ShieldCheck size={22} color="#818cf8" />}
          label="Total Verification Scans"
          value={stats.verifications}
          color="#818cf8"
          bgColor="var(--info-bg)"
        />
        <StatCard
          icon={<AlertTriangle size={22} color="var(--danger)" />}
          label="Counterfeit Alerts"
          value={stats.counterfeits}
          color="var(--danger)"
          bgColor="var(--danger-bg)"
        />
      </div>

      {/* Quick Actions */}
      <div className="glass-panel" style={{ padding: '1.5rem' }}>
        <h3 style={{ fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <TrendingUp size={18} color="var(--primary)" /> Quick Start Guide
        </h3>
        <div className="quick-start-grid">
          <div style={{ padding: '1rem', background: 'rgba(15, 23, 42, 0.5)', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
              <span style={{ background: 'var(--primary)', color: '#fff', width: '24px', height: '24px', borderRadius: '50%', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.8rem', fontWeight: 700 }}>1</span>
              <strong style={{ fontSize: '0.95rem' }}>Add Manufacturer</strong>
            </div>
            <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)', lineHeight: 1.5 }}>
              Register the pharmaceutical company that produces the drug in the Manufacturers tab.
            </p>
          </div>

          <div style={{ padding: '1rem', background: 'rgba(15, 23, 42, 0.5)', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
              <span style={{ background: 'var(--success)', color: '#fff', width: '24px', height: '24px', borderRadius: '50%', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.8rem', fontWeight: 700 }}>2</span>
              <strong style={{ fontSize: '0.95rem' }}>Register Drug Batch</strong>
            </div>
            <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)', lineHeight: 1.5 }}>
              Add the drug with batch number, NAFDAC reg, manufacturing & expiry dates in the Drugs tab.
            </p>
          </div>

          <div style={{ padding: '1rem', background: 'rgba(15, 23, 42, 0.5)', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
              <span style={{ background: 'var(--warning)', color: '#fff', width: '24px', height: '24px', borderRadius: '50%', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', fontSize: '0.8rem', fontWeight: 700 }}>3</span>
              <strong style={{ fontSize: '0.95rem' }}>Serialize & Generate QR</strong>
            </div>
            <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)', lineHeight: 1.5 }}>
              Click the QR icon on any drug batch to generate unique serialized units with printable QR labels.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
