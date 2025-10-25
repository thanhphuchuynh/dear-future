'use client';

import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react';
import { useRouter, useParams } from 'next/navigation';
import ProtectedRoute from '@/components/ProtectedRoute';
import Navigation from '@/components/Navigation';
import { apiClient } from '@/lib/api-client';
import type {
  Message,
  Attachment,
  RecurrencePattern,
  DeliveryMethod,
} from '@/lib/types';

export default function EditMessagePage() {
  const router = useRouter();
  const params = useParams();
  const messageId = params.id as string;

  const [message, setMessage] = useState<Message | null>(null);
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    delivery_date: '',
    timezone: '',
    delivery_method: 'email' as DeliveryMethod,
    recurrence: 'none' as RecurrencePattern,
  });
  const MAX_ATTACHMENTS = 5;
  const MAX_FILE_SIZE = 10 * 1024 * 1024;
  const recurrenceOptions: { value: RecurrencePattern; label: string }[] = [
    { value: 'none', label: 'One-time delivery' },
    { value: 'daily', label: 'Daily' },
    { value: 'weekly', label: 'Weekly' },
    { value: 'monthly', label: 'Monthly' },
    { value: 'yearly', label: 'Yearly' },
  ];

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [reminderMinutes, setReminderMinutes] = useState('');
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [attachmentError, setAttachmentError] = useState('');
  const [attachmentLoading, setAttachmentLoading] = useState(false);

  useEffect(() => {
    const fetchMessage = async () => {
      try {
        const data = await apiClient.getMessage(messageId);
        setMessage(data);
        setAttachments(data.attachments ?? []);
        setReminderMinutes(
          typeof data.reminder_minutes === 'number'
            ? data.reminder_minutes.toString()
            : ''
        );

        const deliveryDate = new Date(data.delivery_date);
        const formattedDate = deliveryDate.toISOString().slice(0, 16);

        setFormData({
          title: data.title,
          content: data.content,
          delivery_date: formattedDate,
          timezone: data.timezone,
          delivery_method: data.delivery_method,
          recurrence: data.recurrence,
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load message');
      } finally {
        setLoading(false);
      }
    };

    fetchMessage();
  }, [messageId]);

  const handleAttachmentUpload = async (e: ChangeEvent<HTMLInputElement>) => {
    if (!message || !e.target.files) return;

    const files = Array.from(e.target.files);
    const availableSlots = MAX_ATTACHMENTS - attachments.length;
    if (availableSlots <= 0) {
      setAttachmentError(`Maximum of ${MAX_ATTACHMENTS} attachments reached.`);
      e.target.value = '';
      return;
    }

    setAttachmentError('');
    setAttachmentLoading(true);

    try {
      for (const file of files.slice(0, availableSlots)) {
        if (file.size > MAX_FILE_SIZE) {
          setAttachmentError(`"${file.name}" exceeds the 10MB limit.`);
          continue;
        }

        const uploaded = await apiClient.uploadAttachment(message.id, file);
        setAttachments((prev) => [...prev, uploaded]);
      }
    } catch (err) {
      setAttachmentError(
        err instanceof Error ? err.message : 'Failed to upload attachment'
      );
    } finally {
      setAttachmentLoading(false);
      e.target.value = '';
    }
  };

  const handleAttachmentDelete = async (attachmentId: string) => {
    if (!message) return;
    if (!confirm('Remove this attachment?')) return;

    setAttachmentError('');
    setAttachmentLoading(true);
    try {
      await apiClient.deleteAttachment(message.id, attachmentId);
      setAttachments((prev) => prev.filter((att) => att.id !== attachmentId));
    } catch (err) {
      setAttachmentError(
        err instanceof Error ? err.message : 'Failed to delete attachment'
      );
    } finally {
      setAttachmentLoading(false);
    }
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');
    setSaving(true);

    try {
      const payload = {
        ...formData,
        reminder_minutes: reminderMinutes ? Number(reminderMinutes) : undefined,
      };

      await apiClient.updateMessage(messageId, payload);
      router.push('/messages');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update message');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this message?')) {
      return;
    }

    try {
      await apiClient.deleteMessage(messageId);
      router.push('/messages');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete message');
    }
  };

  if (loading) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
          <Navigation />
          <div className="max-w-3xl mx-auto py-6 sm:px-6 lg:px-8">
            <div className="px-4 py-6 sm:px-0">
              <p className="text-gray-500 dark:text-gray-400">Loading...</p>
            </div>
          </div>
        </div>
      </ProtectedRoute>
    );
  }

  if (error && !message) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
          <Navigation />
          <div className="max-w-3xl mx-auto py-6 sm:px-6 lg:px-8">
            <div className="px-4 py-6 sm:px-0">
              <p className="text-red-600 dark:text-red-400">{error}</p>
            </div>
          </div>
        </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <Navigation />
        <div className="max-w-3xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="px-4 py-6 sm:px-0">
            <div className="flex justify-between items-center mb-6">
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                Edit Message
              </h1>
              <button
                onClick={handleDelete}
                className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              >
                Delete Message
              </button>
            </div>

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
                    placeholder="Leave blank for none"
                    value={reminderMinutes}
                    onChange={(e) => setReminderMinutes(e.target.value)}
                  />
                </div>

                {message && (
                  <div className="bg-gray-50 dark:bg-gray-900/50 p-4 rounded-md">
                    <div className="text-xs text-gray-500 dark:text-gray-400 space-y-1">
                      <p>Status: <span className="font-medium">{message.status}</span></p>
                      <p>Created: {new Date(message.created_at).toLocaleString()}</p>
                      <p>Last Updated: {new Date(message.updated_at).toLocaleString()}</p>
                      <p>Recurrence: {message.recurrence === 'none' ? 'One-time' : message.recurrence}</p>
                    </div>
                  </div>
                )}

                {message && (
                  <div>
                    <div className="flex items-center justify-between">
                      <div>
                        <h2 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                          Attachments
                        </h2>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {attachments.length} / {MAX_ATTACHMENTS} files attached
                        </p>
                      </div>
                      <label className="inline-flex items-center px-3 py-2 text-sm font-medium rounded-md bg-blue-50 text-blue-700 dark:bg-blue-900/40 dark:text-blue-200 hover:bg-blue-100 cursor-pointer">
                        <input
                          type="file"
                          className="hidden"
                          multiple
                          disabled={attachments.length >= MAX_ATTACHMENTS}
                          onChange={handleAttachmentUpload}
                        />
                        {attachmentLoading ? 'Uploading...' : 'Add Files'}
                      </label>
                    </div>
                    {attachmentError && (
                      <p className="mt-2 text-sm text-red-600 dark:text-red-400">{attachmentError}</p>
                    )}
                    {attachments.length === 0 ? (
                      <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
                        No attachments yet.
                      </p>
                    ) : (
                      <ul className="mt-3 space-y-2">
                        {attachments.map((attachment) => (
                          <li
                            key={attachment.id}
                            className="flex items-center justify-between rounded-md border border-gray-200 dark:border-gray-700 px-3 py-2 text-sm"
                          >
                            <div>
                              <p className="text-gray-800 dark:text-gray-100">
                                {attachment.file_name}
                              </p>
                              <p className="text-xs text-gray-500 dark:text-gray-400">
                                {(attachment.file_size / (1024 * 1024)).toFixed(2)} MB ·{' '}
                                {new Date(attachment.uploaded_at).toLocaleString()}
                              </p>
                            </div>
                            <div className="flex items-center gap-3">
                              {attachment.download_url && (
                                <a
                                  href={attachment.download_url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="text-blue-600 dark:text-blue-300 hover:underline"
                                >
                                  Download
                                </a>
                              )}
                              <button
                                type="button"
                                onClick={() => handleAttachmentDelete(attachment.id)}
                                className="text-red-500 hover:text-red-600"
                              >
                                Remove
                              </button>
                            </div>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                )}

                <div className="flex gap-4 pt-4">
                  <button
                    type="submit"
                    disabled={saving}
                    className="flex-1 inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {saving ? 'Saving...' : 'Save Changes'}
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
