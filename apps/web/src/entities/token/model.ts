export interface Token {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  createdAt: string;
  lastUsedAt?: string;
  expiresAt?: string;
}

export interface CreateTokenResult {
  id: string;
  name: string;
  scopes: string[];
  token: string;
  prefix: string;
}
