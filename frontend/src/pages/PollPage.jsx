import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  CheckCircle2,
  AlertCircle,
  Share2,
  Copy,
  Check,
  Radio,
  BarChart3,
  Lock,
  Vote,
  Sparkles,
} from 'lucide-react';
import confetti from 'canvas-confetti';
import { usePoll } from '../hooks/usePoll';
import StatusBadge from '../components/StatusBadge';
import ProgressBar from '../components/ProgressBar';

export default function PollPage() {
  const { shareCode } = useParams();
  const { poll, results, loading, error, isConnected, submitVote } = usePoll(shareCode);

  const [selectedOption, setSelectedOption] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [voteSuccess, setVoteSuccess] = useState('');
  const [voteError, setVoteError] = useState('');
  const [copied, setCopied] = useState(false);
  const [viewResultsMode, setViewResultsMode] = useState(false);

  // Check if this browser already voted for this poll
  const votedStorageKey = `pulsevote_voted_${shareCode}`;
  const [hasVoted, setHasVoted] = useState(false);
  const [votedOptionId, setVotedOptionId] = useState(null);

  useEffect(() => {
    const savedVote = localStorage.getItem(votedStorageKey);
    if (savedVote) {
      setHasVoted(true);
      setVotedOptionId(savedVote);
      setViewResultsMode(true);
    }
  }, [votedStorageKey]);

  const handleVote = async (optionId) => {
    if (!poll?.isActive || submitting) return;

    try {
      setSubmitting(true);
      setSelectedOption(optionId);
      setVoteError('');

      const res = await submitVote(optionId);
      if (res.success) {
        setHasVoted(true);
        setVotedOptionId(optionId);
        localStorage.setItem(votedStorageKey, optionId);
        setVoteSuccess('Vote submitted successfully!');
        setViewResultsMode(true);

        // Confetti celebration
        try {
          confetti({
            particleCount: 60,
            spread: 60,
            origin: { y: 0.7 },
          });
        } catch (e) {
          // Ignore
        }
      } else {
        setVoteError(res.error || 'Failed to submit vote');
      }
    } catch (err) {
      setVoteError(err.message || 'Error casting vote');
    } finally {
      setSubmitting(false);
    }
  };

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch (e) {
      console.error('Failed to copy', e);
    }
  };

  if (loading) {
    return (
      <div style={{ maxWidth: '640px', margin: '4rem auto', textAlign: 'center' }}>
        <div className="badge-pulse-dot" style={{ margin: '0 auto 1.5rem auto', width: '16px', height: '16px' }} />
        <h3 style={{ fontSize: '1.25rem' }}>Connecting to Live Poll Stream...</h3>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginTop: '0.5rem' }}>
          Establishing Gorilla WebSocket connection & Redis pub/sub listener...
        </p>
      </div>
    );
  }

  if (error || !poll) {
    return (
      <div style={{ maxWidth: '520px', margin: '4rem auto' }}>
        <div className="card" style={{ textAlign: 'center', padding: '3rem 2rem' }}>
          <div style={{ color: '#fb7185', marginBottom: '1rem' }}>
            <AlertCircle size={48} style={{ margin: '0 auto' }} />
          </div>
          <h2 style={{ fontSize: '1.5rem', marginBottom: '0.75rem' }}>Poll Unavailable</h2>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1.75rem' }}>
            {error || "The poll you're looking for does not exist or has been deleted."}
          </p>
          <Link to="/" className="btn btn-primary">
            Return to PulseVote Home
          </Link>
        </div>
      </div>
    );
  }

  const isClosed = !poll.isActive;
  const optionsList = results?.options || poll.options?.map((opt) => ({ ...opt, votes: 0, percentage: 0 })) || [];
  const totalVotes = results?.totalVotes || 0;

  // Determine highest voted option
  const maxVotes = Math.max(...optionsList.map((o) => o.votes || 0), 0);

  return (
    <div style={{ maxWidth: '680px', margin: '1.5rem auto 0 auto' }}>
      {/* Top Banner & WebSocket Status Indicator */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '0.75rem',
          marginBottom: '1.25rem',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
          <StatusBadge isActive={poll.isActive} isConnected={isConnected} />
          {isConnected && (
            <span style={{ fontSize: '0.78rem', color: '#34d399', fontWeight: 500 }}>
              Live updating without reload
            </span>
          )}
        </div>

        <button
          onClick={handleCopyLink}
          className="btn btn-outline btn-sm"
          id="btn-public-share-link"
          title="Share Poll Link"
        >
          {copied ? <Check size={14} /> : <Share2 size={14} />}
          <span>{copied ? 'Link Copied!' : 'Share Poll'}</span>
        </button>
      </div>

      {/* Closed Poll Warning Notice */}
      {isClosed && (
        <div
          className="alert"
          style={{
            background: 'rgba(245, 158, 11, 0.15)',
            border: '1px solid rgba(245, 158, 11, 0.35)',
            color: '#fde68a',
            marginBottom: '1.5rem',
          }}
          id="poll-closed-notice"
        >
          <Lock size={18} style={{ flexShrink: 0 }} />
          <span>
            <strong>This poll is closed.</strong> Voting is disabled. Existing results can still be viewed.
          </span>
        </div>
      )}

      {/* Success Notification */}
      {voteSuccess && (
        <div className="alert alert-success" id="poll-vote-success-alert">
          <CheckCircle2 size={18} style={{ flexShrink: 0 }} />
          <span>{voteSuccess}</span>
        </div>
      )}

      {/* Error Notification */}
      {voteError && (
        <div className="alert alert-error" id="poll-vote-error-alert">
          <AlertCircle size={18} style={{ flexShrink: 0 }} />
          <span>{voteError}</span>
        </div>
      )}

      {/* Main Card */}
      <div className="card" style={{ padding: '2.25rem' }}>
        {/* Question Header */}
        <div style={{ marginBottom: '2rem' }}>
          <span
            style={{
              fontSize: '0.8rem',
              color: 'var(--text-muted)',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontWeight: 600,
              display: 'block',
              marginBottom: '0.5rem',
            }}
          >
            Audience Poll
          </span>
          <h1
            style={{
              fontSize: 'clamp(1.4rem, 3vw, 1.85rem)',
              lineHeight: 1.3,
              fontWeight: 700,
            }}
            id="public-poll-question"
          >
            {poll.question}
          </h1>
        </div>

        {/* Voting UI: Interactive Options */}
        {poll.isActive && !viewResultsMode && (
          <div style={{ marginBottom: '2rem' }}>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: '1rem' }}>
              Select an option to cast your vote:
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {poll.options.map((option) => {
                const isSelected = selectedOption === option.id;
                return (
                  <button
                    key={option.id}
                    onClick={() => handleVote(option.id)}
                    disabled={submitting}
                    className={`vote-option-btn ${isSelected ? 'selected' : ''}`}
                    id={`btn-vote-option-${option.id}`}
                  >
                    <span>{option.text}</span>
                    <span
                      style={{
                        fontSize: '0.85rem',
                        color: 'var(--text-muted)',
                        padding: '0.2rem 0.6rem',
                        borderRadius: 'var(--radius-sm)',
                        background: 'rgba(255,255,255,0.05)',
                      }}
                    >
                      Vote
                    </span>
                  </button>
                );
              })}
            </div>
          </div>
        )}

        {/* Live Results UI: Bars, Percentages, and Counts */}
        {(viewResultsMode || isClosed) && (
          <div style={{ marginBottom: '1.5rem' }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '1.25rem',
                paddingBottom: '0.75rem',
                borderBottom: '1px solid var(--border-subtle)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-primary)' }}>
                <BarChart3 size={18} color="var(--primary)" />
                <h3 style={{ fontSize: '1.1rem' }}>Live Results</h3>
              </div>
              <span style={{ fontSize: '0.88rem', color: 'var(--text-secondary)' }}>
                Total: <strong style={{ color: 'var(--text-primary)' }}>{totalVotes}</strong>{' '}
                {totalVotes === 1 ? 'vote' : 'votes'}
              </span>
            </div>

            {optionsList.map((opt) => (
              <ProgressBar
                key={opt.id}
                option={opt}
                isUserVoted={votedOptionId === opt.id}
                isLeader={maxVotes > 0 && opt.votes === maxVotes}
              />
            ))}
          </div>
        )}

        {/* View Toggle (Vote vs Results) if poll is active */}
        {poll.isActive && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              marginTop: '1.5rem',
              paddingTop: '1.25rem',
              borderTop: '1px solid var(--border-subtle)',
              fontSize: '0.88rem',
            }}
          >
            <button
              onClick={() => setViewResultsMode(!viewResultsMode)}
              className="btn btn-outline btn-sm"
              id="btn-toggle-results-view"
            >
              <BarChart3 size={15} />
              <span>{viewResultsMode ? 'Show Voting Buttons' : 'View Current Results'}</span>
            </button>

            <span style={{ color: 'var(--text-muted)', fontSize: '0.82rem' }}>
              {totalVotes} total votes recorded
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
