'use client';

import React from 'react';
import ProtectedRoute from '../../../components/protected-route';
import { UserRole } from '../../../types';
import { useQuery } from '@tanstack/react-query';
import { fetchProducts } from '../../../services/api';
import { formatIDR } from '../../../utils/format';
import { TrendingUp, DollarSign, PieChart, Award, RefreshCw, BarChart2 } from 'lucide-react';

export default function FinancePortalPage() {
  const { data: products = [], isLoading, refetch } = useQuery({
    queryKey: ['products'],
    queryFn: fetchProducts,
  });

  const totalCatalogValue = products.reduce((sum, p) => sum + p.price * p.stock_quantity, 0);
  // Estimated COGS (approx 55% of selling price for F&B)
  const estimatedCOGS = totalCatalogValue * 0.55;
  const grossProfit = totalCatalogValue - estimatedCOGS;
  const profitMarginPercent = totalCatalogValue > 0 ? (grossProfit / totalCatalogValue) * 100 : 0;

  return (
    <ProtectedRoute allowedRoles={[UserRole.ADMIN, UserRole.CASHIER]}>
      <div className="space-y-6">
        <header className="flex justify-between items-center bg-gray-900 border border-gray-800 p-6 rounded-2xl">
          <div className="flex items-center space-x-3">
            <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl border border-brand-500/20">
              <TrendingUp className="w-7 h-7" />
            </div>
            <div>
              <h1 className="text-xl font-black text-white">Financial Profit & Loss (Laba Rugi) Portal</h1>
              <p className="text-xs text-gray-400">COGS (HPP), gross profit margins, and revenue analytics</p>
            </div>
          </div>

          <button
            onClick={() => refetch()}
            className="flex items-center space-x-2 bg-dark-800 hover:bg-dark-700 text-gray-200 px-4 py-2.5 rounded-xl text-xs font-bold transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            <span>Refresh Financials</span>
          </button>
        </header>

        {/* Financial KPI Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
            <div className="flex justify-between items-center mb-3">
              <span className="text-xs font-semibold text-gray-400">Est. Gross Revenue</span>
              <div className="p-2 bg-brand-500/10 text-brand-500 rounded-xl">
                <DollarSign className="w-5 h-5" />
              </div>
            </div>
            <span className="text-2xl font-black text-white">{formatIDR(totalCatalogValue)}</span>
            <span className="text-[10px] text-gray-400 block mt-1">Potential inventory turnover value</span>
          </div>

          <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
            <div className="flex justify-between items-center mb-3">
              <span className="text-xs font-semibold text-gray-400">Est. Total COGS (HPP)</span>
              <div className="p-2 bg-amber-500/10 text-amber-400 rounded-xl">
                <PieChart className="w-5 h-5" />
              </div>
            </div>
            <span className="text-2xl font-black text-amber-400">{formatIDR(estimatedCOGS)}</span>
            <span className="text-[10px] text-gray-400 block mt-1">Cost of Goods Sold (~55%)</span>
          </div>

          <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
            <div className="flex justify-between items-center mb-3">
              <span className="text-xs font-semibold text-gray-400">Gross Profit (Laba Kotor)</span>
              <div className="p-2 bg-emerald-500/10 text-emerald-400 rounded-xl">
                <TrendingUp className="w-5 h-5" />
              </div>
            </div>
            <span className="text-2xl font-black text-emerald-400">{formatIDR(grossProfit)}</span>
            <span className="text-[10px] text-gray-400 block mt-1">Revenue minus COGS</span>
          </div>

          <div className="glass-panel p-5 rounded-2xl border border-gray-700/50">
            <div className="flex justify-between items-center mb-3">
              <span className="text-xs font-semibold text-gray-400">Profit Margin %</span>
              <div className="p-2 bg-blue-500/10 text-blue-400 rounded-xl">
                <BarChart2 className="w-5 h-5" />
              </div>
            </div>
            <span className="text-2xl font-black text-blue-400">{profitMarginPercent.toFixed(1)}%</span>
            <span className="text-[10px] text-gray-400 block mt-1">Average catalog margin</span>
          </div>
        </div>

        {/* Top Profitable Items Table */}
        <div className="glass-panel rounded-2xl border border-gray-700/50 p-6">
          <div className="flex items-center space-x-2 mb-6">
            <Award className="w-5 h-5 text-brand-500" />
            <h2 className="text-base font-bold text-white">Top Profitable Menu Items (Margin Ranking)</h2>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="bg-dark-800/80 text-gray-400 border-b border-gray-700/50">
                  <th className="p-4 font-bold">Rank</th>
                  <th className="p-4 font-bold">SKU</th>
                  <th className="p-4 font-bold">Menu Item Name</th>
                  <th className="p-4 font-bold">Selling Price</th>
                  <th className="p-4 font-bold">Est. Cost (HPP)</th>
                  <th className="p-4 font-right">Est. Profit Margin</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800">
                {products.slice(0, 10).map((product, idx) => {
                  const cogs = product.price * 0.55;
                  const margin = product.price > 0 ? ((product.price - cogs) / product.price) * 100 : 0;
                  return (
                    <tr key={product.id} className="hover:bg-dark-800/40 transition-colors">
                      <td className="p-4 font-black text-brand-500">#{idx + 1}</td>
                      <td className="p-4 font-mono text-gray-300">{product.sku}</td>
                      <td className="p-4 font-bold text-white">{product.name}</td>
                      <td className="p-4 font-mono text-gray-200">{formatIDR(product.price)}</td>
                      <td className="p-4 font-mono text-amber-400">{formatIDR(cogs)}</td>
                      <td className="p-4 text-right font-mono font-bold text-emerald-400">{margin.toFixed(1)}%</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
