'use client';

import React, { useState } from 'react';
import { useAuth, User } from '@/context/auth-context';
import { api } from '@/services/api';
import { Lock, User as UserIcon, Loader2, Sparkles } from 'lucide-react';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const response = await api.post<{ token: string; user: User }>('/auth/login', {
        username,
        password,
      });

      const { token, user } = response.data;
      login(token, user);
    } catch (err: any) {
      console.error('Login error:', err);
      setError(err.response?.data?.error || 'Invalid username or password. Please try again.');
      setIsLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-gray-950 via-gray-900 to-black px-4">
      <div className="w-full max-w-md rounded-2xl border border-gray-800 bg-gray-900/80 p-8 shadow-2xl backdrop-blur-xl">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-blue-600/20 text-blue-400 border border-blue-500/30">
            <Sparkles className="h-7 w-7" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white">Smart AI POS Portal</h1>
          <p className="text-sm text-gray-400 mt-1">Sign in to your merchant engine account</p>
        </div>

        {error && (
          <div className="mb-6 rounded-lg bg-red-500/10 border border-red-500/20 p-4 text-sm text-red-400">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-5">
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2">
              Username
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-gray-500">
                <UserIcon className="h-5 w-5" />
              </div>
              <input
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter username"
                className="w-full rounded-xl border border-gray-700 bg-gray-800/50 py-3 pl-11 pr-4 text-white placeholder-gray-500 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2">
              Password
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3.5 pointer-events-none text-gray-500">
                <Lock className="h-5 w-5" />
              </div>
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full rounded-xl border border-gray-700 bg-gray-800/50 py-3 pl-11 pr-4 text-white placeholder-gray-500 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 transition-all"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full flex items-center justify-center rounded-xl bg-blue-600 py-3.5 font-semibold text-white shadow-lg shadow-blue-600/30 hover:bg-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/40 disabled:opacity-50 transition-all"
          >
            {isLoading ? (
              <>
                <Loader2 className="h-5 w-5 animate-spin mr-2" />
                Signing in...
              </>
            ) : (
              'Sign In'
            )}
          </button>
        </form>

        {/* Quick Demo Accounts for Testing */}
        <div className="mt-8 pt-6 border-t border-gray-800">
          <p className="text-xs font-semibold text-gray-400 mb-3 text-center">Quick Demo Accounts (Click to Fill)</p>
          <div className="grid grid-cols-2 gap-2">
            {[
              { name: 'Admin (All Access)', user: 'admin', pass: 'admin123', badge: 'bg-purple-500/20 text-purple-400' },
              { name: 'Cashier (POS & P&L)', user: 'cashier', pass: 'cashier123', badge: 'bg-blue-500/20 text-blue-400' },
              { name: 'Kitchen (KDS Queue)', user: 'kitchen', pass: 'kitchen123', badge: 'bg-amber-500/20 text-amber-400' },
              { name: 'Warehouse (Stock)', user: 'warehouse', pass: 'warehouse123', badge: 'bg-emerald-500/20 text-emerald-400' },
            ].map((demo) => (
              <button
                key={demo.user}
                type="button"
                onClick={() => {
                  setUsername(demo.user);
                  setPassword(demo.pass);
                }}
                className="p-2.5 rounded-xl bg-gray-800/60 hover:bg-gray-800 border border-gray-700/50 text-left transition-all group hover:border-blue-500/40"
              >
                <div className="flex justify-between items-center mb-1">
                  <span className="text-xs font-bold text-white capitalize">{demo.user}</span>
                  <span className={`text-[9px] px-1.5 py-0.5 rounded font-mono ${demo.badge}`}>Demo</span>
                </div>
                <span className="text-[10px] text-gray-400 block truncate">{demo.name}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="mt-6 text-center text-xs text-gray-500">
          Protected by enterprise-grade JWT authentication & RBAC.
        </div>
      </div>
    </div>
  );
}
