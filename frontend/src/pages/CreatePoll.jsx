import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  Plus,
  Trash2,
  AlertCircle,
  CheckCircle,
  Copy,
  ExternalLink,
  Check,
  Sparkles,
} from 'lucide-react';
import confetti from 'canvas-confetti';
import api from '../services/api';

export default function CreatePoll() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState(['', '']); // Initial 2 options
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [createdPoll, setCreatedPoll] = useState(null);
  const [copied, setCopied] = useState(false);

  const handleOptionChange = (index, value) => {
    const updated = [...options];
    updated[index] = value;
    setOptions(updated);
    setError('');
  };

  const addOption = () => {
    if (options.length >= 10) {
      setError('A poll can have at most 10 options.');
      return;
    }
    setOptions([...options, '']);
    setError('');
  };

  const removeOption = (index) => {
    if (options.length <= 2) {
      setError('A poll must have at least 2 options.');
      return;
    }
    const updated = options.filter((_, i) => i !== index);
    setOptions(updated);
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    // Validations
    if (!question.trim()) {
      setError('Poll question is required.');
      return;
    }

    if (options.length < 2) {
      setError('Minimum 2 options are required.');
      return;
    }

    if (options.length > 10) {
      setError('Maximum 10 options are allowed.');
      return;
    }

    // Check for empty options
    for (let i = 0; i < options.length; i++) {
      if (!options[i].trim()) {
        setError(`Option ${i + 1} cannot be empty.`);
        return;
      }
    }

    // Check for duplicates
    const seen = new Set();
    for (const opt of options) {
      const lower = opt.trim().toLowerCase();
      if (seen.has(lower)) {
        setError(`Duplicate option detected: "${opt.trim()}". Each option must be distinct.`);
        return;
      }
      seen.add(lower);
    }

    try {
      setLoading(true);
      const res = await api.polls.create({
        question: question.trim(),
        options: options.map((opt) => opt.trim()),
      });

      setCreatedPoll(res);

      // Trigger celebratory confetti effect
      try {
        confetti({
          particleCount: 80,
          spread: 70,
          origin: { y: 0.6 },
        });
      } catch (e) {
        // ignore confetti errors in headless env
      }
    } catch (err) {
      setError(err.message || 'Failed to create poll. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleCopyLink = async () => {
    if (!createdPoll) return;
    const shareUrl = `${window.location.origin}/poll/${createdPoll.shareCode}`;
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch (e) {
      console.error('Failed to copy', e);
    }
  };

  return (
    <div style={{ maxWidth: '640px', margin: '2rem auto 0 auto' }}>
      {createdPoll ? (
        /* Success Screen */
        <div className="card" style={{ textAlign: 'center', padding: '3rem 2rem' }}>
          <div
            style={{
              width: '56px',
              height: '56px',
              borderRadius: '50%',
              background: 'rgba(16, 185, 129, 0.15)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 1.25rem auto',
              color: '#34d399',
            }}
          >
            <CheckCircle size={32} />
          </div>

          <h2 style={{ fontSize: '1.8rem', marginBottom: '0.5rem' }}>
            Poll created successfully!
          </h2>

          <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem', fontSize: '0.95rem' }}>
            Your live poll is now active and ready for your audience to cast their votes.
          </p>

          <div
            style={{
              padding: '1.25rem',
              background: 'rgba(0, 0, 0, 0.3)',
              borderRadius: 'var(--radius-md)',
              border: '1px solid var(--border-subtle)',
              marginBottom: '2rem',
              textAlign: 'left',
            }}
          >
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', display: 'block', marginBottom: '0.35rem' }}>
              SHARE LINK
            </span>
            <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
              <input
                type="text"
                readOnly
                value={`${window.location.origin}/poll/${createdPoll.shareCode}`}
                className="form-input"
                style={{ fontSize: '0.9rem' }}
                id="created-poll-share-link"
              />
              <button
                onClick={handleCopyLink}
                className={`btn ${copied ? 'btn-secondary' : 'btn-primary'}`}
                id="btn-created-copy-link"
              >
                {copied ? <Check size={16} /> : <Copy size={16} />}
                <span>{copied ? 'Copied!' : 'Copy Link'}</span>
              </button>
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', flexWrap: 'wrap' }}>
            <Link
              to={`/poll/${createdPoll.shareCode}`}
              className="btn btn-primary btn-lg"
              id="btn-created-open-poll"
            >
              <span>Open Poll</span>
              <ExternalLink size={18} />
            </Link>

            <button
              onClick={() => {
                setCreatedPoll(null);
                setQuestion('');
                setOptions(['', '']);
              }}
              className="btn btn-secondary btn-lg"
              id="btn-created-create-another"
            >
              <Sparkles size={18} />
              <span>Create Another Poll</span>
            </button>

            <Link to="/dashboard" className="btn btn-outline btn-lg" id="btn-created-go-dashboard">
              Go to Dashboard
            </Link>
          </div>
        </div>
      ) : (
        /* Create Poll Form */
        <div className="card">
          <div style={{ marginBottom: '1.75rem' }}>
            <h2 style={{ fontSize: '1.75rem', marginBottom: '0.4rem' }}>Create New Poll</h2>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
              Set up your question and 2 to 10 voting choices.
            </p>
          </div>

          {error && (
            <div className="alert alert-error" id="create-poll-alert-error">
              <AlertCircle size={18} style={{ flexShrink: 0 }} />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit}>
            <div className="form-group" style={{ marginBottom: '1.75rem' }}>
              <label className="form-label" htmlFor="poll-question">
                Question <span style={{ color: 'var(--accent-rose)' }}>*</span>
              </label>
              <input
                id="poll-question"
                type="text"
                className="form-input"
                placeholder="e.g. Which programming language do you prefer?"
                value={question}
                onChange={(e) => {
                  setQuestion(e.target.value);
                  setError('');
                }}
                required
              />
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: '0.75rem',
                }}
              >
                <label className="form-label" style={{ marginBottom: 0 }}>
                  Options ({options.length}/10) <span style={{ color: 'var(--accent-rose)' }}>*</span>
                </label>
                <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                  Minimum 2 options
                </span>
              </div>

              {options.map((opt, idx) => (
                <div
                  key={idx}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '0.5rem',
                    marginBottom: '0.75rem',
                  }}
                >
                  <span
                    style={{
                      width: '28px',
                      fontSize: '0.85rem',
                      color: 'var(--text-muted)',
                      textAlign: 'right',
                    }}
                  >
                    {idx + 1}.
                  </span>
                  <input
                    type="text"
                    className="form-input"
                    placeholder={`Option ${idx + 1}`}
                    value={opt}
                    onChange={(e) => handleOptionChange(idx, e.target.value)}
                    required
                    id={`poll-option-input-${idx + 1}`}
                  />
                  {options.length > 2 && (
                    <button
                      type="button"
                      onClick={() => removeOption(idx)}
                      className="btn btn-danger btn-sm"
                      title="Remove Option"
                      id={`btn-remove-option-${idx + 1}`}
                    >
                      <Trash2 size={16} />
                    </button>
                  )}
                </div>
              ))}

              {options.length < 10 && (
                <button
                  type="button"
                  onClick={addOption}
                  className="btn btn-secondary btn-sm"
                  style={{ marginTop: '0.5rem', marginLeft: '2rem' }}
                  id="btn-add-option"
                >
                  <Plus size={16} />
                  <span>Add Option</span>
                </button>
              )}
            </div>

            <button
              type="submit"
              className="btn btn-primary btn-lg"
              style={{ width: '100%', marginTop: '1rem' }}
              disabled={loading}
              id="btn-submit-create-poll"
            >
              {loading ? 'Creating Poll...' : 'Create Poll'}
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
