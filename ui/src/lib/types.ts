// API Types matching backend responses

export interface User {
  id: string;
  email: string;
  name?: string;
  timezone?: string;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name?: string;
  timezone?: string;
}

export type DeliveryMethod = 'email' | 'push';
export type RecurrencePattern = 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';

export interface Attachment {
  id: string;
  message_id: string;
  file_name: string;
  file_type: string;
  file_size: number;
  download_url?: string;
  uploaded_at: string;
}

export interface Message {
  id: string;
  user_id: string;
  title: string;
  content: string;
  delivery_date: string;
  timezone: string;
  status: 'scheduled' | 'sent' | 'failed' | 'cancelled';
  delivery_method: DeliveryMethod;
  recurrence: RecurrencePattern;
  reminder_minutes?: number;
  attachment_count: number;
  attachments?: Attachment[];
  created_at: string;
  updated_at: string;
}

export interface CreateMessageRequest {
  title: string;
  content: string;
  delivery_date: string;
  timezone: string;
  delivery_method: DeliveryMethod;
  recurrence: RecurrencePattern;
  reminder_minutes?: number;
}

export interface UpdateMessageRequest {
  title?: string;
  content?: string;
  delivery_date?: string;
  timezone?: string;
  delivery_method?: DeliveryMethod;
  recurrence?: RecurrencePattern;
  reminder_minutes?: number;
}

export interface ApiError {
  error: string;
}

export interface HealthStatus {
  status: string;
  services?: {
    database?: { status: string };
    email?: { status: string };
  };
}

export interface AnalyticsSummary {
  totals: {
    total: number;
    scheduled: number;
    delivered: number;
    failed: number;
    cancelled: number;
  };
  delivery_rate: number;
  upcoming?: UpcomingSummary;
  recent_messages: MessageOverview[];
  attachment_count: number;
}

export interface UpcomingSummary {
  id: string;
  title: string;
  delivery_date: string;
  delivery_method: DeliveryMethod;
}

export interface MessageOverview {
  id: string;
  title: string;
  status: string;
  delivery_date: string;
  delivery_method: DeliveryMethod;
}
