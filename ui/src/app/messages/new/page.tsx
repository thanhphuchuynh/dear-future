'use client';

import { useState, type ChangeEvent, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import ProtectedRoute from '@/components/ProtectedRoute';
import Navigation from '@/components/Navigation';
import { apiClient } from '@/lib/api-client';
import type { DeliveryMethod, RecurrencePattern } from '@/lib/types';

export default function NewMessagePage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    delivery_date: '',
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
    delivery_method: 'email' as DeliveryMethod,
    recurrence: 'none' as RecurrencePattern,
  });
  const [reminderMinutes, setReminderMinutes] = useState('');
  const [attachments, setAttachments] = useState<File[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');

  const MAX_ATTACHMENTS = 5;
  const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
  const recurrenceOptions: { value: RecurrencePattern; label: string }[] = [
    { value: 'none', label: 'One-time delivery' },
    { value: 'daily', label: 'Daily' },
    { value: 'weekly', label: 'Weekly' },
    { value: 'monthly', label: 'Monthly' },
    { value: 'yearly', label: 'Yearly' },
  ];

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;

    const selected = Array.from(e.target.files);
    const availableSlots = MAX_ATTACHMENTS - attachments.length;
    if (availableSlots <= 0) {
      setError(`You can attach up to ${MAX_ATTACHMENTS} files.`);
      return;
    }

    const accepted: File[] = [];
    for (const file of selected.slice(0, availableSlots)) {
      if (file.size > MAX_FILE_SIZE) {
        setError(`"${file.name}" exceeds the 10MB size limit.`);
        continue;
      }
      accepted.push(file);
    }

    if (accepted.length) {
      setAttachments((prev) => [...prev, ...accepted]);
      setError('');
    }

    e.target.value = '';
  };

  const removeAttachment = (index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const payload = {
        ...formData,
        reminder_minutes: reminderMinutes ? Number(reminderMinutes) : undefined,
      };

      const message = await apiClient.createMessage(payload);

      if (attachments.length) {
        setUploading(true);
        try {
          for (const file of attachments) {
            await apiClient.uploadAttachment(message.id, file);
          }
        } catch (uploadErr) {
          console.error(uploadErr);
          alert(
            'Message created, but at least one attachment failed to upload. You can manage attachments from the message details page.'
          );
          router.push(`/messages/${message.id}`);
          return;
        } finally {
          setUploading(false);
        }
      }

      router.push('/messages');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create message');
    } finally {
      setLoading(false);
    }
  };

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <Navigation />
        <div className="max-w-3xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="px-4 py-6 sm:px-0">
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-6">
              Create New Message
            </h1>

            <form
              onSubmit={handleSubmit}
              className="bg-white dark:bg-gray-800 shadow rounded-lg p-6"
            >
              {error && (
                <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4 mb-6">
                  <p className="text-sm text-red-800 dark:text-red-200">{error}</p>
                </div>
              )}

              <div className="space-y-6">
                <div>
                  <label
                    htmlFor="title"
                    className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    Title *
                  </label>
                  <input
                    type="text"
                    id="title"
                    required
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="Give your message a title"
                    value={formData.title}
                    onChange={(e) =>
                      setFormData({ ...formData, title: e.target.value })
                    }
                  />
                </div>

                <div>
                  <label
                    htmlFor="content"
                    className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    Message Content *
                  </label>
                  <textarea
                    id="content"
                    rows={8}
                    required
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="Write a message to your future self..."
                    value={formData.content}
                    onChange={(e) =>
                      setFormData({ ...formData, content: e.target.value })
                    }
                  />
                </div>

                <div>
                  <label
                    htmlFor="delivery_date"
                    className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    Delivery Date & Time *
                  </label>
                  <input
                    type="datetime-local"
                    id="delivery_date"
                    required
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    value={formData.delivery_date}
                    onChange={(e) =>
                      setFormData({ ...formData, delivery_date: e.target.value })
                    }
                  />
                  <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    Choose when you want to receive this message
                  </p>
                </div>

                <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
                  <div>
                    <label
                      htmlFor="delivery_method"
                      className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      Delivery Method *
                    </label>
                    <select
                      id="delivery_method"
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      value={formData.delivery_method}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          delivery_method: e.target.value as DeliveryMethod,
                        })
                      }
                    >
                      <option value="email">Email</option>
                      <option value="push">Push Notification</option>
                    </select>
                  </div>

                  <div>
                    <label
                      htmlFor="recurrence"
                      className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      Recurrence
                    </label>
                    <select
                      id="recurrence"
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      value={formData.recurrence}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          recurrence: e.target.value as RecurrencePattern,
                        })
                      }
                    >
                      {recurrenceOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                <div>
                  <label
                    htmlFor="reminder_minutes"
                    className="block text-sm font-medium text-gray-700 dark:text-gray-300"
                  >
                    Reminder (minutes before delivery)
                  </label>
                  <input
                    type="number"
                    min={0}
                    id="reminder_minutes"
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    placeholder="e.g., 60 for 1 hour"
                    value={reminderMinutes}
                    onChange={(e) => setReminderMinutes(e.target.value)}
                  />
                  <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    Leave blank for no reminder.
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                    Attach Files
                  </label>
                  <input
                    type="file"
                    multiple
                    onChange={handleFileChange}
                    className="mt-1 block w-full text-sm text-gray-900 dark:text-gray-100 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-600 hover:file:bg-blue-100 dark:file:bg-blue-900/40 dark:file:text-blue-200"
                  />
                  <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    Up to {MAX_ATTACHMENTS} files, 10MB each.
                  </p>
                  {attachments.length > 0 && (
                    <ul className="mt-3 space-y-2">
                      {attachments.map((file, index) => (
                        <li
                          key={`${file.name}-${index}`}
                          className="flex items-center justify-between rounded-md border border-gray-200 dark:border-gray-700 px-3 py-2 text-sm"
                        >
                          <span className="text-gray-700 dark:text-gray-200">
                            {file.name} ({(file.size / (1024 * 1024)).toFixed(2)} MB)
                          </span>
                          <button
                            type="button"
                            onClick={() => removeAttachment(index)}
                            className="text-red-500 hover:text-red-600"
                          >
                            Remove
                          </button>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>

                <div className="flex gap-4 pt-4">
                  <button
                    type="submit"
                    disabled={loading || uploading}
                    className="flex-1 inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {loading || uploading ? 'Saving...' : 'Create Message'}
                  </button>
                  <button
                    type="button"
                    onClick={() => router.back()}
                    className="px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
