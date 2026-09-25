import React, { useState } from 'react';
import { X, QrCode, Loader2, Printer, Download, Copy, Check } from 'lucide-react';
import { api } from '../../api/client';

export default function SerializationModal({ isOpen, onClose, drug }) {
  const [quantity, setQuantity] = useState(10);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [generatedUnits, setGeneratedUnits] = useState([]);
  const [copiedIdx, setCopiedIdx] = useState(null);

  if (!isOpen || !drug) return null;

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.drugs.generateUnits(drug.id, quantity);
      if (res.success && res.data) {
        setGeneratedUnits(res.data);
      } else {
        setError('Failed to generate units.');
      }
    } catch (err) {
      setError(err.message || 'Generation failed');
    } finally {
      setLoading(false);
    }
  };

  const copySerial = (serial, idx) => {
    navigator.clipboard.writeText(serial);
    setCopiedIdx(idx);
    setTimeout(() => setCopiedIdx(null), 1500);
  };

  const handlePrintLabels = () => {
    const printWindow = window.open('', '_blank');
    const rows = generatedUnits.map(
      (u) => `
      <div style="display:inline-block;border:1px solid #ccc;padding:12px;margin:8px;text-align:center;width:200px;">
        <img src="${api.drugs.getUnitQRUrl(u.id, 180)}" alt="QR" style="width:160px;height:160px;" />
        <div style="margin-top:6px;font-family:monospace;font-weight:bold;font-size:12px;letter-spacing:0.05em;">${u.serial_number}</div>
        <div style="font-size:10px;color:#666;margin-top:2px;">${drug.name} — ${drug.batch_number}</div>
      </div>`
    );
    printWindow.document.write(`
      <html>
        <head><title>QR Labels — ${drug.name}</title></head>
        <body style="font-family:Arial,sans-serif;padding:20px;">
          <h3>${drug.name} — Batch ${drug.batch_number}</h3>
          <p>Generated ${generatedUnits.length} serialized unit labels</p>
          <div>${rows.join('')}</div>
          <script>window.onload = function() { window.print(); }</script>
        </body>
      </html>
    `);
    printWindow.document.close();
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '680px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <QrCode size={20} color="var(--primary)" />
            <div>
              <h3 style={{ fontSize: '1.15rem' }}>Unit Serialization & QR Generation</h3>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                {drug.name} — Batch {drug.batch_number}
              </span>
            </div>
          </div>
          <button onClick={onClose} className="btn btn-ghost btn-sm" style={{ padding: '0.3rem', borderRadius: '50%' }}>
            <X size={18} />
          </button>
        </div>

        <div className="modal-body">
          {error && (
            <div style={{ padding: '0.75rem', marginBottom: '1rem', background: 'var(--danger-bg)', border: '1px solid rgba(244, 63, 94, 0.3)', borderRadius: 'var(--radius-md)', color: 'var(--danger)', fontSize: '0.85rem' }}>
              {error}
            </div>
          )}

          {generatedUnits.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2rem 0' }}>
              <QrCode size={48} color="var(--primary)" style={{ margin: '0 auto 1rem', opacity: 0.6 }} />
              <h4 style={{ marginBottom: '0.5rem' }}>Generate Serialized Drug Units</h4>
              <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginBottom: '1.5rem', maxWidth: '420px', margin: '0 auto 1.5rem' }}>
                Each unit receives a cryptographically unique 16-character hex serial number and an associated QR code for consumer verification.
              </p>

              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '1rem' }}>
                <div className="form-group" style={{ margin: 0, width: '120px' }}>
                  <label className="form-label" style={{ textAlign: 'center' }}>Quantity</label>
                  <input
                    type="number"
                    min="1"
                    max="500"
                    className="form-control"
                    value={quantity}
                    onChange={(e) => setQuantity(e.target.value)}
                    style={{ textAlign: 'center' }}
                  />
                </div>
                <button
                  onClick={handleGenerate}
                  disabled={loading}
                  className="btn btn-success btn-lg"
                  style={{ marginTop: '1.1rem' }}
                >
                  {loading ? <Loader2 size={18} className="spin" /> : <QrCode size={18} />}
                  <span>{loading ? 'Generating...' : 'Generate Units'}</span>
                </button>
              </div>
            </div>
          ) : (
            <div>
              <div style={{
                display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem',
                padding: '0.75rem 1rem', background: 'var(--success-bg)', border: '1px solid rgba(16,185,129,0.3)',
                borderRadius: 'var(--radius-md)'
              }}>
                <span style={{ color: 'var(--success)', fontWeight: 600, fontSize: '0.9rem' }}>
                  ✓ {generatedUnits.length} serialized units generated successfully
                </span>
                <button onClick={handlePrintLabels} className="btn btn-outline btn-sm">
                  <Printer size={15} /> Print QR Labels
                </button>
              </div>

              <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
                <table className="data-table" style={{ width: '100%' }}>
                  <thead>
                    <tr>
                      <th>#</th>
                      <th>Serial Number</th>
                      <th>QR Code</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {generatedUnits.map((unit, idx) => (
                      <tr key={unit.id}>
                        <td style={{ color: 'var(--text-dim)' }}>{idx + 1}</td>
                        <td>
                          <code style={{
                            fontFamily: 'monospace', fontSize: '0.85rem', fontWeight: 600,
                            letterSpacing: '0.06em', color: 'var(--primary)'
                          }}>
                            {unit.serial_number}
                          </code>
                        </td>
                        <td>
                          <img
                            src={api.drugs.getUnitQRUrl(unit.id, 80)}
                            alt="QR"
                            style={{ width: '48px', height: '48px', borderRadius: '4px', border: '1px solid var(--border-subtle)' }}
                            loading="lazy"
                          />
                        </td>
                        <td>
                          <button
                            onClick={() => copySerial(unit.serial_number, idx)}
                            className="btn btn-ghost btn-sm"
                            style={{ padding: '0.25rem 0.5rem' }}
                          >
                            {copiedIdx === idx ? <Check size={14} color="var(--success)" /> : <Copy size={14} />}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
