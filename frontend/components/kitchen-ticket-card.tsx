'use client';

import React, { useState, useEffect } from 'react';
import { Order } from '../types';
import { Clock, CheckCircle, ChefHat, ArrowRight } from 'lucide-react';
import { formatIDR } from '@/utils/format';

interface KitchenTicketCardProps {
  order: Order;
  onUpdateStatus: (orderId: string, newStatus: string) => void;
}

export default function KitchenTicketCard({ order, onUpdateStatus }: KitchenTicketCardProps) {
  const [elapsedMinutes, setElapsedMinutes] = useState<number>(0);

  useEffect(() => {
    const calculateElapsed = () => {
      const created = new Date(order.created_at).getTime();
      const now = new Date().getTime();
      const diffMins = Math.floor((now - created) / 60000);
      setElapsedMinutes(diffMins >= 0 ? diffMins : 0);
    };

    calculateElapsed();
    const interval = setInterval(calculateElapsed, 10000);
    return () => clearInterval(interval);
  }, [order.created_at]);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'PENDING':
        return <span className="bg-amber-100 text-amber-800 px-2.5 py-1 rounded-full text-xs font-bold animate-pulse">PENDING</span>;
      case 'PREPARING':
        return <span className="bg-blue-100 text-blue-800 px-2.5 py-1 rounded-full text-xs font-bold flex items-center gap-1"><ChefHat className="w-3.5 h-3.5" /> PREPARING</span>;
      case 'READY':
        return <span className="bg-emerald-100 text-emerald-800 px-2.5 py-1 rounded-full text-xs font-bold flex items-center gap-1"><CheckCircle className="w-3.5 h-3.5" /> READY</span>;
      case 'SERVED':
      case 'COMPLETED':
        return <span className="bg-gray-100 text-gray-800 px-2.5 py-1 rounded-full text-xs font-bold">SERVED</span>;
      default:
        return <span className="bg-gray-100 text-gray-800 px-2.5 py-1 rounded-full text-xs font-bold">{status}</span>;
    }
  };

  return (
    <div className={`bg-white rounded-xl shadow-md border-2 p-5 flex flex-col justify-between transition-all ${order.status === 'PENDING' ? 'border-amber-400 bg-amber-50/20' :
      order.status === 'PREPARING' ? 'border-blue-400 bg-blue-50/20' :
        order.status === 'READY' ? 'border-emerald-400 bg-emerald-50/20' : 'border-gray-200'
      }`}>
      <div>
        <div className="flex justify-between items-start mb-3">
          <div>
            <h3 className="text-lg font-bold text-gray-900">Order #{order.id.slice(0, 8)}</h3>
            <p className="text-xs text-gray-500 font-mono">Tx: {order.transaction_id || 'POS-ORDER'}</p>
          </div>
          {getStatusBadge(order.status)}
        </div>

        <div className="flex items-center gap-2 text-xs text-gray-600 mb-4 bg-gray-50 p-2 rounded-lg">
          <Clock className="w-4 h-4 text-gray-400" />
          <span>Created {new Date(order.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
          <span className="ml-auto font-semibold text-gray-700">({elapsedMinutes}m ago)</span>
        </div>

        <div className="space-y-2 mb-4">
          <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider">Order Items</p>
          <div className="divide-y divide-gray-100 max-h-48 overflow-y-auto">
            {order.items && order.items.map((item, idx) => (
              <div key={item.id || idx} className="py-2 flex justify-between items-center text-sm">
                <span className="font-medium text-gray-800">
                  <span className="bg-gray-200 text-gray-800 px-2 py-0.5 rounded font-bold text-xs mr-2">{item.quantity}x</span>
                  Product ID: {item.product_id.slice(0, 8)}...
                </span>
                <span className="text-gray-600 font-mono">{formatIDR(item.unit_price * item.quantity)}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="pt-3 border-t border-gray-100 flex gap-2">
        {order.status === 'PENDING' && (
          <button
            onClick={() => onUpdateStatus(order.id, 'PREPARING')}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2.5 px-4 rounded-lg font-semibold text-sm flex items-center justify-center gap-2 shadow-sm transition-colors"
          >
            Start Preparing <ArrowRight className="w-4 h-4" />
          </button>
        )}
        {order.status === 'PREPARING' && (
          <button
            onClick={() => onUpdateStatus(order.id, 'READY')}
            className="w-full bg-emerald-600 hover:bg-emerald-700 text-white py-2.5 px-4 rounded-lg font-semibold text-sm flex items-center justify-center gap-2 shadow-sm transition-colors"
          >
            Mark Ready <CheckCircle className="w-4 h-4" />
          </button>
        )}
        {order.status === 'READY' && (
          <button
            onClick={() => onUpdateStatus(order.id, 'SERVED')}
            className="w-full bg-gray-800 hover:bg-gray-900 text-white py-2.5 px-4 rounded-lg font-semibold text-sm flex items-center justify-center gap-2 shadow-sm transition-colors"
          >
            Mark Served / Complete
          </button>
        )}
        {order.status === 'SERVED' && (
          <div className="w-full text-center py-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">
            Order Completed
          </div>
        )}
      </div>
    </div>
  );
}
