'use client';

import { useQuery } from '@tanstack/react-query';
import { fetchProducts } from '../services/api';
import {
  TrendingUp,
  DollarSign,
  ShoppingBag,
  AlertTriangle,
  PackageCheck,
  RefreshCw,
} from 'lucide-react';

export function SalesChart() {
  const { data: products = [], isLoading } = useQuery({
    queryKey: ['products'],
    queryFn: fetchProducts,
  });

  const totalCatalogItems = products.length;
  const outOfStockItems = products.filter((p) => p.stock_quantity <= 0).length;
  const lowStockItems = products.filter((p) => p.stock_quantity > 0 && p.stock_quantity <= 10).length;
  const totalStockUnits = products.reduce((sum, p) => sum + p.stock_quantity, 0);
  const totalCatalogValue = products.reduce((sum, p) => sum + p.price * p.stock_quantity, 0);

  if (isLoading) {
    return (
      <div className="glass-panel p-12 rounded-2xl flex items-center justify-center">
        <RefreshCw className="w-8 h-8 text-brand-500 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Metric Cards Top Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-gray-400">Total Catalog Value</span>
            <div className="p-2 bg-brand-500/10 text-brand-500 rounded-xl">
              <DollarSign className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-black text-white">
              ${totalCatalogValue.toFixed(2)}
            </span>
            <span className="text-[10px] text-gray-400 block mt-0.5">
              Based on active product inventory
            </span>
          </div>
        </div>

        <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-gray-400">Active SKUs</span>
            <div className="p-2 bg-blue-500/10 text-blue-400 rounded-xl">
              <ShoppingBag className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-black text-white">{totalCatalogItems}</span>
            <span className="text-[10px] text-gray-400 block mt-0.5">
              {totalStockUnits} total units in stock
            </span>
          </div>
        </div>

        <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-gray-400">Low Stock Alert</span>
            <div className="p-2 bg-amber-500/10 text-amber-400 rounded-xl">
              <AlertTriangle className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-black text-amber-400">{lowStockItems}</span>
            <span className="text-[10px] text-gray-400 block mt-0.5">
              Items with 10 or fewer units
            </span>
          </div>
        </div>

        <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-gray-400">Out of Stock</span>
            <div className="p-2 bg-red-500/10 text-red-400 rounded-xl">
              <PackageCheck className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            <span className="text-2xl font-black text-red-400">{outOfStockItems}</span>
            <span className="text-[10px] text-gray-400 block mt-0.5">
              Requires immediate reorder
            </span>
          </div>
        </div>
      </div>

      {/* Visual Stock Level Progress Bars */}
      <div className="glass-panel p-6 rounded-2xl border border-gray-700/50">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h3 className="text-base font-bold text-gray-100 flex items-center space-x-2">
              <TrendingUp className="w-5 h-5 text-brand-500" />
              <span>Real-time Inventory Stock Levels</span>
            </h3>
            <p className="text-xs text-gray-400 mt-1">
              Live stock metrics auto-updated via WebSocket push events
            </p>
          </div>
        </div>

        <div className="space-y-4">
          {products.map((product) => {
            const percentage = Math.min(100, Math.max(0, (product.stock_quantity / 100) * 100));
            const isLow = product.stock_quantity <= 10;
            const isOut = product.stock_quantity <= 0;

            return (
              <div key={product.id} className="space-y-1.5">
                <div className="flex justify-between text-xs">
                  <span className="font-semibold text-gray-200">{product.name} ({product.sku})</span>
                  <span
                    className={`font-bold ${
                      isOut ? 'text-red-400' : isLow ? 'text-amber-400' : 'text-brand-500'
                    }`}
                  >
                    {product.stock_quantity} units
                  </span>
                </div>

                <div className="w-full h-2.5 bg-dark-800 rounded-full overflow-hidden border border-gray-700/40">
                  <div
                    className={`h-full rounded-full transition-all duration-500 ${
                      isOut
                        ? 'bg-red-500'
                        : isLow
                        ? 'bg-amber-500'
                        : 'bg-gradient-to-r from-brand-600 to-brand-500'
                    }`}
                    style={{ width: `${percentage}%` }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
