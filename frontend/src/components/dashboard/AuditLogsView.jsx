import React, { useState, useEffect, useCallback } from 'react';
import {
  Search, RefreshCw, ChevronLeft, ChevronRight, Loader2,
  FileSearch, Filter,
} from 'lucide-react';
import { api } from '../../api/client';
import Badge from '../common/Badge';

export default function AuditLogsView() {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState({ page: 1, limit: 20, total: 0, total_pages: 0 });
  const [serialSearch, setSerialSearch] = useState('');
  const [resultFilter, setResultFilter] = useState('');

  const fetchLogs = useCallback(async (page = 1) => {
    setLoading(true);
    try {
      const params = { page, limit: pagination.limit };
      if (serialSearch) params.serial_number = serialSearch;
      if (resultFilter) params.result = resultFilter;
      const res = await api.logs.list(params);
      if (res.success) {
        setLogs(res.data || []);
        if (res.pagination) setPagination(res.pagination);
      }
    } catch (err) {
      console.error('Failed to fetch audit logs:', err);
    } finally {
      setLoading(false);
    }
  }, [serialSearch, resultFilter, pagination.limit]);

  useEffect(() => {
    fetchLogs(1);
  }, [serialSearch, resultFilter]);

  const formatTime = (t) => {
    if (!t) return '—';
    const d = new Date(t);
    return d.toLocaleString('en-GB', {
      day: '2-digit', month: 'short', year: 'numeric',
      hour: '2-digit', minute: '2-digit', second: '2-digit',
    });
  };

  return (
    <div>
      {/* Toolbar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem', marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', gap: '0.75rem', flex: 1 }}>
          <div style={{ position: 'relative', flex: 1, maxWidth: '280px' }}>
            <Search size={16} color="var(--text-dim)" style={{ position: 'absolute', left: '0.85rem', top: '50%', transform: 'translateY(-50%)' }} />
            <input
              type="text"
              className="form-control"
              placeholder="Search serial number..."
              value={serialSearch}
              onChange={(e) => setSerialSearch(e.target.value)}
              style={{ paddingLeft: '2.5rem', height: '40px', fontSize: '0.85rem' }}
            />
          </div>
          <select
            className="form-control"
            value={resultFilter}
            onChange={(e) => setResultFilter(e.target.value)}
            style={{ width: '160px', height: '40px', fontSize: '0.85rem' }}
          >
            <option value="">All Results</option>
            <option value="verified">Verified</option>
            <option value="not_found">Not Found</option>
            <option value="expired">Expired</option>
            <option value="recalled">Recalled</option>
            <option value="suspended">Suspended</option>
            <option value="invalid">Invalid</option>
          </select>
        </div>
        <button onClick={() => fetchLogs(pagination.page)} className="btn btn-ghost btn-sm">
          <RefreshCw size={15} />
        </button>
      </div>

      {/* Log Table */}
      <div className="glass-panel" style={{ overflow: 'hidden' }}>
        <div className="table-responsive">
          <table className="data-table">
            <thead>
              <tr>
                <th>Timestamp</th>
                <th>Serial Number</th>
                <th>Verdict</th>
                <th>Authentic</th>
                <th>IP Address</th>
                <th>User Agent</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan="6" style={{ textAlign: 'center', padding: '3rem' }}>
                    <Loader2 size={24} className="spin" color="var(--primary)" />
                    <p style={{ color: 'var(--text-muted)', marginTop: '0.5rem', fontSize: '0.85rem' }}>Loading audit records...</p>
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan="6" style={{ textAlign: 'center', padding: '3rem' }}>
                    <FileSearch size={32} color="var(--text-dim)" style={{ marginBottom: '0.5rem' }} />
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>No verification records found.</p>
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id}>
                    <td style={{ fontSize: '0.825rem', whiteSpace: 'nowrap' }}>
                      {formatTime(log.verified_at)}
                    </td>
                    <td>
                      <code style={{ fontFamily: 'monospace', fontSize: '0.825rem', fontWeight: 600, letterSpacing: '0.04em', color: 'var(--primary)' }}>
                        {log.serial_number}
                      </code>
                    </td>
                    <td><Badge status={log.result} /></td>
                    <td>
                      {log.verified ? (
                        <span style={{ color: 'var(--success)', fontWeight: 600, fontSize: '0.85rem' }}>✓ Yes</span>
                      ) : (
                        <span style={{ color: 'var(--danger)', fontWeight: 600, fontSize: '0.85rem' }}>✗ No</span>
                      )}
                    </td>
                    <td style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>{log.ip_address || '—'}</td>
                    <td style={{ fontSize: '0.75rem', color: 'var(--text-dim)', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {log.user_agent || '—'}
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
              Page {pagination.page} of {pagination.total_pages} ({pagination.total} total scans)
            </span>
            <div style={{ display: 'flex', gap: '0.4rem' }}>
              <button
                disabled={pagination.page <= 1}
                onClick={() => fetchLogs(pagination.page - 1)}
                className="btn btn-ghost btn-sm"
              >
                <ChevronLeft size={16} />
              </button>
              <button
                disabled={pagination.page >= pagination.total_pages}
                onClick={() => fetchLogs(pagination.page + 1)}
                className="btn btn-ghost btn-sm"
              >
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
