'use client';

import { useState } from 'react';
import { useAuth } from '@/lib/auth';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const { login, user } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  // Redirect if already logged in
  if (user) {
    router.push('/');
    return null;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      router.push('/');
    } catch (err: any) {
      setError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-950">
      <div className="w-full max-w-sm rounded-xl border border-slate-800 bg-slate-900 p-8 shadow-2xl">
        <div className="mb-6 text-center">
          <h1 className="text-2xl font-bold text-white">NeXus PMS</h1>
          <p className="mt-1 text-sm text-slate-400">Sign in to your account</p>
        </div>

        {error && (
          <div className="mb-4 rounded-lg border border-red-800 bg-red-950/50 px-4 py-3 text-sm text-red-400">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-400">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="owner@villa.test"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white placeholder-slate-500 focus:border-nexus-500 focus:outline-none"
              required
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-400">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••"
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white placeholder-slate-500 focus:border-nexus-500 focus:outline-none"
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-lg bg-nexus-600 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-nexus-500 disabled:opacity-50"
          >
            {loading ? 'Signing in…' : 'Sign In'}
          </button>
        </form>

        <div className="mt-6 border-t border-slate-800 pt-4">
          <p className="mb-2 text-xs font-medium text-slate-500">Demo Villa Owners</p>
          <div className="space-y-1 text-xs text-slate-400">
            <button
              onClick={() => { setEmail('ramiz@villa.test'); setPassword('villa123'); }}
              className="block w-full rounded bg-slate-800/50 px-2 py-1 text-left hover:bg-slate-800"
            >
              Ramiz Haddad — ramiz@villa.test / villa123
            </button>
            <button
              onClick={() => { setEmail('owner1@villa.test'); setPassword('villa123'); }}
              className="block w-full rounded bg-slate-800/50 px-2 py-1 text-left hover:bg-slate-800"
            >
              Ahmad Khalil — owner1@villa.test / villa123
            </button>
            <button
              onClick={() => { setEmail('owner2@villa.test'); setPassword('villa123'); }}
              className="block w-full rounded bg-slate-800/50 px-2 py-1 text-left hover:bg-slate-800"
            >
              Sarah Nassar — owner2@villa.test / villa123
            </button>
            <button
              onClick={() => { setEmail('owner4@villa.test'); setPassword('villa123'); }}
              className="block w-full rounded bg-slate-800/50 px-2 py-1 text-left hover:bg-slate-800"
            >
              Layla Farhat — owner4@villa.test / villa123
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
