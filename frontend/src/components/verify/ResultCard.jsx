import React from 'react';
import {
  ShieldCheck,
  ShieldAlert,
  AlertTriangle,
  XCircle,
  Calendar,
  Building2,
  FileText,
  Hash,
  Printer,
  Share2,
  AlertOctagon,
} from 'lucide-react';
import Badge from '../common/Badge';

export default function ResultCard({ result, onReset }) {
  if (!result) return null;

  const { verified, result: status, message, drug, drug_unit } = result;

  const handlePrint = () => {
    window.print();
  };

  const handleShare = () => {
    if (navigator.share) {
      navigator.share({
        title: 'MedVerify Drug Verification Result',
        text: `MedVerify Result: ${drug ? drug.name : 'Unknown Product'} - ${status.toUpperCase()}`,
        url: window.location.href,
      }).catch(() => {});
    } else {
      navigator.clipboard.writeText(window.location.href);
      alert('Verification link copied to clipboard!');
    }
  };

  // Card themes based on verification outcome
  let bannerBg = 'rgba(16, 185, 129, 0.15)';
  let borderColor = 'var(--success)';
  let iconComponent = <ShieldCheck size={48} color="var(--success)" />;
  let headline = 'Genuine & Verified Authentic';
  let badgeStatus = 'verified';

  if (status === 'not_found' || status === 'invalid') {
    bannerBg = 'rgba(244, 63, 94, 0.18)';
    borderColor = 'var(--danger)';
    iconComponent = <ShieldAlert size={48} color="var(--danger)" />;
    headline = 'POTENTIAL COUNTERFEIT / UNVERIFIED';
    badgeStatus = 'recalled';
  } else if (status === 'expired') {
    bannerBg = 'rgba(245, 158, 11, 0.18)';
    borderColor = 'var(--warning)';
    iconComponent = <AlertTriangle size={48} color="var(--warning)" />;
    headline = 'EXPIRED PRODUCT — DO NOT USE';
    badgeStatus = 'expired';
  } else if (status === 'recalled' || status === 'suspended') {
    bannerBg = 'rgba(244, 63, 94, 0.2)';
    borderColor = 'var(--danger)';
    iconComponent = <AlertOctagon size={48} color="var(--danger)" />;
    headline = 'PRODUCT RECALLED BY MANUFACTURER';
    badgeStatus = 'recalled';
  }

  return (
    <div
      className="glass-panel"
      style={{
        border: `2px solid ${borderColor}`,
        borderRadius: 'var(--radius-lg)',
        overflow: 'hidden',
        marginTop: '2rem',
        boxShadow: `0 12px 35px -8px ${borderColor}33`,
      }}
    >
      {/* Top Banner */}
      <div
        style={{
          background: bannerBg,
          padding: '1.75rem',
          display: 'flex',
          alignItems: 'center',
          gap: '1.25rem',
          borderBottom: `1px solid ${borderColor}55`,
        }}
      >
        <div style={{ flexShrink: 0 }}>{iconComponent}</div>
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.35rem' }}>
            <Badge status={badgeStatus} label={status} />
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Verified at {new Date().toLocaleTimeString()}
            </span>
          </div>
          <h2 style={{ fontSize: '1.4rem', fontWeight: 800 }}>{headline}</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginTop: '0.25rem' }}>
            {message}
          </p>
        </div>
      </div>

      {/* Details Section */}
      <div style={{ padding: '1.75rem' }}>
        {drug ? (
          <div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.25rem', marginBottom: '1.5rem' }}>
              <div style={{ background: 'rgba(15, 23, 42, 0.5)', padding: '1rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Commercial Drug Name</span>
                <p style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--text-main)', marginTop: '0.25rem' }}>{drug.name}</p>
                <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>{drug.generic_name}</p>
              </div>

              <div style={{ background: 'rgba(15, 23, 42, 0.5)', padding: '1rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Manufacturer</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.35rem' }}>
                  <Building2 size={16} color="var(--primary)" />
                  <p style={{ fontWeight: 600, color: 'var(--text-main)' }}>{drug.manufacturer}</p>
                </div>
              </div>

              <div style={{ background: 'rgba(15, 23, 42, 0.5)', padding: '1rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Batch & Regulatory</span>
                <div style={{ marginTop: '0.35rem', display: 'flex', flexDirection: 'column', gap: '0.2rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Batch No:</span>
                    <strong style={{ color: 'var(--text-main)' }}>{drug.batch_number}</strong>
                  </div>
                  {drug.nafdac_number && (
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
                      <span style={{ color: 'var(--text-muted)' }}>NAFDAC Reg:</span>
                      <strong style={{ color: 'var(--primary)' }}>{drug.nafdac_number}</strong>
                    </div>
                  )}
                </div>
              </div>

              <div style={{ background: 'rgba(15, 23, 42, 0.5)', padding: '1rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-subtle)' }}>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Shelf Life Timeline</span>
                <div style={{ marginTop: '0.35rem', display: 'flex', flexDirection: 'column', gap: '0.2rem' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Mfg Date:</span>
                    <span>{drug.manufacturing_date}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem' }}>
                    <span style={{ color: 'var(--text-muted)' }}>Expiry Date:</span>
                    <strong style={{ color: status === 'expired' ? 'var(--warning)' : 'var(--text-main)' }}>
                      {drug.expiry_date}
                    </strong>
                  </div>
                </div>
              </div>
            </div>

            {drug_unit && (
              <div
                style={{
                  background: 'rgba(14, 165, 233, 0.08)',
                  border: '1px dashed var(--primary)',
                  borderRadius: 'var(--radius-md)',
                  padding: '0.85rem 1.25rem',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  marginBottom: '1.5rem',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
                  <Hash size={18} color="var(--primary)" />
                  <div>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Individual Serial Identifier</span>
                    <p style={{ fontFamily: 'monospace', fontSize: '1rem', fontWeight: 700, letterSpacing: '0.08em' }}>
                      {drug_unit.serial_number}
                    </p>
                  </div>
                </div>
                <Badge status={drug_unit.status} />
              </div>
            )}
          </div>
        ) : (
          <div
            style={{
              padding: '1.5rem',
              background: 'rgba(244, 63, 94, 0.08)',
              border: '1px solid rgba(244, 63, 94, 0.25)',
              borderRadius: 'var(--radius-md)',
              marginBottom: '1.5rem',
            }}
          >
            <h4 style={{ color: 'var(--danger)', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <ShieldAlert size={18} /> Consumer Safety Advisory
            </h4>
            <p style={{ fontSize: '0.9rem', color: 'var(--text-muted)', lineHeight: '1.6' }}>
              The serial code entered does not match any authentic registered pharmaceutical record in the national database.
              Do not consume this product. Please report the pharmacy or vendor where you acquired this package immediately to regulatory authorities.
            </p>
          </div>
        )}

        {/* Action Controls */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem', paddingTop: '1rem', borderTop: '1px solid var(--border-subtle)' }}>
          <button onClick={onReset} className="btn btn-outline btn-sm">
            Verify Another Medicine
          </button>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button onClick={handleShare} className="btn btn-outline btn-sm">
              <Share2 size={15} /> Share Proof
            </button>
            <button onClick={handlePrint} className="btn btn-primary btn-sm">
              <Printer size={15} /> Print Certificate
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
