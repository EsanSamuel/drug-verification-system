import React, { useState, useEffect, useCallback } from 'react';
import { Building2, Plus, RefreshCw, Search, Loader2, X, MapPin } from 'lucide-react';
import { api } from '../../api/client';

export default function ManufacturersView() {
  const [manufacturers, setManufacturers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [name, setName] = useState('');
  const [address, setAddress] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState(null);
  const [search, setSearch] = useState('');

  const fetchManufacturers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.manufacturers.list({ limit: 100 });
      if (res.success) {
        setManufacturers(res.data || []);
      }
    } catch (err) {
      console.error('Failed to fetch manufacturers:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchManufacturers();
  }, [fetchManufacturers]);

  const handleCreate = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      await api.manufacturers.create({ name, address });
      setShowModal(false);
      setName('');
      setAddress('');
      fetchManufacturers();
    } catch (err) {
      setError(err.message || 'Failed to create manufacturer');
    } finally {
      setSaving(false);
    }
  };

  const filtered = manufacturers.filter(
    (m) => m.name.toLowerCase().includes(search.toLowerCase()) ||
           m.address.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div>
      {/* Toolbar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
        <div style={{ position: 'relative', flex: 1, maxWidth: '320px' }}>
          <Search size={16} color="var(--text-dim)" style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)' }} />
          <input
            type="text"
            className="form-control"
            placeholder="Search manufacturers..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{ paddingLeft: '2.5rem', height: '40px', fontSize: '0.85rem' }}
          />
        </div>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button onClick={fetchManufacturers} className="btn btn-ghost btn-sm">
            <RefreshCw size={15} />
          </button>
          <button onClick={() => setShowModal(true)} className="btn btn-primary btn-sm">
            <Plus size={16} /> Add Manufacturer
          </button>
        </div>
      </div>

      {/* Cards Grid */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem 0' }}>
          <Loader2 size={28} className="spin" color="var(--primary)" />
          <p style={{ color: 'var(--text-muted)', marginTop: '0.5rem', fontSize: '0.85rem' }}>Loading manufacturers...</p>
        </div>
      ) : filtered.length === 0 ? (
        <div className="glass-panel" style={{ textAlign: 'center', padding: '3rem' }}>
          <Building2 size={36} color="var(--text-dim)" style={{ marginBottom: '0.5rem' }} />
          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
            {search ? 'No manufacturers match your search.' : 'No manufacturers registered yet.'}
          </p>
          <button onClick={() => setShowModal(true)} className="btn btn-primary btn-sm" style={{ marginTop: '0.75rem' }}>
            <Plus size={15} /> Register First Manufacturer
          </button>
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '1rem' }}>
          {filtered.map((m) => (
            <div
              key={m.id}
              className="glass-panel"
              style={{
                padding: '1.25rem',
                transition: 'border-color var(--transition-fast)',
                cursor: 'default',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.75rem' }}>
                <div
                  style={{
                    width: '40px', height: '40px', borderRadius: 'var(--radius-md)',
                    background: 'rgba(14, 165, 233, 0.12)', display: 'flex',
                    alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                  }}
                >
                  <Building2 size={20} color="var(--primary)" />
                </div>
                <div style={{ flex: 1 }}>
                  <h4 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.3rem' }}>{m.name}</h4>
                  {m.address && (
                    <div style={{ display: 'flex', alignItems: 'flex-start', gap: '0.35rem' }}>
                      <MapPin size={13} color="var(--text-dim)" style={{ marginTop: '2px', flexShrink: 0 }} />
                      <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', lineHeight: 1.4 }}>
                        {m.address}
                      </span>
                    </div>
                  )}
                  <div style={{ fontSize: '0.7rem', color: 'var(--text-dim)', marginTop: '0.5rem' }}>
                    Added {new Date(m.created_at).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Manufacturer Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '460px' }}>
            <div className="modal-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Building2 size={20} color="var(--primary)" />
                <h3 style={{ fontSize: '1.15rem' }}>Register Manufacturer</h3>
              </div>
              <button onClick={() => setShowModal(false)} className="btn btn-ghost btn-sm" style={{ padding: '0.3rem', borderRadius: '50%' }}>
                <X size={18} />
              </button>
            </div>
            <form onSubmit={handleCreate}>
              <div className="modal-body">
                {error && (
                  <div style={{ padding: '0.75rem', marginBottom: '1rem', background: 'var(--danger-bg)', border: '1px solid rgba(244, 63, 94, 0.3)', borderRadius: 'var(--radius-md)', color: 'var(--danger)', fontSize: '0.85rem' }}>
                    {error}
                  </div>
                )}
                <div className="form-group">
                  <label className="form-label">Company Name *</label>
                  <input
                    type="text"
                    required
                    className="form-control"
                    placeholder="e.g. GlaxoSmithKline Pharmaceuticals"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Address</label>
                  <input
                    type="text"
                    className="form-control"
                    placeholder="e.g. 12 Industrial Layout, Lagos, Nigeria"
                    value={address}
                    onChange={(e) => setAddress(e.target.value)}
                  />
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" onClick={() => setShowModal(false)} className="btn btn-outline">Cancel</button>
                <button type="submit" disabled={saving} className="btn btn-primary">
                  {saving && <Loader2 size={16} className="spin" />}
                  <span>Register</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
