'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import ProtectedRoute from '@/components/ProtectedRoute';
import Navigation from '@/components/Navigation';
import { apiClient } from '@/lib/api-client';
import type { Message, AnalyticsSummary } from '@/lib/types';

export default function DashboardPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [analytics, setAnalytics] = useState<AnalyticsSummary | null>(null);
  const [analyticsError, setAnalyticsError] = useState('');
  const [analyticsLoading, setAnalyticsLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    setAnalyticsLoading(true);
    const [messagesResult, analyticsResult] = await Promise.allSettled([
      apiClient.getMessages(),
      apiClient.getAnalyticsSummary(),
    ]);

    if (messagesResult.status === 'fulfilled') {
      setMessages(messagesResult.value);
      setError('');
    } else {
      setError(
        messagesResult.reason instanceof Error
          ? messagesResult.reason.message
          : 'Failed to load messages'
      );
    }

    if (analyticsResult.status === 'fulfilled') {
      setAnalytics(analyticsResult.value);
      setAnalyticsError('');
    } else {
      setAnalyticsError(
        analyticsResult.reason instanceof Error
          ? analyticsResult.reason.message
          : 'Analytics unavailable'
      );
    }

    setLoading(false);
    setAnalyticsLoading(false);
  };

  const totals = analytics?.totals ?? {
    total: messages.length,
    scheduled: messages.filter((m) => m.status === 'scheduled').length,
    delivered: messages.filter((m) => m.status === 'delivered').length,
    failed: messages.filter((m) => m.status === 'failed').length,
    cancelled: messages.filter((m) => m.status === 'cancelled').length,
  };
  const deliveryRate = analytics?.delivery_rate ?? 0;
  const upcoming = analytics?.upcoming ?? null;
  const recentMessages = analytics?.recent_messages ?? messages.slice(0, 5).map((m) => ({
    id: m.id,
    title: m.title,
    status: m.status,
    delivery_date: m.delivery_date,
    delivery_method: m.delivery_method,
  }));

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <Navigation />
        <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="px-4 py-6 sm:px-0">
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">
              Dashboard
            </h1>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
              {[
                {
                  label: 'Total Messages',
                  value: totals.total,
                  icon: (
                    <svg className="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                      />
                    </svg>
                  ),
                },
                {
                  label: 'Scheduled',
                  value: totals.scheduled,
                  icon: (
                    <svg className="h-6 w-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  ),
                },
                {
                  label: 'Delivered',
                  value: `${totals.delivered} (${deliveryRate.toFixed(1)}%)`,
                  icon: (
                    <svg className="h-6 w-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  ),
                },
                {
                  label: 'Attachments',
                  value:
                    analytics && !analyticsLoading
                      ? analytics.attachment_count
                      : analyticsLoading
                      ? '…'
                      : '—',
                  icon: (
                    <svg className="h-6 w-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.415-6.414a4 4 0 00-5.657-5.657L6.343 9.172"
                      />
                    </svg>
                  ),
                },
              ].map((card) => (
                <div key={card.label} className="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                  <div className="p-5">
                    <div className="flex items-center">
                      <div className="flex-shrink-0">{card.icon}</div>
                      <div className="ml-5 w-0 flex-1">
                        <dl>
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                            {card.label}
                          </dt>
                          <dd className="text-lg font-semibold text-gray-900 dark:text-white">
                            {card.value}
                          </dd>
                        </dl>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2 mb-8">
              <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Quick Actions
                </h2>
                <div className="flex gap-4 flex-wrap">
                  <Link
                    href="/messages/new"
                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    <svg
                      className="h-5 w-5 mr-2"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    New Message
                  </Link>
                  <Link
                    href="/messages"
                    className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    View All Messages
                  </Link>
                </div>
              </div>

              <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Upcoming Delivery
                </h2>
                {analyticsLoading ? (
                  <p className="text-gray-500 dark:text-gray-400">Loading analytics...</p>
                ) : analyticsError ? (
                  <p className="text-red-600 dark:text-red-400 text-sm">{analyticsError}</p>
                ) : upcoming ? (
                  <div className="space-y-1 text-sm text-gray-700 dark:text-gray-200">
                    <p className="text-base font-medium text-gray-900 dark:text-white">
                      {upcoming.title}
                    </p>
                    <p>{new Date(upcoming.delivery_date).toLocaleString()}</p>
                    <p>Via {upcoming.delivery_method}</p>
                  </div>
                ) : (
                  <p className="text-gray-500 dark:text-gray-400 text-sm">
                    No upcoming deliveries scheduled.
                  </p>
                )}
              </div>
            </div>

            {/* Recent Messages */}
            <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Recent Messages
                </h2>
                {loading && recentMessages.length === 0 ? (
                  <p className="text-gray-500 dark:text-gray-400">Loading...</p>
                ) : error ? (
                  <p className="text-red-600 dark:text-red-400">{error}</p>
                ) : recentMessages.length === 0 ? (
                  <div className="text-center py-8">
                    <svg
                      className="mx-auto h-12 w-12 text-gray-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
                      />
                    </svg>
                    <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
                      No messages
                    </h3>
                    <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                      Get started by creating a new message to your future self.
                    </p>
                    <div className="mt-6">
                      <Link
                        href="/messages/new"
                        className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                      >
                        Create your first message
                      </Link>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    {recentMessages.map((message) => (
                      <div
                        key={message.id}
                        className="border-l-4 border-blue-500 pl-4 py-2"
                      >
                        <div className="flex items-center justify-between">
                          <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                            {message.title}
                          </h3>
                          <span
                            className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                              message.status === 'scheduled'
                                ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                                : message.status === 'delivered'
                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
                            }`}
                          >
                            {message.status}
                          </span>
                        </div>
                        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                          Deliver on: {new Date(message.delivery_date).toLocaleDateString()}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          Via {message.delivery_method}
                        </p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
