'use client';

import React, { useState } from 'react';
import { api } from '../services/api';
import { useQueryClient } from '@tanstack/react-query';
import { X, PlusCircle, CheckCircle2, AlertCircle, RefreshCw } from 'lucide-react';

interface AddProductModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function AddProductModal({ isOpen, onClose }: AddProductModalProps) {
  const queryClient = useQueryClient();
  const [sku, setSku] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [stockQuantity, setStockQuantity] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const numericPrice = parseFloat(price);
      const numericStock = parseInt(stockQuantity, 10);

      if (!sku || !name || isNaN(numericPrice) || isNaN(numericStock)) {
        throw new Error('Please fill in all required fields with valid values.');
      }

      await api.post('/products', {
        sku,
        name,
        description,
        price: numericPrice,
        stock_quantity: numericStock,
      });

      setSuccess(true);
      queryClient.invalidateQueries({ queryKey: ['products'] });
      setTimeout(() => {
        setSuccess(false);
        setSku('');
        setName('');
        setDescription('');
        setPrice('');
        setStockQuantity('');
        onClose();
      }, 1200);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to create product');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="glass-panel border border-gray-700 rounded-2xl max-w-lg w-full p-6 shadow-2xl relative animate-scaleUp">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-white bg-dark-800 p-2 rounded-xl"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="flex items-center space-x-3 mb-6">
          <div className="p-3 bg-brand-500/10 text-brand-500 rounded-xl border border-brand-500/25">
            <PlusCircle className="w-6 h-6" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Add New Menu / Product</h2>
            <p className="text-xs text-gray-400">Register a new item into the Harmoni Cafe & Resto catalog</p>
          </div>
        </div>

        {success ? (
          <div className="py-12 text-center text-emerald-400 flex flex-col items-center">
            <CheckCircle2 className="w-12 h-12 mb-2 animate-bounce" />
            <p className="font-bold text-base">Menu item successfully added!</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="bg-red-500/10 border border-red-500/30 text-red-400 p-3 rounded-xl text-xs flex items-center space-x-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{error}</span>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">SKU Code *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. SKU-MENU-101"
                  value={sku}
                  onChange={(e) => setSku(e.target.value)}
                  className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">Menu Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Nasi Goreng Spesial"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500"
                />
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-gray-300 mb-1">Description</label>
              <textarea
                rows={2}
                placeholder="Delicious ingredients..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 resize-none"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">Price (IDR) *</label>
                <input
                  type="number"
                  step="1000"
                  required
                  placeholder="e.g. 45000"
                  value={price}
                  onChange={(e) => setPrice(e.target.value)}
                  className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 font-mono"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-gray-300 mb-1">Initial Stock Qty *</label>
                <input
                  type="number"
                  required
                  placeholder="e.g. 50"
                  value={stockQuantity}
                  onChange={(e) => setStockQuantity(e.target.value)}
                  className="w-full bg-dark-800 border border-gray-700 rounded-xl px-3.5 py-2.5 text-xs text-white focus:outline-none focus:border-brand-500 font-mono"
                />
              </div>
            </div>

            <div className="pt-4 flex space-x-3">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 bg-dark-800 hover:bg-dark-700 text-gray-300 font-bold py-3 rounded-xl text-xs transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="flex-1 bg-brand-500 hover:bg-brand-600 text-black font-bold py-3 rounded-xl text-xs flex items-center justify-center space-x-2 transition-all shadow-lg shadow-brand-500/20"
              >
                {loading ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    <span>Saving...</span>
                  </>
                ) : (
                  <span>Save Menu Item</span>
                )}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
