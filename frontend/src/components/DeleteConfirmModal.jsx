import React from 'react';
import { AlertTriangle, Trash2, X } from 'lucide-react';

export default function DeleteConfirmModal({ pollTitle, onConfirm, onCancel, isDeleting }) {
  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', color: '#fb7185' }}>
            <AlertTriangle size={22} />
            <h3 style={{ fontSize: '1.2rem', color: '#fff' }}>Delete Poll?</h3>
          </div>
          <button
            onClick={onCancel}
            className="btn btn-outline btn-sm"
            style={{ padding: '0.35rem', borderRadius: '50%' }}
            id="btn-cancel-delete-modal-x"
          >
            <X size={16} />
          </button>
        </div>

        <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem', marginBottom: '1rem' }}>
          Are you sure you want to permanently delete:
        </p>

        <div
          style={{
            padding: '0.85rem 1rem',
            background: 'rgba(255, 255, 255, 0.03)',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--border-subtle)',
            fontWeight: 600,
            marginBottom: '1.5rem',
            color: 'var(--text-primary)',
          }}
        >
          &ldquo;{pollTitle}&rdquo;
        </div>

        <p style={{ color: 'var(--accent-rose)', fontSize: '0.85rem', marginBottom: '1.5rem' }}>
          This will permanently delete the poll, all votes recorded in MongoDB, and its Redis cache.
        </p>

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
          <button onClick={onCancel} className="btn btn-secondary" disabled={isDeleting} id="btn-cancel-delete">
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="btn btn-danger"
            disabled={isDeleting}
            id="btn-confirm-delete"
          >
            <Trash2 size={16} />
            <span>{isDeleting ? 'Deleting...' : 'Delete Permanently'}</span>
          </button>
        </div>
      </div>
    </div>
  );
}
