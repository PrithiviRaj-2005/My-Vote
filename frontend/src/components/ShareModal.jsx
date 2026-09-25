import React, { useState } from 'react';
import { Copy, Check, ExternalLink, X, Share2 } from 'lucide-react';

export default function ShareModal({ poll, onClose }) {
  const [copied, setCopied] = useState(false);
  const shareUrl = `${window.location.origin}/poll/${poll.shareCode}`;

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch (e) {
      console.error('Failed to copy to clipboard', e);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <Share2 size={20} color="var(--primary)" />
            <h3 style={{ fontSize: '1.25rem' }}>Share Poll</h3>
          </div>
          <button
            onClick={onClose}
            className="btn btn-outline btn-sm"
            style={{ padding: '0.35rem', borderRadius: '50%' }}
            id="btn-close-share-modal"
          >
            <X size={16} />
          </button>
        </div>

        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '1.25rem' }}>
          Anyone with this link can participate and vote live in real time without creating an account:
        </p>

        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem' }}>
          <input
            type="text"
            readOnly
            value={shareUrl}
            className="form-input"
            style={{ fontSize: '0.88rem', background: 'rgba(0,0,0,0.3)' }}
            id="share-link-input"
          />
          <button
            onClick={handleCopy}
            className={`btn ${copied ? 'btn-secondary' : 'btn-primary'}`}
            id="btn-copy-share-link"
          >
            {copied ? <Check size={16} /> : <Copy size={16} />}
            <span>{copied ? 'Copied!' : 'Copy'}</span>
          </button>
        </div>

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
          <button onClick={onClose} className="btn btn-secondary">
            Done
          </button>
          <a
            href={`/poll/${poll.shareCode}`}
            target="_blank"
            rel="noopener noreferrer"
            className="btn btn-primary"
            id="btn-open-share-link-newtab"
          >
            <span>Open Poll</span>
            <ExternalLink size={16} />
          </a>
        </div>
      </div>
    </div>
  );
}
