'use client';

import { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Store, ShoppingBag, BarChart3, Receipt } from 'lucide-react';
import './globals.css';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      })
  );

  const pathname = usePathname();

  const navLinks = [
    { href: '/checkout', label: 'POS Terminal', icon: ShoppingBag },
    { href: '/dashboard', label: 'Analytics Dashboard', icon: BarChart3 },
    { href: '/receipts', label: 'AI Receipt Auditor', icon: Receipt },
  ];

  return (
    <html lang="en">
      <head>
        <title>Smart AI POS & Merchant Engine</title>
        <meta
          name="description"
          content="Real-time synchronized POS checkout register with AI Vision receipt auditing and pessimistic stock controls."
        />
      </head>
      <body>
        <QueryClientProvider client={queryClient}>
          <div className="min-h-screen flex flex-col">
            {/* Global Header Navigation */}
            <header className="glass-panel border-b border-gray-700/50 sticky top-0 z-40">
              <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
                <Link href="/checkout" className="flex items-center space-x-3 group">
                  <div className="w-9 h-9 rounded-xl bg-brand-500/20 text-brand-500 flex items-center justify-center border border-brand-500/30 group-hover:scale-105 transition-transform">
                    <Store className="w-5 h-5" />
                  </div>
                  <div>
                    <span className="text-base font-black tracking-tight text-white flex items-center space-x-1.5">
                      <span>Smart AI POS</span>
                      <span className="text-[10px] font-bold text-brand-500 bg-brand-500/10 px-2 py-0.5 rounded-full border border-brand-500/20">
                        v1.0
                      </span>
                    </span>
                    <span className="text-[10px] text-gray-400 block -mt-0.5">
                      Merchant Engine & Real-time Sync
                    </span>
                  </div>
                </Link>

                {/* Navigation Links */}
                <nav className="flex items-center space-x-1 sm:space-x-2">
                  {navLinks.map((link) => {
                    const Icon = link.icon;
                    const isActive = pathname === link.href || (pathname === '/' && link.href === '/checkout');
                    return (
                      <Link
                        key={link.href}
                        href={link.href}
                        className={`flex items-center space-x-2 px-3.5 py-2 rounded-xl text-xs font-bold transition-all ${
                          isActive
                            ? 'bg-brand-500 text-black shadow-md shadow-brand-500/20'
                            : 'text-gray-300 hover:bg-dark-800 hover:text-white'
                        }`}
                      >
                        <Icon className="w-4 h-4" />
                        <span className="hidden sm:inline">{link.label}</span>
                      </Link>
                    );
                  })}
                </nav>
              </div>
            </header>

            {/* Main Page Content */}
            <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
              {children}
            </main>

            {/* Footer */}
            <footer className="border-t border-gray-800/80 py-4 text-center text-xs text-gray-500">
              Smart AI POS & Merchant Engine • Hexagonal Golang Backend + React Next.js Frontend
            </footer>
          </div>
        </QueryClientProvider>
      </body>
    </html>
  );
}
