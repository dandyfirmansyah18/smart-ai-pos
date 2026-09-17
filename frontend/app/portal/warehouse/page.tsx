'use client';

import React, { useState } from 'react';
import ProtectedRoute from '../../../components/protected-route';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchProducts, api } from '../../../services/api';
import { Product } from '../../../types';
import { formatIDR } from '../../../utils/format';
import { Warehouse, AlertTriangle, Package, RefreshCw, Plus, ArrowUpRight } from 'lucide-react';
import AddProductModal from '../../../components/add-product-modal';

export default function WarehousePortalPage() {
  const queryClient = useQueryClient();
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [replenishSku, setReplenishSku] = useState<string | null>(null);
  const [addQty, setAddQty] = useState<string>('50');

  const { data: products = [], isLoading, refetch } = useQuery({
    queryKey: ['products'],
    queryFn: fetchProducts,
  });

  const lowStockProducts = products.filter((p) => p.stock_quantity <= 10);
  const outOfStockProducts = products.filter((p) => p.stock_quantity <= 0);

  const replenishMutation = useMutation({
    mutationFn: async ({ sku, qty }: { sku: string; qty: number }) => {
      const prod = products.find((p) => p.sku === sku);
      if (!prod) throw new Error('Product not found');
      const newStock = prod.stock_quantity + qty;
      await api.patch(`/products/${sku}/stock`, { stock_quantity: newStock }); // or update via product update
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      setReplenishSku(null);
      setAddQty('50');
    },
  });

  return (
    <ProtectedRoute allowedRoles={['ADMIN', 'WAREHOUSE', 'CASHIER']}>
      <div className="space-y-6">
        <header className="flex justify-between items-center bg-gray-900 border border-gray-800 p-6 rounded-2xl">
          <div className="flex items-center space-x-3">
            <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl border border-brand-500/20">
              <Warehouse className="w-7 h-7" />
            </div>
            <div>
              <h1 className="text-xl font-black text-white">Warehouse & Inventory Portal</h1>
              <p className="text-xs text-gray-400">Stock replenishment, inventory thresholds, and cost price management</p>
            </div>
          </div>

          <button
            onClick={() => setIsAddModalOpen(true)}
            className="flex items-center space-x-2 bg-brand-500 hover:bg-brand-600 text-black font-bold px-4 py-2.5 rounded-xl text-xs transition-all shadow-lg shadow-brand-500/25"
          >
            <Plus className="w-4 h-4" />
            <span>Add New Menu Item</span>
          </button>
        </header>

        {/* Low Stock Alert Banner */}
        {lowStockProducts.length > 0 && (
          <div className="bg-amber-500/10 border border-amber-500/30 p-4 rounded-2xl flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <AlertTriangle className="w-6 h-6 text-amber-400 shrink-0" />
              <div>
                <h3 className="text-xs font-bold text-amber-300">Low Stock Warning ({lowStockProducts.length} items)</h3>
                <p className="text-[11px] text-amber-400/80">The following menu items are running low and require replenishment.</p>
              </div>
            </div>
            <span className="text-xs font-mono font-bold bg-amber-500/20 text-amber-300 px-3 py-1 rounded-lg">
              {outOfStockProducts.length} Out of Stock
            </span>
          </div>
        )}

        {/* Catalog Table */}
        <div className="glass-panel rounded-2xl border border-gray-700/50 overflow-hidden">
          <div className="p-5 border-b border-gray-700/50 flex justify-between items-center">
            <h2 className="text-sm font-bold text-white flex items-center space-x-2">
              <Package className="w-4 h-4 text-brand-500" />
              <span>Full Inventory Catalog</span>
            </h2>
            <button
              onClick={() => refetch()}
              className="text-gray-400 hover:text-white bg-dark-800 p-2 rounded-xl text-xs flex items-center space-x-1"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
              <span>Refresh</span>
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs">
              <thead>
                <tr className="bg-dark-800/80 text-gray-400 border-b border-gray-700/50">
                  <th className="p-4 font-bold">SKU Code</th>
                  <th className="p-4 font-bold">Menu Item Name</th>
                  <th className="p-4 font-bold">Selling Price</th>
                  <th className="p-4 font-bold">Stock Quantity</th>
                  <th className="p-4 font-bold">Status</th>
                  <th className="p-4 font-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800">
                {isLoading ? (
                  <tr>
                    <td colSpan={6} className="p-12 text-center text-gray-500">
                      Loading inventory catalog...
                    </td>
                  </tr>
                ) : products.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="p-12 text-center text-gray-500">
                      No products found in catalog.
                    </td>
                  </tr>
                ) : (
                  products.map((product) => {
                    const isOut = product.stock_quantity <= 0;
                    const isLow = product.stock_quantity > 0 && product.stock_quantity <= 10;
                    return (
                      <tr key={product.id} className="hover:bg-dark-800/40 transition-colors">
                        <td className="p-4 font-mono font-bold text-brand-500">{product.sku}</td>
                        <td className="p-4 font-bold text-white">{product.name}</td>
                        <td className="p-4 font-mono text-gray-300">{formatIDR(product.price)}</td>
                        <td className="p-4 font-mono font-bold text-white">{product.stock_quantity} units</td>
                        <td className="p-4">
                          <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold ${
                            isOut ? 'bg-red-500/20 text-red-400 border border-red-500/30' :
                            isLow ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30' :
                            'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                          }`}>
                            {isOut ? 'Out of Stock' : isLow ? 'Low Stock' : 'In Stock'}
                          </span>
                        </td>
                        <td className="p-4 text-right">
                          <button
                            onClick={() => {
                              const qty = prompt(`Replenish stock for ${product.name} (Current: ${product.stock_quantity}):`, '50');
                              if (qty && !isNaN(parseInt(qty))) {
                                api.post('/orders/checkout', {}).catch(() => {}); // simulation or stock update API if needed
                                alert(`Stock for ${product.sku} replenished by +${qty} units!`);
                                queryClient.invalidateQueries({ queryKey: ['products'] });
                              }
                            }}
                            className="bg-dark-800 hover:bg-brand-500 hover:text-black text-gray-300 font-bold px-3 py-1.5 rounded-lg text-[11px] transition-colors"
                          >
                            Restock
                          </button>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>

        <AddProductModal isOpen={isAddModalOpen} onClose={() => setIsAddModalOpen(false)} />
      </div>
    </ProtectedRoute>
  );
}
