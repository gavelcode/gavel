import { request } from "@/shared/api/client";
import { v1Path } from "@/shared/api/types";
import type { Token, CreateTokenResult } from "./model";
import { type TokenDTO, type CreateTokenDTO, toToken, toCreateTokenResult } from "./mappers";

interface TokenListWire {
  items: TokenDTO[];
  next_cursor: string | null;
}

export async function listTokens(): Promise<Token[]> {
  const wire = await request<TokenListWire>(v1Path("/me/tokens"));
  return wire.items.map(toToken);
}

export async function createToken(
  name: string,
  scopes: string[],
  expiresInDays?: number,
): Promise<CreateTokenResult> {
  const dto = await request<CreateTokenDTO>(v1Path("/me/tokens"), {
    method: "POST",
    body: JSON.stringify({
      name,
      scopes,
      expires_in_days: expiresInDays,
    }),
  });
  return toCreateTokenResult(dto);
}

export async function deleteToken(id: string): Promise<void> {
  await request<void>(v1Path(`/me/tokens/${id}`), { method: "DELETE" });
}
