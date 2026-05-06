export type NotificationStatus =
  | "queued"
  | "processing"
  | "sent"
  | "delivered"
  | "failed_temporary"
  | "failed_permanent"
  | "bounced"
  | "complained"
  | "suppressed"
  | "cancelled";

export interface Notification {
  id: string;
  channel: string;
  template_key?: string;
  recipient_ref?: string;
  status: NotificationStatus;
  metadata?: Record<string, string>;
  created_at: string;
  updated_at?: string;
}

export interface Delivery {
  id: string;
  notification_id: string;
  channel: string;
  provider_type?: string;
  provider_message_id?: string;
  status: NotificationStatus;
  attempt_count: number;
  last_error_code?: string;
  last_error_message?: string;
  sent_at?: string;
  delivered_at?: string;
  failed_at?: string;
  created_at: string;
  updated_at?: string;
}

export interface DeliveryEvent {
  id: string;
  delivery_id: string;
  event_type: string;
  provider_type?: string;
  provider_event_id?: string;
  occurred_at: string;
  created_at: string;
}

export interface ProviderConnection {
  id: string;
  provider_type: string;
  channel: string;
  display_name: string;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
}

export interface Template {
  id: string;
  key: string;
  name: string;
  channel: string;
  active_version?: number;
  created_at: string;
  updated_at?: string;
}

export interface TemplateVersion {
  id: string;
  template_id: string;
  version_number: number;
  subject?: string;
  html_body?: string;
  text_body?: string;
  status: string;
  created_at: string;
}
