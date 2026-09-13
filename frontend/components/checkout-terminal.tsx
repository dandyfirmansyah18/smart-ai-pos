'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchProducts, checkoutOrder } from '../services/api';
import { Product, CartItem, Order } from '../types';
import { useWebSocketSync } from '../hooks/use-websocket';
import {
  ShoppingBag,
  Plus,
  Minus,
  Trash2,
  CheckCircle2,
  AlertCircle,
  RefreshCw,
  Search,
  Zap,
  Coffee,
  Package,
} from 'lucide-react';

export function CheckoutTerminal() {
  const queryClient = useQueryClient();
  const { isConnected, lastEvent } = useWebSocketSync();

  const [cart, setCart] = useState<CartItem[]>([]);
  const [search, setSearch] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('ALL');
  const [lastCompletedOrder, setLastCompletedOrder] = useState<Order | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Fetch products using TanStack Query
  const {
    data: products = [],
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['products'],
    queryFn: fetchProducts,
  });

  // Checkout Mutation
  const checkoutMutation = useMutation({
    mutationFn: checkoutOrder,
    onSuccess: (completedOrder) => {
      setLastCompletedOrder(completedOrder);
      setCart([]);
      setErrorMessage(null);
      // Invalidate products to ensure server state alignment
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
    onError: (err: any) => {
      const msg = err.response?.data?.error || err.message || 'Checkout failed';
      setErrorMessage(msg);
    },
  });

  const addToCart = (product: Product) => {
    if (product.stock_quantity <= 0) return;

    setCart((prev) => {
      const existing = prev.find((item) => item.product.id === product.id);
      if (existing) {
        if (existing.quantity >= product.stock_quantity) return prev;
        return prev.map((item) =>
          item.product.id === product.id ? { ...item, quantity: item.quantity + 1 } : item
        );
      }
      return [...prev, { product, quantity: 1 }];
    });
  };

  const updateQuantity = (productId: string, delta: number) => {
    setCart((prev) =>
      prev
        .map((item) => {
          if (item.product.id === productId) {
            const newQty = item.quantity + delta;
            if (newQty > item.product.stock_quantity) return item;
            return { ...item, quantity: newQty };
          }
          return item;
        })
        .filter((item) => item.quantity > 0)
    );
  };

  const removeFromCart = (productId: string) => {
    setCart((prev) => prev.filter((item) => item.product.id !== productId));
  };

  const subtotal = cart.reduce((sum, item) => sum + item.product.price * item.quantity, 0);
  const tax = subtotal * 0.08;
  const total = subtotal + tax;

  const handleCheckout = () => {
    if (cart.length === 0) return;

    const idempotencyKey = `IDEM-POS-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`;
    const payload = {
      idempotency_key: idempotencyKey,
      items: cart.map((item) => ({
        sku: item.product.sku,
        quantity: item.quantity,
      })),
    };

    checkoutMutation.mutate(payload);
  };

  const filteredProducts = products.filter((p) => {
    const matchesSearch =
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase());
    const matchesCategory =
      selectedCategory === 'ALL' ||
      (selectedCategory === 'COFFEE' && p.sku.includes('COFFEE')) ||
      (selectedCategory === 'FOOD' && p.sku.includes('FOOD')) ||
      (selectedCategory === 'DRINK' && p.sku.includes('DRINK'));
    return matchesSearch && matchesCategory;
  });

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 min-h-[calc(100vh-6rem)]">
      {/* Catalog Grid Section (8 Cols) */}
      <div className="lg:col-span-8 flex flex-col space-y-6">
        {/* Top Filter Bar */}
        <div className="glass-panel p-4 rounded-2xl flex flex-wrap items-center justify-between gap-4">
          <div className="relative flex-1 min-w-[240px]">
            <Search className="absolute left-3.5 top-3 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search product name or SKU..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full bg-dark-800 border border-gray-700/60 rounded-xl pl-10 pr-4 py-2.5 text-sm text-gray-100 placeholder-gray-400 focus:outline-none focus:border-brand-500 transition-colors"
            />
          </div>

          <div className="flex items-center space-x-2 overflow-x-auto pb-1 sm:pb-0">
            {['ALL', 'COFFEE', 'FOOD', 'DRINK'].map((cat) => (
              <button
                key={cat}
                onClick={() => setSelectedCategory(cat)}
                className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold tracking-wide transition-all ${
                  selectedCategory === cat
                    ? 'bg-brand-500 text-black shadow-lg shadow-brand-500/20'
                    : 'bg-dark-800 text-gray-300 hover:bg-dark-700 border border-gray-700/50'
                }`}
              >
                {cat}
              </button>
            ))}
          </div>

          {/* WebSocket Live Sync Badge */}
          <div className="flex items-center space-x-2 bg-dark-800 border border-gray-700/50 px-3 py-1.5 rounded-xl text-xs">
            <div
              className={`w-2 h-2 rounded-full ${
                isConnected ? 'bg-brand-500 animate-pulse' : 'bg-red-500'
              }`}
            />
            <span className="text-gray-300 font-medium">
              {isConnected ? 'WS Live Sync' : 'Connecting...'}
            </span>
          </div>
        </div>

        {/* Live Event Toast Banner */}
        {lastEvent && (
          <div className="bg-brand-500/10 border border-brand-500/30 text-brand-500 px-4 py-2.5 rounded-xl text-xs flex items-center justify-between animate-fadeIn">
            <div className="flex items-center space-x-2">
              <Zap className="w-4 h-4 text-brand-500" />
              <span>
                Real-time Stock Update: <strong className="font-bold">{lastEvent.sku}</strong> new
                stock is <strong className="font-bold">{lastEvent.new_stock}</strong> units.
              </span>
            </div>
            <span className="text-gray-400 text-[10px]">Just now</span>
          </div>
        )}

        {/* Product Catalog Cards */}
        {isLoading ? (
          <div className="flex-1 glass-panel rounded-2xl flex items-center justify-center p-12">
            <RefreshCw className="w-8 h-8 text-brand-500 animate-spin" />
          </div>
        ) : isError ? (
          <div className="flex-1 glass-panel rounded-2xl flex flex-col items-center justify-center p-12 text-center">
            <AlertCircle className="w-10 h-10 text-red-500 mb-3" />
            <h3 className="text-lg font-bold text-gray-200">Failed to connect to backend server</h3>
            <p className="text-sm text-gray-400 mt-1 mb-4">
              Ensure Golang REST API server is running on http://localhost:8080
            </p>
            <button
              onClick={() => refetch()}
              className="bg-brand-500 text-black px-4 py-2 rounded-xl text-xs font-bold hover:bg-brand-600 transition-colors"
            >
              Retry Connection
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 flex-1">
            {filteredProducts.map((product) => {
              const isOutOfStock = product.stock_quantity <= 0;
              const cartQuantity =
                cart.find((item) => item.product.id === product.id)?.quantity || 0;

              return (
                <div
                  key={product.id}
                  className={`glass-card p-5 rounded-2xl flex flex-col justify-between transition-all hover:border-brand-500/40 group ${
                    isOutOfStock ? 'opacity-50' : ''
                  }`}
                >
                  <div>
                    <div className="flex items-start justify-between mb-3">
                      <span className="text-[10px] font-bold tracking-wider text-brand-500 bg-brand-500/10 px-2.5 py-1 rounded-lg border border-brand-500/20">
                        {product.sku}
                      </span>
                      <span
                        className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${
                          isOutOfStock
                            ? 'bg-red-500/20 text-red-400 border border-red-500/30'
                            : 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                        }`}
                      >
                        {isOutOfStock ? 'Out of Stock' : `${product.stock_quantity} left`}
                      </span>
                    </div>

                    <h3 className="text-base font-bold text-gray-100 group-hover:text-brand-500 transition-colors">
                      {product.name}
                    </h3>
                    <p className="text-xs text-gray-400 mt-1 line-clamp-2">
                      {product.description || 'Premium store catalog item.'}
                    </p>
                  </div>

                  <div className="mt-4 pt-4 border-t border-gray-700/50 flex items-center justify-between">
                    <div className="text-lg font-black text-white">
                      ${product.price.toFixed(2)}
                    </div>

                    <button
                      onClick={() => addToCart(product)}
                      disabled={isOutOfStock}
                      className={`flex items-center space-x-1.5 px-3.5 py-2 rounded-xl text-xs font-bold transition-all ${
                        isOutOfStock
                          ? 'bg-gray-800 text-gray-500 cursor-not-allowed'
                          : 'bg-brand-500 text-black hover:bg-brand-600 shadow-md shadow-brand-500/20 active:scale-95'
                      }`}
                    >
                      <Plus className="w-3.5 h-3.5" />
                      <span>{cartQuantity > 0 ? `Added (${cartQuantity})` : 'Add'}</span>
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Cart Summary Sidebar (4 Cols) */}
      <div className="lg:col-span-4 glass-panel rounded-2xl p-6 flex flex-col justify-between border border-gray-700/50">
        <div>
          <div className="flex items-center justify-between pb-4 border-b border-gray-700/50 mb-4">
            <div className="flex items-center space-x-2">
              <ShoppingBag className="w-5 h-5 text-brand-500" />
              <h2 className="text-base font-bold text-gray-100">Current Order Cart</h2>
            </div>
            <span className="text-xs font-semibold text-gray-400 bg-dark-800 px-2.5 py-1 rounded-lg">
              {cart.reduce((s, i) => s + i.quantity, 0)} items
            </span>
          </div>

          {/* Error Alert */}
          {errorMessage && (
            <div className="bg-red-500/10 border border-red-500/30 text-red-400 p-3 rounded-xl text-xs mb-4 flex items-start space-x-2">
              <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* Cart Item List */}
          {cart.length === 0 ? (
            <div className="py-16 text-center text-gray-500 flex flex-col items-center">
              <Package className="w-12 h-12 text-gray-600 stroke-[1.5] mb-2" />
              <p className="text-sm font-medium">Cart is empty</p>
              <p className="text-xs text-gray-500 mt-1">
                Select items from catalog to start POS checkout
              </p>
            </div>
          ) : (
            <div className="space-y-3 max-h-[380px] overflow-y-auto pr-1">
              {cart.map((item) => (
                <div
                  key={item.product.id}
                  className="bg-dark-800/80 p-3.5 rounded-xl border border-gray-700/40 flex items-center justify-between"
                >
                  <div className="flex-1 pr-2">
                    <h4 className="text-xs font-bold text-gray-200">{item.product.name}</h4>
                    <span className="text-[10px] text-gray-400">
                      ${item.product.price.toFixed(2)} each
                    </span>
                  </div>

                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => updateQuantity(item.product.id, -1)}
                      className="w-6 h-6 rounded-lg bg-dark-700 text-gray-300 hover:bg-dark-600 flex items-center justify-center transition-colors"
                    >
                      <Minus className="w-3 h-3" />
                    </button>
                    <span className="text-xs font-bold text-white w-5 text-center">
                      {item.quantity}
                    </span>
                    <button
                      onClick={() => updateQuantity(item.product.id, 1)}
                      className="w-6 h-6 rounded-lg bg-dark-700 text-gray-300 hover:bg-dark-600 flex items-center justify-center transition-colors"
                    >
                      <Plus className="w-3 h-3" />
                    </button>
                    <button
                      onClick={() => removeFromCart(item.product.id)}
                      className="w-6 h-6 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-500/10 flex items-center justify-center ml-1 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Pricing Summary & Checkout Button */}
        <div className="pt-4 border-t border-gray-700/50 mt-6 space-y-3">
          <div className="flex justify-between text-xs text-gray-400">
            <span>Subtotal</span>
            <span>${subtotal.toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-xs text-gray-400">
            <span>Est. Sales Tax (8%)</span>
            <span>${tax.toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-base font-black text-white pt-2 border-t border-gray-700/40">
            <span>Total Pay</span>
            <span className="text-brand-500">${total.toFixed(2)}</span>
          </div>

          <button
            onClick={handleCheckout}
            disabled={cart.length === 0 || checkoutMutation.isPending}
            className={`w-full py-3.5 rounded-xl font-bold text-sm flex items-center justify-center space-x-2 transition-all ${
              cart.length === 0 || checkoutMutation.isPending
                ? 'bg-gray-800 text-gray-500 cursor-not-allowed'
                : 'bg-brand-500 text-black hover:bg-brand-600 shadow-xl shadow-brand-500/25 active:scale-[0.98]'
            }`}
          >
            {checkoutMutation.isPending ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" />
                <span>Processing Order...</span>
              </>
            ) : (
              <>
                <CheckCircle2 className="w-4 h-4" />
                <span>Complete Checkout</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Checkout Success Modal Overlay */}
      {lastCompletedOrder && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="glass-panel border border-brand-500/40 rounded-2xl max-w-md w-full p-6 text-center shadow-2xl animate-scaleUp">
            <div className="w-14 h-14 bg-brand-500/20 text-brand-500 rounded-full flex items-center justify-center mx-auto mb-4 border border-brand-500/30">
              <CheckCircle2 className="w-8 h-8" />
            </div>
            <h3 className="text-xl font-black text-white">Checkout Completed!</h3>
            <p className="text-xs text-gray-400 mt-1">
              Transaction ID:{' '}
              <strong className="text-brand-500 font-mono">
                {lastCompletedOrder.transaction_id}
              </strong>
            </p>

            <div className="bg-dark-800/80 p-4 rounded-xl border border-gray-700/50 my-4 text-left text-xs space-y-2">
              <div className="flex justify-between">
                <span className="text-gray-400">Total Paid:</span>
                <span className="font-bold text-white">
                  ${lastCompletedOrder.total_amount.toFixed(2)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Idempotency Key:</span>
                <span className="font-mono text-gray-300">
                  {lastCompletedOrder.idempotency_key}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Status:</span>
                <span className="font-bold text-brand-500">{lastCompletedOrder.status}</span>
              </div>
            </div>

            <button
              onClick={() => setLastCompletedOrder(null)}
              className="w-full bg-brand-500 text-black py-2.5 rounded-xl font-bold text-xs hover:bg-brand-600 transition-colors"
            >
              Start New Order
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
