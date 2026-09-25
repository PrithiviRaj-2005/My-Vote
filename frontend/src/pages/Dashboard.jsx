import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  PlusCircle,
  LogOut,
  ExternalLink,
  Share2,
  Lock,
  Trash2,
  Calendar,
  Layers,
  AlertCircle,
  CheckCircle,
  RefreshCw,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';
import StatusBadge from '../components/StatusBadge';
import ShareModal from '../components/ShareModal';
import DeleteConfirmModal from '../components/DeleteConfirmModal';

export default function Dashboard() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMessage, setActionMessage] = useState('');

  // Modals state
  const [sharingPoll, setSharingPoll] = useState(null);
  const [deletingPoll, setDeletingPoll] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchPolls = async () => {
    try {
      setLoading(true);
      setError('');
      const data = await api.polls.getUserPolls();
      setPolls(data);
    } catch (err) {
      setError(err.message || 'Failed to load your polls');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPolls();
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const handleClosePoll = async (pollId) => {
    try {
      await api.polls.close(pollId);
      setActionMessage('Poll closed successfully. Further voting has been disabled.');
      setPolls((prev) =>
        prev.map((p) => (p.id === pollId ? { ...p, isActive: false } : p))
      );
      setTimeout(() => setActionMessage(''), 4000);
    } catch (err) {
      setError(err.message || 'Failed to close poll');
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deletingPoll) return;
    try {
      setIsDeleting(true);
      await api.polls.delete(deletingPoll.id);
      setPolls((prev) => prev.filter((p) => p.id !== deletingPoll.id));
      setDeletingPoll(null);
      setActionMessage('Poll deleted successfully.');
      setTimeout(() => setActionMessage(''), 3000);
    } catch (err) {
      setError(err.message || 'Failed to delete poll');
    } finally {
      setIsDeleting(false);
    }
  };

  const formatDate = (isoString) => {
    if (!isoString) return '';
    const date = new Date(isoString);
    return date.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div style={{ maxWidth: '1080px', margin: '0 auto' }}>
      {/* Header Bar */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '1rem',
          marginBottom: '2rem',
          paddingBottom: '1.5rem',
          borderBottom: '1px solid var(--border-subtle)',
        }}
      >
        <div>
          <h1 style={{ fontSize: '2rem', marginBottom: '0.25rem' }}>
            Welcome, <span style={{ color: 'var(--primary)' }}>{user?.name}</span>
          </h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.95rem' }}>
            Manage your live polls, view real-time feedback, and share with your audience.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <Link to="/create" className="btn btn-primary" id="dashboard-btn-create-poll">
            <PlusCircle size={18} />
            <span>Create New Poll</span>
          </Link>
          <button onClick={handleLogout} className="btn btn-secondary" id="dashboard-btn-logout">
            <LogOut size={18} />
            <span>Logout</span>
          </button>
        </div>
      </div>

      {/* Notifications */}
      {actionMessage && (
        <div className="alert alert-success" id="dashboard-alert-action">
          <CheckCircle size={18} style={{ flexShrink: 0 }} />
          <span>{actionMessage}</span>
        </div>
      )}

      {error && (
        <div className="alert alert-error" id="dashboard-alert-error">
          <AlertCircle size={18} style={{ flexShrink: 0 }} />
          <span>{error}</span>
        </div>
      )}

      {/* Polls Section */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
        <h2 style={{ fontSize: '1.35rem' }}>
          Your Polls ({polls.length})
        </h2>
        <button
          onClick={fetchPolls}
          className="btn btn-outline btn-sm"
          style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}
          id="dashboard-btn-refresh"
        >
          <RefreshCw size={14} />
          <span>Refresh</span>
        </button>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: '4rem 1rem', color: 'var(--text-muted)' }}>
          <div className="badge-pulse-dot" style={{ margin: '0 auto 1rem auto', width: '14px', height: '14px' }} />
          <p>Loading your polls from MongoDB...</p>
        </div>
      ) : polls.length === 0 ? (
        <div
          className="card"
          style={{
            textAlign: 'center',
            padding: '4rem 2rem',
            borderStyle: 'dashed',
            borderColor: 'rgba(255, 255, 255, 0.15)',
          }}
        >
          <div
            style={{
              width: '56px',
              height: '56px',
              borderRadius: '50%',
              background: 'rgba(99, 102, 241, 0.1)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 1.25rem auto',
              color: 'var(--primary)',
            }}
          >
            <Layers size={28} />
          </div>
          <h3 style={{ fontSize: '1.35rem', marginBottom: '0.5rem' }}>No Polls Yet</h3>
          <p style={{ color: 'var(--text-secondary)', maxWidth: '420px', margin: '0 auto 1.75rem auto' }}>
            You haven&apos;t created any live polls yet. Create one now and share it with your audience!
          </p>
          <Link to="/create" className="btn btn-primary" id="dashboard-btn-create-first-poll">
            <PlusCircle size={18} />
            <span>Create Your First Poll</span>
          </Link>
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '1.25rem' }}>
          {polls.map((poll) => (
            <div
              key={poll.id}
              className="card"
              style={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
              }}
              id={`dashboard-poll-card-${poll.shareCode}`}
            >
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '0.5rem', marginBottom: '1rem' }}>
                  <StatusBadge isActive={poll.isActive} isConnected={false} />
                  <span
                    style={{
                      fontFamily: 'monospace',
                      fontSize: '0.8rem',
                      padding: '0.2rem 0.5rem',
                      background: 'rgba(255,255,255,0.05)',
                      borderRadius: 'var(--radius-sm)',
                      color: 'var(--text-muted)',
                    }}
                  >
                    #{poll.shareCode}
                  </span>
                </div>

                <h3
                  style={{
                    fontSize: '1.15rem',
                    marginBottom: '1rem',
                    lineHeight: 1.35,
                    wordBreak: 'break-word',
                  }}
                >
                  {poll.question}
                </h3>

                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '1.25rem',
                    color: 'var(--text-secondary)',
                    fontSize: '0.85rem',
                    marginBottom: '1.5rem',
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <Layers size={15} />
                    <span>{poll.options?.length || 0} Options</span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <Calendar size={15} />
                    <span>{formatDate(poll.createdAt)}</span>
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div
                style={{
                  display: 'flex',
                  flexWrap: 'wrap',
                  gap: '0.5rem',
                  paddingTop: '1rem',
                  borderTop: '1px solid var(--border-subtle)',
                }}
              >
                <Link
                  to={`/poll/${poll.shareCode}`}
                  className="btn btn-primary btn-sm"
                  style={{ flex: 1 }}
                  id={`btn-open-poll-${poll.shareCode}`}
                >
                  <ExternalLink size={14} />
                  <span>Open</span>
                </Link>

                <button
                  onClick={() => setSharingPoll(poll)}
                  className="btn btn-secondary btn-sm"
                  id={`btn-share-poll-${poll.shareCode}`}
                  title="Copy Share Link"
                >
                  <Share2 size={14} />
                  <span>Share</span>
                </button>

                {poll.isActive && (
                  <button
                    onClick={() => handleClosePoll(poll.id)}
                    className="btn btn-outline btn-sm"
                    id={`btn-close-poll-${poll.shareCode}`}
                    title="Close Poll"
                  >
                    <Lock size={14} />
                    <span>Close</span>
                  </button>
                )}

                <button
                  onClick={() => setDeletingPoll(poll)}
                  className="btn btn-danger btn-sm"
                  id={`btn-delete-poll-${poll.shareCode}`}
                  title="Delete Poll"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Share Modal */}
      {sharingPoll && <ShareModal poll={sharingPoll} onClose={() => setSharingPoll(null)} />}

      {/* Delete Confirmation Modal */}
      {deletingPoll && (
        <DeleteConfirmModal
          pollTitle={deletingPoll.question}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeletingPoll(null)}
          isDeleting={isDeleting}
        />
      )}
    </div>
  );
}
