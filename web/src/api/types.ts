// API Types matching the backend models

export interface Account {
  id: number;
  name: string;
  server: string;
  port: number;
  username: string;
  password?: string;
  tls: boolean;
  created_at: string;
  updated_at: string;
}

export interface AccountCreate {
  name: string;
  server: string;
  port: number;
  username: string;
  password: string;
  tls: boolean;
}

export interface Rule {
  id: number;
  account_id: number;
  name: string;
  pattern: string;
  pattern_type: 'sender' | 'subject' | 'from_domain';
  move_to_folder: string;
  enabled: boolean;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface RuleCreate {
  name: string;
  pattern: string;
  pattern_type: 'sender' | 'subject' | 'from_domain';
  move_to_folder: string;
  enabled: boolean;
  priority: number;
}

export interface Message {
  uid: number;
  seq_num: number;
  from: string;
  to: string;
  subject: string;
  date: string;
  flags: string[];
  matched_rule?: Rule;
}

export interface Folder {
  name: string;
  delimiter: string;
  attributes: string[];
}

export interface ConnectionStatus {
  success: boolean;
  message: string;
  folders?: Folder[];
  total_emails?: number;
}

export interface PreviewResult {
  total_messages: number;
  matched_messages: number;
  messages: Message[];
  rule_matches: Record<number, number>;
}

export interface WSMessage {
  type: 'progress' | 'result' | 'error' | 'pong';
  payload?: unknown;
  error?: string;
}

export interface PreviewProgress {
  stage: 'connecting' | 'connected' | 'selecting' | 'fetching' | 'processing';
  current: number;
  total: number;
  message: string;
  message_data?: Message;
}
