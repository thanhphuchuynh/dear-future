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

export interface Message {
  id: string;
  user_id: string;
  title: string;
  content: string;
  delivery_date: string;
  timezone: string;
  status: 'scheduled' | 'sent' | 'failed' | 'cancelled';
  delivery_method: 'email' | 'sms' | 'push';
  created_at: string;
  updated_at: string;
}

export interface CreateMessageRequest {
  title: string;
  content: string;
  delivery_date: string;
  timezone: string;
  delivery_method: 'email' | 'sms' | 'push';
}

export interface UpdateMessageRequest {
  title?: string;
  content?: string;
  delivery_date?: string;
  timezone?: string;
  delivery_method?: 'email' | 'sms' | 'push';
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
