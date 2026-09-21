'use client';

import React, { useState, useEffect } from 'react';
import { getCurrentCashShift, openCashShift, closeCashShift } from '../services/api';
import { CashShift } from '../types';
import { formatIDR } from '../utils/format';
import { Wallet, X, Lock, Unlock, CheckCircle2, AlertCircle, RefreshCw } from 'lucide-react';

interface CashDrawerModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function CashDrawerModal({ isOpen, onClose }: CashDrawerModalProps) {
  const [currentShift, setCurrentShift] = useState<CashShift | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [openingCash, setOpeningCash] = useState<string>('500000');
  const [closingCash, setClosingCash] = useState<string>('');
  const [notes, setNotes] = useState<string>('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<boolean>(false);

  const fetchShift = async () => {
    try {
      setLoading(true);
      setError(null);
      const shiftData = await getCurrentCashShift();
      setCurrentShift(shiftData);
    } catch (err: any) {
      if (err.response?.status === 404) {
        setCurrentShift(null);
      } else {
        setError(err.response?.data?.error || 'Failed to fetch cash shift');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchShift();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleOpenShift = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setError(null);
    try {
      await openCashShift({
        opening_cash: parseFloat(openingCash) || 0,
        notes,
      });
      setSuccess('Cash drawer opened successfully!');
      setTimeout(() => setSuccess(null), 2000);
      fetchShift();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to open shift');
    } finally {
      setActionLoading(false);
    }
  };

  const handleCloseShift = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setError(null);
    try {
      const res = await closeCashShift({
        closing_cash: parseFloat(closingCash) || 0,
        notes,
      });
      setSuccess(`Shift closed. Difference: ${formatIDR(res.difference)}`);
      setTimeout(() => {
        setSuccess(null);
        setCurrentShift(null);
        onClose();
      }, 2500);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to close shift');
    } finally {
      setActionLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="glass-panel border border-gray-700 rounded-2xl max-w-md w-full p-6 shadow-2xl relative animate-scaleUp">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-white bg-dark-800 p-2 rounded-xl"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="flex items-center space-x-3 mb-6">
          <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl border border-brand-500/25">
            <Wallet className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Buka / Tutup Kasir (Cash Shift)</h2>
            <p className="text-xs text-gray-400">Manage daily cash register drawer and cashflow</p>
          </div>
        </div>

        {error && (
          <div className="bg-red-500/10 border border-red-500/30 text-red-400 p-3 rounded-xl text-xs mb-4 flex items-center space-x-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {success && (
          <div className="bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 p-3 rounded-xl text-xs mb-4 flex items-center space-x-2">
            <CheckCircle2 className="w-4 h-4 shrink-0" />
            <span>{success}</span>
          </div>
        )}

        {loading ? (
          <div className="py-12 text-center text-gray-500 flex flex-col items-center">
            <RefreshCw className="w-6 h-6 animate-spin mb-2" />
            <span>Checking drawer status...</span>
          </div>
        ) : currentShift ? (
          <form onSubmit={handleCloseShift} className="space-y-4">
            <div className="bg-dark-800/80 p-4 rounded-xl border border-gray-700/50 space-y-2 text-xs">
              <div className="flex justify-between items-center text-emerald-400 font-bold">
                <span>Status: DRAWER OPEN</span>
                <Unlock className="w-4 h-4" />
              </div>
              <div className="flex justify-between text-gray-400">
                <span>Opened At:</span>
                <span>{new Date(currentShift.opened_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
              </div>
              <div className="flex justify-between text-gray-400">
                <span>Starting Cash:</span>
                <span className="font-mono text-white">{formatIDR(currentShift.opening_cash)}</span>
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-gray-300 mb-1">Actual Closing Cash (Drawer Count) *</label>
              <input
                type="number"
                step="1000"
                required
                placeholder="e.g. 750000"
                value={closingCash}
                onChange={(e) => setClosingCash(e.target.value)}
                className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 font-mono"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-gray-300 mb-1">Shift Closing Notes</label>
              <textarea
                rows={2}
                placeholder="Optional notes or discrepancy remarks..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 resize-none"
              />
            </div>

            <button
              type="submit"
              disabled={actionLoading}
              className="w-full bg-red-600 hover:bg-red-700 text-white font-bold py-3 rounded-xl text-xs flex items-center justify-center space-x-2 transition-all shadow-lg"
            >
              {actionLoading ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Lock className="w-4 h-4" />}
              <span>Close Shift & Balance Drawer</span>
            </button>
          </form>
        ) : (
          <form onSubmit={handleOpenShift} className="space-y-4">
            <div className="bg-dark-800/80 p-4 rounded-xl border border-gray-700/50 space-y-2 text-xs text-amber-400 font-bold flex items-center justify-between">
              <span>Status: DRAWER CLOSED</span>
              <Lock className="w-4 h-4" />
            </div>

            <div>
              <label className="block text-xs font-semibold text-gray-300 mb-1">Opening Cash (Modal Awal) *</label>
              <input
                type="number"
                step="1000"
                required
                placeholder="e.g. 500000"
                value={openingCash}
                onChange={(e) => setOpeningCash(e.target.value)}
                className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 font-mono"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-gray-300 mb-1">Opening Notes</label>
              <textarea
                rows={2}
                placeholder="Optional opening remarks..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 resize-none"
              />
            </div>

            <button
              type="submit"
              disabled={actionLoading}
              className="w-full bg-brand-500 hover:bg-brand-600 text-black font-bold py-3 rounded-xl text-xs flex items-center justify-center space-x-2 transition-all shadow-lg shadow-brand-500/20"
            >
              {actionLoading ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Unlock className="w-4 h-4" />}
              <span>Open Cash Drawer (Buka Kasir)</span>
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
