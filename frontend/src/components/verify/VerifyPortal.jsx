import React, { useState, useEffect } from 'react';
import { Search, Camera, ShieldCheck, Loader2, Sparkles, CheckCircle2, AlertTriangle } from 'lucide-react';
import { api } from '../../api/client';
import QRScanner from './QRScanner';
import ResultCard from './ResultCard';

export default function VerifyPortal() {
  const [serial, setSerial] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);
  const [showScanner, setShowScanner] = useState(false);

  useEffect(() => {
    // Check if URL has ?serial=... or /verify?serial=...
    const params = new URLSearchParams(window.location.search);
    const initialSerial = params.get('serial');
    if (initialSerial) {
      setSerial(initialSerial);
      performVerification(initialSerial);
    }
  }, []);

  const performVerification = async (serialToVerify) => {
    const target = (serialToVerify || serial).trim();
    if (!target) {
      setError('Please provide or scan a serial number.');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const res = await api.verify.check(target);
      if (res.success && res.data) {
        setResult(res.data);
      } else {
        setError(res.error || 'Verification query failed.');
      }
    } catch (err) {
      setError(err.message || 'Could not reach verification server.');
    } finally {
      setLoading(false);
    }
  };

  const handleScanSuccess = (scannedSerial) => {
    setShowScanner(false);
    setSerial(scannedSerial);
    performVerification(scannedSerial);
  };

  const handleReset = () => {
    setResult(null);
    setSerial('');
    setError(null);
  };

  return (
    <div className="verify-page">
      {/* Hero Section */}
      <div className="verify-hero">
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.4rem 1rem',
            background: 'rgba(14, 165, 233, 0.12)',
            borderRadius: 'var(--radius-full)',
            border: '1px solid rgba(14, 165, 233, 0.3)',
            marginBottom: '1rem',
          }}
        >
          <Sparkles size={16} color="var(--primary)" />
          <span style={{ fontSize: '0.825rem', fontWeight: 600, color: 'var(--primary)', letterSpacing: '0.04em' }}>
            NATIONAL PHARMACEUTICAL VERIFICATION NETWORK
          </span>
        </div>

        <h1 className="verify-title">
          Instant Drug Authenticity Verification
        </h1>
        <p className="verify-description">
          Scan the QR code on your medicine pack or enter the unit serial number below to verify against registered pharmaceutical batches in real-time.
        </p>
      </div>

      {/* Verification Input Box */}
      <div
        className="glass-panel verify-panel"
        style={{
          boxShadow: 'var(--shadow-lg)',
          position: 'relative',
        }}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            performVerification();
          }}
          className="verify-form"
        >
          <div className="verify-input-wrap">
            <Search
              size={18}
              color="var(--text-dim)"
              style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)' }}
            />
            <input
              type="text"
              className="form-control"
              placeholder="e.g. 4B892F1AC9083DE1 or Batch Serial"
              value={serial}
              onChange={(e) => setSerial(e.target.value)}
              style={{
                paddingLeft: '2.75rem',
                height: '50px',
                fontSize: '1rem',
                fontFamily: 'monospace',
                letterSpacing: '0.05em',
              }}
            />
          </div>

          <button
            type="button"
            onClick={() => setShowScanner(!showScanner)}
            className={`btn verify-action ${showScanner ? 'btn-danger' : 'btn-outline'}`}
            style={{ height: '50px', padding: '0 1.25rem' }}
          >
            <Camera size={18} />
            <span>{showScanner ? 'Close Scanner' : 'Scan QR'}</span>
          </button>

          <button
            type="submit"
            disabled={loading}
            className="btn btn-primary verify-action"
            style={{ height: '50px', padding: '0 1.75rem', minWidth: '130px' }}
          >
            {loading ? (
              <>
                <Loader2 size={18} className="spin" />
                <span>Verifying...</span>
              </>
            ) : (
              <>
                <ShieldCheck size={18} />
                <span>Verify Now</span>
              </>
            )}
          </button>
        </form>

        {/* Live Camera Scanner Box */}
        {showScanner && (
          <QRScanner
            onScan={handleScanSuccess}
            onClose={() => setShowScanner(false)}
          />
        )}

        {/* Error notification */}
        {error && (
          <div
            style={{
              marginTop: '1.25rem',
              padding: '0.85rem 1.25rem',
              background: 'var(--danger-bg)',
              border: '1px solid rgba(244, 63, 94, 0.3)',
              borderRadius: 'var(--radius-md)',
              color: 'var(--danger)',
              fontSize: '0.9rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
            }}
          >
            <AlertTriangle size={18} />
            <span>{error}</span>
          </div>
        )}
      </div>

      {/* Result Card */}
      {result && <ResultCard result={result} onReset={handleReset} />}

      {/* Trust Badges */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: '1.25rem',
          marginTop: '3.5rem',
          textAlign: 'center',
        }}
      >
        <div style={{ padding: '1rem' }}>
          <CheckCircle2 size={28} color="var(--success)" style={{ margin: '0 auto 0.5rem' }} />
          <h4 style={{ fontSize: '0.95rem', marginBottom: '0.25rem' }}>100% Tamper Proof</h4>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Cryptographically unique 16-hex unit identifiers generated at production.
          </p>
        </div>

        <div style={{ padding: '1rem' }}>
          <ShieldCheck size={28} color="var(--primary)" style={{ margin: '0 auto 0.5rem' }} />
          <h4 style={{ fontSize: '0.95rem', marginBottom: '0.25rem' }}>NAFDAC & Batch Tracking</h4>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Direct manufacturer association with verified expiry and recall safeguards.
          </p>
        </div>

        <div style={{ padding: '1rem' }}>
          <Sparkles size={28} color="var(--warning)" style={{ margin: '0 auto 0.5rem' }} />
          <h4 style={{ fontSize: '0.95rem', marginBottom: '0.25rem' }}>Audit Trail Logging</h4>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
            Every public verification is logged to detect counterfeit duplication clusters.
          </p>
        </div>
      </div>
    </div>
  );
}
