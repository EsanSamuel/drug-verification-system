import React, { useState, useEffect } from 'react';
import { X, Pill, Loader2 } from 'lucide-react';
import { api } from '../../api/client';

export default function DrugModal({ isOpen, onClose, onSuccess, initialData = null }) {
  const [manufacturers, setManufacturers] = useState([]);
  const [name, setName] = useState('');
  const [genericName, setGenericName] = useState('');
  const [manufacturerId, setManufacturerId] = useState('');
  const [batchNumber, setBatchNumber] = useState('');
  const [nafdacNumber, setNafdacNumber] = useState('');
  const [mfgDate, setMfgDate] = useState('');
  const [expDate, setExpDate] = useState('');
  const [quantity, setQuantity] = useState(100);
  const [status, setStatus] = useState('active');

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (isOpen) {
      fetchManufacturers();
      if (initialData) {
        setName(initialData.name || '');
        setGenericName(initialData.generic_name || '');
        setManufacturerId(initialData.manufacturer_id || '');
        setBatchNumber(initialData.batch_number || '');
        setNafdacNumber(initialData.nafdac_number || '');
        setMfgDate(initialData.manufacturing_date ? initialData.manufacturing_date.split('T')[0] : '');
        setExpDate(initialData.expiry_date ? initialData.expiry_date.split('T')[0] : '');
        setQuantity(initialData.quantity || 0);
        setStatus(initialData.status || 'active');
      } else {
        resetForm();
      }
    }
  }, [isOpen, initialData]);

  const resetForm = () => {
    setName('');
    setGenericName('');
    setManufacturerId('');
    setBatchNumber('');
    setNafdacNumber('');
    const now = new Date();
    const future = new Date();
    future.setFullYear(now.getFullYear() + 2);
    setMfgDate(now.toISOString().split('T')[0]);
    setExpDate(future.toISOString().split('T')[0]);
    setQuantity(100);
    setStatus('active');
    setError(null);
  };

  const fetchManufacturers = async () => {
    try {
      const res = await api.manufacturers.list({ limit: 100 });
      if (res.success && res.data) {
        setManufacturers(res.data);
        if (res.data.length > 0 && !manufacturerId) {
          setManufacturerId(res.data[0].id);
        }
      }
    } catch (err) {
      console.error('Failed to load manufacturers:', err);
    }
  };

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    const payload = {
      name,
      generic_name: genericName,
      manufacturer_id: manufacturerId,
      batch_number: batchNumber,
      nafdac_number: nafdacNumber,
      manufacturing_date: mfgDate,
      expiry_date: expDate,
      quantity: Number(quantity),
      status,
    };

    try {
      if (initialData && initialData.id) {
        await api.drugs.update(initialData.id, payload);
      } else {
        await api.drugs.create(payload);
      }
      onSuccess();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to save drug batch');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '600px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <Pill size={20} color="var(--primary)" />
            <h3 style={{ fontSize: '1.2rem' }}>
              {initialData ? 'Update Drug Batch' : 'Register New Drug Batch'}
            </h3>
          </div>
          <button onClick={onClose} className="btn btn-ghost btn-sm" style={{ padding: '0.3rem', borderRadius: '50%' }}>
            <X size={18} />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="modal-body">
            {error && (
              <div style={{ padding: '0.75rem', marginBottom: '1rem', background: 'var(--danger-bg)', border: '1px solid rgba(244, 63, 94, 0.3)', borderRadius: 'var(--radius-md)', color: 'var(--danger)', fontSize: '0.85rem' }}>
                {error}
              </div>
            )}

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div className="form-group">
                <label className="form-label">Brand / Trade Name *</label>
                <input
                  type="text"
                  required
                  className="form-control"
                  placeholder="e.g. Augmentin 625mg"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Active / Generic Ingredient</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="e.g. Amoxicillin / Clavulanate"
                  value={genericName}
                  onChange={(e) => setGenericName(e.target.value)}
                />
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">Manufacturer *</label>
              <select
                required
                className="form-control"
                value={manufacturerId}
                onChange={(e) => setManufacturerId(e.target.value)}
              >
                {manufacturers.length === 0 && <option value="">No manufacturers registered</option>}
                {manufacturers.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name}
                  </option>
                ))}
              </select>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div className="form-group">
                <label className="form-label">Batch / Lot Number *</label>
                <input
                  type="text"
                  required
                  className="form-control"
                  placeholder="e.g. AUG-2026-09"
                  value={batchNumber}
                  onChange={(e) => setBatchNumber(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">NAFDAC Reg. Number</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="e.g. A4-0123"
                  value={nafdacNumber}
                  onChange={(e) => setNafdacNumber(e.target.value)}
                />
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div className="form-group">
                <label className="form-label">Manufacturing Date *</label>
                <input
                  type="date"
                  required
                  className="form-control"
                  value={mfgDate}
                  onChange={(e) => setMfgDate(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Expiry Date *</label>
                <input
                  type="date"
                  required
                  className="form-control"
                  value={expDate}
                  onChange={(e) => setExpDate(e.target.value)}
                />
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div className="form-group">
                <label className="form-label">Initial Quantity Units</label>
                <input
                  type="number"
                  min="0"
                  className="form-control"
                  value={quantity}
                  onChange={(e) => setQuantity(e.target.value)}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Batch Status</label>
                <select
                  className="form-control"
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  <option value="active">Active</option>
                  <option value="expired">Expired</option>
                  <option value="recalled">Recalled</option>
                  <option value="suspended">Suspended</option>
                </select>
              </div>
            </div>
          </div>

          <div className="modal-footer">
            <button type="button" onClick={onClose} className="btn btn-outline">
              Cancel
            </button>
            <button type="submit" disabled={loading} className="btn btn-primary">
              {loading && <Loader2 size={16} className="spin" />}
              <span>{initialData ? 'Save Changes' : 'Register Batch'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
