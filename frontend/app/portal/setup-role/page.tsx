'use client';

import React, { useState, useEffect } from 'react';
import ProtectedRoute from '../../../components/protected-route';
import { UserRole } from '../../../types';
import { api } from '../../../services/api';
import { ShieldCheck, RefreshCw, CheckCircle2, AlertCircle, Lock } from 'lucide-react';

interface AccessMenu {
  id: string;
  key: string;
  name: string;
  path: string;
  icon: string;
}

interface RoleAccessConfig {
  role: string;
  menu_key: string;
  can_access: boolean;
}

const ROLES = ['ADMIN', 'CASHIER', 'KITCHEN', 'WAREHOUSE'];

export default function SetupRolePage() {
  const [menus, setMenus] = useState<AccessMenu[]>([]);
  const [mappings, setMappings] = useState<RoleAccessConfig[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [menusRes, mappingsRes] = await Promise.all([
        api.get<AccessMenu[]>('/access-menus'),
        api.get<RoleAccessConfig[]>('/access-menus/roles'),
      ]);
      setMenus(menusRes.data || []);
      setMappings(mappingsRes.data || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load access menu configuration');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const hasAccess = (role: string, menuKey: string): boolean => {
    const found = mappings.find((m) => m.role === role && m.menu_key === menuKey);
    return found ? found.can_access : false;
  };

  const handleToggle = async (role: string, menuKey: string, currentAccess: boolean) => {
    if (role === 'ADMIN' && menuKey === 'SETUP_ROLE') {
      return; // Admin always has setup role access
    }

    try {
      setSaving(true);
      setError(null);
      setMessage(null);

      await api.put('/access-menus/roles', {
        role,
        menu_key: menuKey,
        can_access: !currentAccess,
      });

      // Optimistic local state update
      setMappings((prev) => {
        const existing = prev.find((m) => m.role === role && m.menu_key === menuKey);
        if (existing) {
          return prev.map((m) => (m.role === role && m.menu_key === menuKey ? { ...m, can_access: !currentAccess } : m));
        }
        return [...prev, { role, menu_key: menuKey, can_access: !currentAccess }];
      });

      setMessage(`Updated access for ${role} -> ${menuKey}`);
      setTimeout(() => setMessage(null), 2500);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update access permission');
    } finally {
      setSaving(false);
    }
  };

  return (
    <ProtectedRoute allowedRoles={[UserRole.ADMIN]}>
      <div className="space-y-6">
        <header className="flex justify-between items-center bg-gray-900 border border-gray-800 p-6 rounded-2xl">
          <div className="flex items-center space-x-3">
            <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl border border-brand-500/20">
              <ShieldCheck className="w-7 h-7" />
            </div>
            <div>
              <h1 className="text-xl font-black text-white">Setup Role Access Controls</h1>
              <p className="text-xs text-gray-400">Dynamic RBAC configuration managing which roles can access specific portal menus</p>
            </div>
          </div>

          <button
            onClick={fetchData}
            className="flex items-center space-x-2 bg-dark-800 hover:bg-dark-700 text-gray-200 px-4 py-2.5 rounded-xl text-xs font-bold transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh Config</span>
          </button>
        </header>

        {message && (
          <div className="bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 p-3 rounded-xl text-xs flex items-center space-x-2">
            <CheckCircle2 className="w-4 h-4 shrink-0" />
            <span>{message}</span>
          </div>
        )}

        {error && (
          <div className="bg-red-500/10 border border-red-500/30 text-red-400 p-3 rounded-xl text-xs flex items-center space-x-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <div className="glass-panel rounded-2xl border border-gray-700/50 overflow-hidden p-6">
          <div className="mb-6">
            <h2 className="text-sm font-bold text-white">Role Permission Matrix</h2>
            <p className="text-xs text-gray-400 mt-1">Check or uncheck access permissions per role. Changes apply immediately.</p>
          </div>

          {loading ? (
            <div className="py-16 text-center text-gray-500">Loading access menu configurations...</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="bg-dark-800/80 text-gray-400 border-b border-gray-700/50">
                    <th className="p-4 font-bold">Access Menu</th>
                    <th className="p-4 font-bold">Route Path</th>
                    {ROLES.map((role) => (
                      <th key={role} className="p-4 font-bold text-center">{role}</th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-800">
                  {menus.map((menu) => (
                    <tr key={menu.key} className="hover:bg-dark-800/40 transition-colors">
                      <td className="p-4 font-bold text-white flex items-center space-x-2">
                        <span>{menu.name}</span>
                        <span className="text-[10px] text-brand-500 font-mono bg-brand-500/10 px-2 py-0.5 rounded">
                          {menu.key}
                        </span>
                      </td>
                      <td className="p-4 font-mono text-gray-400">{menu.path}</td>
                      {ROLES.map((role) => {
                        const allowed = hasAccess(role, menu.key);
                        const isLockedAdmin = role === 'ADMIN' && menu.key === 'SETUP_ROLE';
                        return (
                          <td key={role} className="p-4 text-center">
                            {isLockedAdmin ? (
                              <span className="inline-flex items-center justify-center p-1.5 bg-brand-500/20 text-brand-400 rounded-lg" title="Required for Admin">
                                <Lock className="w-4 h-4" />
                              </span>
                            ) : (
                              <input
                                type="checkbox"
                                checked={allowed}
                                disabled={saving}
                                onChange={() => handleToggle(role, menu.key, allowed)}
                                className="w-4 h-4 accent-brand-500 cursor-pointer rounded"
                              />
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </ProtectedRoute>
  );
}
