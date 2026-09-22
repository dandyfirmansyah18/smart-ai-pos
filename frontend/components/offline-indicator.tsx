'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchSyncStatus, triggerSync, SyncStatus } from '../services/api';
import { Wifi, WifiOff, RefreshCw } from 'lucide-react';

export function OfflineIndicator() {
  const queryClient = useQueryClient();

  const { data: status, isLoading } = useQuery<SyncStatus>({
    queryKey: ['syncStatus'],
    queryFn: fetchSyncStatus,
    refetchInterval: 15000,
  });

  const syncMutation = useMutation({
    mutationFn: triggerSync,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['syncStatus'] });
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });

  if (isLoading || !status) {
    return null;
  }

  const isOnline = status.is_online;
  const totalPending = (status.pending_orders || 0) + (status.pending_shifts || 0) + (status.pending_payments || 0);

  return (
    <div className="flex items-center space-x-2 text-xs font-bold">
      {isOnline ? (
        <div className="flex items-center space-x-1.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2.5 py-1.5 rounded-xl">
          <Wifi className="w-3.5 h-3.5 animate-pulse text-emerald-400" />
          <span className="hidden sm:inline">Online (Cloud)</span>
        </div>
      ) : (
        <div className="flex items-center space-x-1.5 bg-amber-500/10 text-amber-400 border border-amber-500/20 px-2.5 py-1.5 rounded-xl">
          <WifiOff className="w-3.5 h-3.5 text-amber-400" />
          <span className="hidden sm:inline">Offline (SQLite)</span>
          {totalPending > 0 && (
            <span className="bg-amber-500 text-gray-950 px-1.5 py-0.2 rounded-full text-[10px] font-black">
              {totalPending}
            </span>
          )}
        </div>
      )}

      {(!isOnline || totalPending > 0) && (
        <button
          onClick={() => syncMutation.mutate()}
          disabled={syncMutation.isPending}
          className="flex items-center space-x-1 bg-blue-600/20 text-blue-400 border border-blue-500/30 px-3 py-1.5 rounded-xl hover:bg-blue-600/30 transition-all disabled:opacity-50"
          title="Force Sync Now"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${syncMutation.isPending ? 'animate-spin' : ''}`} />
          <span className="hidden md:inline">Sync Now</span>
        </button>
      )}
    </div>
  );
}
