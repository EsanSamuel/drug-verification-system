import React, { useState, useEffect, useCallback } from 'react';
import {
  Plus, Search, RefreshCw, Pill, QrCode, Trash2, Edit3, Eye,
  ChevronLeft, ChevronRight, Filter, Loader2, AlertTriangle,
} from 'lucide-react';
import { api } from '../../api/client';
import Badge from '../common/Badge';
import DrugModal from './DrugModal';
import SerializationModal from './SerializationModal';

export default function DrugsTable() {
  const [drugs, setDrugs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState({ page: 1, limit: 15, total: 0, total_pages: 0 });
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const [drugModalOpen, setDrugModalOpen] = useState(false);
  const [editDrug, setEditDrug] = useState(null);
  const [serializeDrug, setSerializeDrug] = useState(null);
  const [deleteConfirm, setDeleteConfirm] = useState(null);
  const [deleting, setDeleting] = useState(false);

  const fetchDrugs = useCallback(async (page = 1) => {
    setLoading(true);
    try {
      const params = { page, limit: pagination.limit };
      if (search) params.name = search;
      if (statusFilter) params.status = statusFilter;
      const res = await api.drugs.list(params);
      if (res.success) {
        setDrugs(res.data || []);
        if (res.pagination) setPagination(res.pagination);
      }
    } catch (err) {
      console.error('Failed to load drugs:', err);
    } finally {
      setLoading(false);
    }
  }, [search, statusFilter, pagination.limit]);

  useEffect(() => {
    fetchDrugs(1);
  }, [search, statusFilter]);

  const handleDelete = async (id) => {
    setDeleting(true);
    try {
      await api.drugs.delete(id);
      setDeleteConfirm(null);
      fetchDrugs(pagination.page);
    } catch (err) {
      alert(err.message || 'Delete failed');
    } finally {
      setDeleting(false);
    }
  };

  const formatDate = (d) => {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  };

  return (
    <div>
      {/* Toolbar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem', marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', gap: '0.75rem', flex: 1, minWidth: '280px' }}>
          <div style={{ position: 'relative', flex: 1, maxWidth: '320px' }}>
            <Search size={16} color="var(--text-dim)" style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)' }} />
            <input
              type="text"
              className="form-control"
              placeholder="Search by drug name..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              style={{ paddingLeft: '2.5rem', height: '40px', fontSize: '0.85rem' }}
            />
          </div>
          <select
            className="form-control"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{ width: '150px', height: '40px', fontSize: '0.85rem' }}
          >
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="expired">Expired</option>
            <option value="recalled">Recalled</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button onClick={() => fetchDrugs(pagination.page)} className="btn btn-ghost btn-sm">
            <RefreshCw size={15} />
          </button>
          <button
            onClick={() => { setEditDrug(null); setDrugModalOpen(true); }}
            className="btn btn-primary btn-sm"
          >
            <Plus size={16} /> Add Drug Batch
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="glass-panel" style={{ overflow: 'hidden' }}>
        <div className="table-responsive">
          <table className="data-table">
            <thead>
              <tr>
                <th>Drug Name</th>
                <th>Batch #</th>
                <th>NAFDAC</th>
                <th>Mfg Date</th>
                <th>Expiry</th>
                <th>Qty</th>
                <th>Status</th>
                <th style={{ textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan="8" style={{ textAlign: 'center', padding: '3rem' }}>
                    <Loader2 size={24} className="spin" color="var(--primary)" />
                    <p style={{ color: 'var(--text-muted)', marginTop: '0.5rem', fontSize: '0.85rem' }}>Loading drug batches...</p>
                  </td>
                </tr>
              ) : drugs.length === 0 ? (
                <tr>
                  <td colSpan="8" style={{ textAlign: 'center', padding: '3rem' }}>
                    <Pill size={32} color="var(--text-dim)" style={{ marginBottom: '0.5rem' }} />
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>No drug batches found.</p>
                    <button
                      onClick={() => { setEditDrug(null); setDrugModalOpen(true); }}
                      className="btn btn-primary btn-sm"
                      style={{ marginTop: '0.75rem' }}
                    >
                      <Plus size={15} /> Register First Batch
                    </button>
                  </td>
                </tr>
              ) : (
                drugs.map((drug) => (
                  <tr key={drug.id}>
                    <td>
                      <div>
                        <span style={{ fontWeight: 600 }}>{drug.name}</span>
                        {drug.generic_name && (
                          <div style={{ fontSize: '0.775rem', color: 'var(--text-muted)' }}>
                            {drug.generic_name}
                          </div>
                        )}
                      </div>
                    </td>
                    <td>
                      <code style={{ fontSize: '0.825rem', color: 'var(--primary)' }}>
                        {drug.batch_number}
                      </code>
                    </td>
                    <td style={{ fontSize: '0.85rem' }}>{drug.nafdac_number || '—'}</td>
                    <td style={{ fontSize: '0.85rem' }}>{formatDate(drug.manufacturing_date)}</td>
                    <td style={{ fontSize: '0.85rem' }}>{formatDate(drug.expiry_date)}</td>
                    <td style={{ fontSize: '0.85rem', fontWeight: 600 }}>{drug.quantity}</td>
                    <td><Badge status={drug.status} /></td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.35rem', justifyContent: 'flex-end' }}>
                        <button
                          onClick={() => setSerializeDrug(drug)}
                          className="btn btn-ghost btn-sm"
                          title="Generate Serialized Units & QR Codes"
                          style={{ padding: '0.3rem 0.5rem' }}
                        >
                          <QrCode size={15} color="var(--success)" />
                        </button>
                        <button
                          onClick={() => { setEditDrug(drug); setDrugModalOpen(true); }}
                          className="btn btn-ghost btn-sm"
                          title="Edit Drug Batch"
                          style={{ padding: '0.3rem 0.5rem' }}
                        >
                          <Edit3 size={15} />
                        </button>
                        <button
                          onClick={() => setDeleteConfirm(drug)}
                          className="btn btn-ghost btn-sm"
                          title="Delete Drug Batch"
                          style={{ padding: '0.3rem 0.5rem' }}
                        >
                          <Trash2 size={15} color="var(--danger)" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {pagination.total_pages > 1 && (
          <div style={{
            display: 'flex', justifyContent: 'space-between', alignItems: 'center',
            padding: '0.85rem 1rem', borderTop: '1px solid var(--border-subtle)',
          }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
              Page {pagination.page} of {pagination.total_pages} ({pagination.total} total records)
            </span>
            <div style={{ display: 'flex', gap: '0.4rem' }}>
              <button
                disabled={pagination.page <= 1}
                onClick={() => fetchDrugs(pagination.page - 1)}
                className="btn btn-ghost btn-sm"
              >
                <ChevronLeft size={16} />
              </button>
              <button
                disabled={pagination.page >= pagination.total_pages}
                onClick={() => fetchDrugs(pagination.page + 1)}
                className="btn btn-ghost btn-sm"
              >
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Delete Confirmation */}
      {deleteConfirm && (
        <div className="modal-overlay" onClick={() => setDeleteConfirm(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '420px' }}>
            <div className="modal-header">
              <h3 style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '1.1rem' }}>
                <AlertTriangle size={20} color="var(--danger)" /> Confirm Deletion
              </h3>
            </div>
            <div className="modal-body">
              <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
                Are you sure you want to delete <strong style={{ color: 'var(--text-main)' }}>{deleteConfirm.name}</strong> (Batch: {deleteConfirm.batch_number})?
                This action cannot be undone.
              </p>
            </div>
            <div className="modal-footer">
              <button onClick={() => setDeleteConfirm(null)} className="btn btn-outline">Cancel</button>
              <button
                onClick={() => handleDelete(deleteConfirm.id)}
                disabled={deleting}
                className="btn btn-danger"
              >
                {deleting ? <Loader2 size={16} className="spin" /> : <Trash2 size={16} />}
                <span>Delete</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Drug Modal */}
      <DrugModal
        isOpen={drugModalOpen}
        onClose={() => { setDrugModalOpen(false); setEditDrug(null); }}
        onSuccess={() => fetchDrugs(pagination.page)}
        initialData={editDrug}
      />

      {/* Serialization Modal */}
      <SerializationModal
        isOpen={!!serializeDrug}
        onClose={() => setSerializeDrug(null)}
        drug={serializeDrug}
      />
    </div>
  );
}
