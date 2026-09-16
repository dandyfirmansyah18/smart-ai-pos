'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth, UserRole } from '@/context/auth-context';
import { ShieldAlert, Loader2 } from 'lucide-react';

interface ProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles?: UserRole[];
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, allowedRoles }) => {
  const { user, token, isLoading, hasRole } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && (!token || !user)) {
      router.push('/login');
    }
  }, [isLoading, token, user, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-900 text-white">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
        <span className="ml-3 text-lg font-medium">Loading session...</span>
      </div>
    );
  }

  if (!token || !user) {
    return null;
  }

  if (allowedRoles && allowedRoles.length > 0 && !hasRole(allowedRoles)) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-gray-900 text-white p-6">
        <ShieldAlert className="h-16 w-16 text-red-500 mb-4" />
        <h1 className="text-2xl font-bold mb-2">Access Denied</h1>
        <p className="text-gray-400 mb-6 text-center max-w-md">
          Your role ({user.role}) does not have permission to access this internal management portal.
        </p>
        <button
          onClick={() => router.push('/dashboard')}
          className="rounded-lg bg-blue-600 px-6 py-2 font-semibold text-white hover:bg-blue-700 transition-colors"
        >
          Return to Dashboard
        </button>
      </div>
    );
  }

  return <>{children}</>;
};

export default ProtectedRoute;
