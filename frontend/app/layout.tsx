'use client';

import { useState } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Store, ShoppingBag, BarChart3, Receipt, LogIn, LogOut, User as UserIcon, ChefHat, Warehouse, TrendingUp, ShieldCheck, ChevronDown } from 'lucide-react';
import { AuthProvider, useAuth } from '@/context/auth-context';
import './globals.css';

function NavbarContent({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const isLoginPage = pathname === '/login';

  const primaryLinks = [
    { href: '/checkout', label: 'POS Terminal', icon: ShoppingBag },
  ];

  const portalLinks = [
    { href: '/portal/kitchen', label: 'Kitchen (KDS)', icon: ChefHat },
    { href: '/portal/warehouse', label: 'Warehouse', icon: Warehouse },
    { href: '/portal/finance', label: 'Finance P&L', icon: TrendingUp },
    { href: '/dashboard', label: 'Analytics', icon: BarChart3 },
    { href: '/receipts', label: 'Receipt Scanner', icon: Receipt },
  ];

  if (user?.role === 'ADMIN') {
    portalLinks.push({ href: '/portal/setup-role', label: 'Setup Role', icon: ShieldCheck });
  }

  const isPortalActive = portalLinks.some((l) => pathname === l.href);

  if (isLoginPage) {
    return <main className="flex-1 w-full">{children}</main>;
  }

  return (
    <div className="min-h-screen flex flex-col">
      {/* Global Header Navigation */}
      <header className="glass-panel border-b border-gray-700/50 sticky top-0 z-40 bg-gray-900/90 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link href="/checkout" className="flex items-center space-x-3 group">
            <div className="w-9 h-9 rounded-xl bg-blue-600/20 text-blue-400 flex items-center justify-center border border-blue-500/30 group-hover:scale-105 transition-transform">
              <Store className="w-5 h-5" />
            </div>
            <div>
              <span className="text-base font-black tracking-tight text-white flex items-center space-x-1.5">
                <span>Smart AI POS</span>
                <span className="text-[10px] font-bold text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded-full border border-blue-500/20">
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
            {primaryLinks.map((link) => {
              const Icon = link.icon;
              const isActive = pathname === link.href || (pathname === '/' && link.href === '/checkout');
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`flex items-center space-x-2 px-3.5 py-2 rounded-xl text-xs font-bold transition-all ${
                    isActive
                      ? 'bg-blue-600 text-white shadow-md shadow-blue-600/20'
                      : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  <span className="hidden sm:inline">{link.label}</span>
                </Link>
              );
            })}

            {/* Portals & Tools Dropdown */}
            <div className="relative">
              <button
                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                className={`flex items-center space-x-2 px-3.5 py-2 rounded-xl text-xs font-bold transition-all ${
                  isPortalActive
                    ? 'bg-blue-600 text-white shadow-md shadow-blue-600/20'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                }`}
              >
                <Store className="w-4 h-4" />
                <span className="hidden sm:inline">Portals & Tools</span>
                <ChevronDown className={`w-3.5 h-3.5 transition-transform ${isDropdownOpen ? 'rotate-180' : ''}`} />
              </button>

              {isDropdownOpen && (
                <div className="absolute right-0 mt-2 w-56 glass-panel rounded-2xl border border-gray-700 shadow-2xl py-2 z-50 bg-gray-900/95 backdrop-blur-xl">
                  {portalLinks.map((link) => {
                    const Icon = link.icon;
                    const isActive = pathname === link.href;
                    return (
                      <Link
                        key={link.href}
                        href={link.href}
                        onClick={() => setIsDropdownOpen(false)}
                        className={`flex items-center space-x-2 px-4 py-2.5 text-xs font-bold transition-colors ${
                          isActive
                            ? 'bg-blue-600/20 text-blue-400 font-black'
                            : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                        }`}
                      >
                        <Icon className="w-4 h-4 text-brand-500" />
                        <span>{link.label}</span>
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>

            {user ? (
              <div className="flex items-center space-x-3 ml-4 pl-4 border-l border-gray-700">
                <div className="hidden md:flex flex-col text-right">
                  <span className="text-xs font-bold text-white">{user.full_name}</span>
                  <span className="text-[10px] text-blue-400 uppercase font-semibold">{user.role}</span>
                </div>
                <button
                  onClick={logout}
                  title="Sign out"
                  className="flex items-center space-x-1.5 px-3 py-2 rounded-xl text-xs font-bold bg-red-500/10 text-red-400 border border-red-500/20 hover:bg-red-500/20 transition-all"
                >
                  <LogOut className="w-4 h-4" />
                  <span className="hidden sm:inline">Logout</span>
                </button>
              </div>
            ) : (
              <Link
                href="/login"
                className="flex items-center space-x-1.5 px-4 py-2 rounded-xl text-xs font-bold bg-blue-600 text-white hover:bg-blue-500 transition-all ml-2"
              >
                <LogIn className="w-4 h-4" />
                <span>Login</span>
              </Link>
            )}
          </nav>
        </div>
      </header>

      {/* Main Page Content */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {children}
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-800/80 py-4 text-center text-xs text-gray-500 bg-gray-950">
        Smart AI POS & Merchant Engine • Hexagonal Golang Backend + React Next.js Frontend
      </footer>
    </div>
  );
}

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
          <AuthProvider>
            <NavbarContent>{children}</NavbarContent>
          </AuthProvider>
        </QueryClientProvider>
      </body>
    </html>
  );
}
