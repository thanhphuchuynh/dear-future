// API Client for Dear Future backend

import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  Message,
  CreateMessageRequest,
  UpdateMessageRequest,
  User,
  HealthStatus,
  ApiError,
  Attachment,
  AnalyticsSummary,
} from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api/v1';

class ApiClient {
  private accessToken: string | null = null;

  constructor() {
    // Load token from localStorage on client side
    if (typeof window !== 'undefined') {
      this.accessToken = localStorage.getItem('access_token');
    }
  }

  setAccessToken(token: string | null) {
    this.accessToken = token;
    if (typeof window !== 'undefined') {
      if (token) {
        localStorage.setItem('access_token', token);
      } else {
        localStorage.removeItem('access_token');
      }
    }
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const headers = new Headers(options.headers);

    const isFormData = typeof FormData !== 'undefined' && options.body instanceof FormData;
    if (!isFormData && !headers.has('Content-Type')) {
      headers.set('Content-Type', 'application/json');
    }

    if (this.accessToken && !endpoint.includes('/auth/')) {
      headers.set('Authorization', `Bearer ${this.accessToken}`);
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: 'An unknown error occurred',
      }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Health Check
  async health(): Promise<HealthStatus> {
    return this.request('/health');
  }

  // Auth Endpoints
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setAccessToken(response.access_token);
    return response;
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setAccessToken(response.access_token);
    return response;
  }

  logout() {
    this.setAccessToken(null);
    if (typeof window !== 'undefined') {
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
    }
  }

  // User Endpoints
  async getProfile(): Promise<User> {
    return this.request<User>('/user/profile');
  }

  async updateProfile(data: Partial<User>): Promise<User> {
    return this.request<User>('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // Message Endpoints
  async getMessages(): Promise<Message[]> {
    return this.request<Message[]>('/messages');
  }

  async getMessage(id: string): Promise<Message> {
    return this.request<Message>(`/messages?id=${encodeURIComponent(id)}`);
  }

  async createMessage(data: CreateMessageRequest): Promise<Message> {
    data.delivery_date = new Date(data.delivery_date).toISOString();

    return this.request<Message>('/messages', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateMessage(id: string, data: UpdateMessageRequest): Promise<Message> {
    const payload: UpdateMessageRequest = { ...data };
    if (payload.delivery_date) {
      payload.delivery_date = new Date(payload.delivery_date).toISOString();
    }

    return this.request<Message>(`/messages?id=${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
  }

  async deleteMessage(id: string): Promise<void> {
    await this.request<void>(`/messages?id=${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  }

  async uploadAttachment(messageId: string, file: File): Promise<Attachment> {
    const formData = new FormData();
    formData.append('file', file);

    return this.request<Attachment>(
      `/messages/attachments?message_id=${encodeURIComponent(messageId)}`,
      {
        method: 'POST',
        body: formData,
      }
    );
  }

  async listAttachments(messageId: string): Promise<Attachment[]> {
    return this.request<Attachment[]>(
      `/messages/attachments?message_id=${encodeURIComponent(messageId)}`
    );
  }

  async deleteAttachment(messageId: string, attachmentId: string): Promise<void> {
    await this.request<void>(
      `/messages/attachments?message_id=${encodeURIComponent(
        messageId
      )}&attachment_id=${encodeURIComponent(attachmentId)}`,
      {
        method: 'DELETE',
      }
    );
  }

  async getAnalyticsSummary(): Promise<AnalyticsSummary> {
    return this.request<AnalyticsSummary>('/analytics/summary');
  }
}

// Export singleton instance
export const apiClient = new ApiClient();
