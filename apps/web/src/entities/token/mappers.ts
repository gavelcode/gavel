import type { components } from "@/shared/api/v1.gen";
import type { Token, CreateTokenResult } from "./model";

export type TokenDTO = components["schemas"]["TokenSummary"];
export type CreateTokenDTO = components["schemas"]["CreatedToken"];

export function toToken(dto: TokenDTO): Token {
  return {
    id: dto.id,
    name: dto.name,
    prefix: dto.prefix,
    scopes: dto.scopes,
    createdAt: dto.created_at,
    lastUsedAt: dto.last_used_at ?? undefined,
    expiresAt: dto.expires_at ?? undefined,
  };
}

export function toCreateTokenResult(dto: CreateTokenDTO): CreateTokenResult {
  return {
    id: dto.id,
    name: dto.name,
    scopes: dto.scopes,
    token: dto.token,
    prefix: dto.prefix,
  };
}
