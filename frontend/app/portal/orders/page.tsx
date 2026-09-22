'use client';

import React, { useState } from 'react';
import ProtectedRoute from '../../../components/protected-route';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchOrderHistory, createPaymentCharge, getPaymentByOrderID, notifyPaymentStatus } from '../../../services/api';
import { OrderStatus, UserRole, PaymentMethod } from '../../../types';
import { formatIDR } from '../../../utils/format';
import { ClipboardList, RefreshCw, Search, CreditCard } from 'lucide-react';

const loadSnapScript = (): Promise<void> => {
  return new Promise((resolve) => {
    if (typeof window !== 'undefined' && window.snap) {
      resolve();
      return;
    }
    const snapUrl = process.env.NEXT_PUBLIC_MIDTRANS_SNAP_URL || 'https://app.sandbox.midtrans.com/snap/snap.js';
    const clientKey = process.env.NEXT_PUBLIC_MIDTRANS_CLIENT_KEY || 'SB-Mid-client-your_client_key';

    const existingScript = document.getElementById('midtrans-snap-script');
    if (existingScript) {
      existingScript.onload = () => resolve();
      return;
    }

    const script = document.createElement('script');
    script.id = 'midtrans-snap-script';
    script.src = snapUrl;
    script.setAttribute('data-client-key', clientKey);
    script.onload = () => resolve();
    document.body.appendChild(script);
  });
};

export default function OrderHistoryPage() {
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState('');
  const [repaymentLoading, setRepaymentLoading] = useState<string | null>(null);

  const { data: orders = [], isLoading, refetch } = useQuery({
    queryKey: ['orderHistory'],
    queryFn: fetchOrderHistory,
  });

  const handleRepay = async (orderId: string, amount: number) => {
    try {
      setRepaymentLoading(orderId);
      await loadSnapScript();

      let snapToken = '';
      try {
        const existingPay = await getPaymentByOrderID(orderId);
        if (existingPay && existingPay.snap_token) {
          snapToken = existingPay.snap_token;
        }
      } catch (e) {
        // If payment record not found, initiate charge
      }

      if (!snapToken) {
        const payRes = await createPaymentCharge({
          order_id: orderId,
          amount,
          payment_method: PaymentMethod.MIDTRANS,
        });
        snapToken = payRes.snap_token || '';
      }

      if (snapToken && window.snap) {
        window.snap.pay(snapToken, {
          onSuccess: async (result) => {
            console.log('Repayment Success:', result);
            const targetOrderId = result?.order_id || result?.orderId || orderId;
            try {
              await notifyPaymentStatus({
                order_id: targetOrderId,
                transaction_status: result?.transaction_status || 'settlement',
                status_code: result?.status_code,
                gross_amount: result?.gross_amount,
              });
            } catch (err) {
              console.error('Failed to notify repayment success:', err);
            }
            queryClient.setQueryData(['orderHistory'], (oldOrders: any[] | undefined) => {
              if (!oldOrders) return [];
              return oldOrders.map((o) =>
                o.id === targetOrderId ? { ...o, status: OrderStatus.PENDING } : o
              );
            });
            queryClient.invalidateQueries({ queryKey: ['orderHistory'] });
            queryClient.invalidateQueries({ queryKey: ['kitchenOrders'] });
            queryClient.invalidateQueries({ queryKey: ['products'] });
            refetch();
          },
          onPending: async (result) => {
            console.log('Repayment Pending:', result);
            try {
              await notifyPaymentStatus({
                order_id: result?.order_id || orderId,
                transaction_status: result?.transaction_status || 'pending',
                status_code: result?.status_code,
                gross_amount: result?.gross_amount,
              });
            } catch (err) {
              console.error('Failed to notify repayment pending:', err);
            }
            queryClient.invalidateQueries({ queryKey: ['orderHistory'] });
            refetch();
          },
          onError: (result) => {
            console.error('Repayment Error:', result);
          },
          onClose: () => {
            console.log('Repayment popup closed');
          },
        });
      }
    } catch (err) {
      console.error('Failed to initiate repayment:', err);
    } finally {
      setRepaymentLoading(null);
    }
  };

  const filteredOrders = orders.filter((o) =>
    o.transaction_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
    o.id.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <ProtectedRoute allowedRoles={[UserRole.ADMIN, UserRole.CASHIER]}>
      <div className="space-y-6">
        <header className="flex justify-between items-center bg-gray-900 border border-gray-800 p-6 rounded-2xl">
          <div className="flex items-center space-x-3">
            <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl border border-brand-500/20">
              <ClipboardList className="w-7 h-7" />
            </div>
            <div>
              <h1 className="text-xl font-black text-white">Order History & Transaction Logs</h1>
              <p className="text-xs text-gray-400">Review completed orders, receipt summaries, and item breakdown</p>
            </div>
          </div>

          <button
            onClick={() => refetch()}
            className="flex items-center space-x-2 bg-dark-800 hover:bg-dark-700 text-gray-200 px-4 py-2.5 rounded-xl text-xs font-bold transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            <span>Refresh Logs</span>
          </button>
        </header>

        {/* Search & Filter Bar */}
        <div className="glass-panel p-4 rounded-2xl border border-gray-700/50 flex items-center justify-between">
          <div className="relative w-full max-w-md">
            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-gray-500">
              <Search className="w-4 h-4" />
            </div>
            <input
              type="text"
              placeholder="Search by Transaction ID or Order ID..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full bg-dark-800 border border-gray-700 rounded-xl py-2.5 pl-10 pr-4 text-xs text-white placeholder-gray-500 focus:outline-none focus:border-brand-500"
            />
          </div>
          <span className="text-xs text-gray-400 font-mono">Total Transactions: {orders.length}</span>
        </div>

        {/* Orders Table */}
        <div className="glass-panel rounded-2xl border border-gray-700/50 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="bg-dark-800/80 text-gray-400 border-b border-gray-700/50">
                  <th className="p-4 font-bold">Transaction ID</th>
                  <th className="p-4 font-bold">Order UUID</th>
                  <th className="p-4 font-bold">Status</th>
                  <th className="p-4 font-bold">Items Summary</th>
                  <th className="p-4 font-bold">Total Amount</th>
                  <th className="p-4 font-bold text-center">Action / Repayment</th>
                  <th className="p-4 font-bold text-right">Created Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800">
                {isLoading ? (
                  <tr>
                    <td colSpan={7} className="p-12 text-center text-gray-500">
                      Loading order history logs...
                    </td>
                  </tr>
                ) : filteredOrders.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="p-12 text-center text-gray-500">
                      No transaction records found.
                    </td>
                  </tr>
                ) : (
                  filteredOrders.map((order) => (
                    <tr key={order.id} className="hover:bg-dark-800/40 transition-colors">
                      <td className="p-4 font-mono font-bold text-brand-500">{order.transaction_id}</td>
                      <td className="p-4 font-mono text-gray-400">{order.id.slice(0, 13)}...</td>
                      <td className="p-4">
                        <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold ${
                          order.status === OrderStatus.UNPAID ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30 animate-pulse' :
                          order.status === OrderStatus.PENDING ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30' :
                          order.status === OrderStatus.PREPARING ? 'bg-indigo-500/20 text-indigo-400 border border-indigo-500/30' :
                          order.status === OrderStatus.READY ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' :
                          order.status === OrderStatus.EXPIRED || order.status === OrderStatus.CANCELLED ? 'bg-red-500/20 text-red-400 border border-red-500/30' :
                          'bg-gray-800 text-gray-300'
                        }`}>
                          {order.status}
                        </span>
                      </td>
                      <td className="p-4 text-gray-300">
                        {order.items && order.items.map((i, idx) => (
                          <div key={i.id || idx} className="text-[11px]">
                            <span className="font-bold text-white">{i.quantity}x</span> {i.name || i.sku}
                          </div>
                        ))}
                      </td>
                      <td className="p-4 font-mono font-black text-white">{formatIDR(order.total_amount)}</td>
                      <td className="p-4 text-center">
                        {order.status === OrderStatus.UNPAID ? (
                          <button
                            onClick={() => handleRepay(order.id, order.total_amount)}
                            disabled={repaymentLoading === order.id}
                            className="bg-blue-600 hover:bg-blue-500 text-white px-3 py-1.5 rounded-lg font-bold text-[11px] flex items-center justify-center space-x-1 mx-auto transition-colors shadow-sm"
                          >
                            {repaymentLoading === order.id ? (
                              <RefreshCw className="w-3 h-3 animate-spin" />
                            ) : (
                              <CreditCard className="w-3 h-3" />
                            )}
                            <span>Bayar Sekarang (Repay)</span>
                          </button>
                        ) : (
                          <span className="text-[11px] text-gray-500 font-medium">-</span>
                        )}
                      </td>
                      <td className="p-4 text-gray-400 text-right">
                        {new Date(order.created_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
