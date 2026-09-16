'use client';

import React, { useState, useEffect, useCallback } from 'react';
import ProtectedRoute from '../../../components/protected-route';
import KitchenTicketCard from '../../../components/kitchen-ticket-card';
import { fetchKitchenOrders, updateOrderStatus } from '../../../services/api';
import { Order, OrderStatusUpdateEvent } from '../../../types';
import { ChefHat, RefreshCw, Bell, Layers } from 'lucide-react';

export default function KitchenPortalPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [filter, setFilter] = useState<'ALL' | 'PENDING' | 'PREPARING' | 'READY'>('ALL');
  const [audioEnabled, setAudioEnabled] = useState<boolean>(false);

  const playChime = useCallback(() => {
    if (!audioEnabled) return;
    try {
      const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
      gain.gain.setValueAtTime(0.1, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.5);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start();
      osc.stop(ctx.currentTime + 0.5);
    } catch (e) {
      console.error('Audio playback error:', e);
    }
  }, [audioEnabled]);

  const loadOrders = useCallback(async () => {
    try {
      setLoading(true);
      const data = await fetchKitchenOrders();
      setOrders(data || []);
    } catch (err) {
      console.error('Failed to fetch kitchen orders:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadOrders();
    const interval = setInterval(loadOrders, 15000); // Polling backup
    return () => clearInterval(interval);
  }, [loadOrders]);

  // WebSocket real-time sync
  useEffect(() => {
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'order_status_update' || data.type === 'ORDER_STATUS_UPDATE' || data.type === 'stock_update') {
          loadOrders();
          playChime();
        }
      } catch (e) {
        console.error('WebSocket parse error:', e);
      }
    };

    return () => {
      ws.close();
    };
  }, [loadOrders, playChime]);

  const handleUpdateStatus = async (orderId: string, newStatus: string) => {
    try {
      await updateOrderStatus(orderId, newStatus);
      loadOrders();
    } catch (err) {
      console.error('Failed to update order status:', err);
    }
  };

  const filteredOrders = orders.filter((o) => {
    if (filter === 'ALL') return true;
    return o.status === filter;
  });

  return (
    <ProtectedRoute allowedRoles={['ADMIN', 'KITCHEN', 'CASHIER']}>
      <div className="min-h-screen bg-gray-900 text-white p-6">
        <header className="flex justify-between items-center mb-8 border-b border-gray-800 pb-4">
          <div className="flex items-center gap-3">
            <div className="bg-amber-500/10 p-3 rounded-xl border border-amber-500/20">
              <ChefHat className="w-8 h-8 text-amber-400" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">Kitchen Display System (KDS)</h1>
              <p className="text-sm text-gray-400">Real-time order ticket queues and preparation workflow</p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <button
              onClick={() => setAudioEnabled(!audioEnabled)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                audioEnabled ? 'bg-amber-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              <Bell className="w-4 h-4" />
              {audioEnabled ? 'Chime Active' : 'Enable Chime'}
            </button>

            <button
              onClick={loadOrders}
              className="flex items-center gap-2 bg-gray-800 hover:bg-gray-700 text-gray-200 px-4 py-2 rounded-lg text-sm font-medium transition-colors"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </button>
          </div>
        </header>

        <div className="flex gap-2 mb-6">
          {(['ALL', 'PENDING', 'PREPARING', 'READY'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setFilter(tab)}
              className={`px-4 py-2 rounded-lg text-sm font-semibold transition-colors ${
                filter === tab
                  ? 'bg-amber-500 text-gray-950'
                  : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
              }`}
            >
              {tab} ({orders.filter(o => tab === 'ALL' || o.status === tab).length})
            </button>
          ))}
        </div>

        {loading && orders.length === 0 ? (
          <div className="flex justify-center items-center py-24 text-gray-400">
            <RefreshCw className="w-8 h-8 animate-spin mr-3" /> Loading active kitchen tickets...
          </div>
        ) : filteredOrders.length === 0 ? (
          <div className="bg-gray-800/50 border border-gray-800 rounded-2xl p-16 text-center">
            <Layers className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h2 className="text-xl font-bold text-gray-300 mb-2">No Active Orders</h2>
            <p className="text-sm text-gray-500">New orders placed at checkout will appear here instantly.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredOrders.map((order) => (
              <KitchenTicketCard key={order.id} order={order} onUpdateStatus={handleUpdateStatus} />
            ))}
          </div>
        )}
      </div>
    </ProtectedRoute>
  );
}
