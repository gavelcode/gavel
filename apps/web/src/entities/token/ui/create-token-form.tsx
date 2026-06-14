import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import type { CreateTokenResult } from "../model";
import * as tokenApi from "../api";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";

const SCOPES = ["ingest", "read", "admin"] as const;

interface CreateTokenFormProps {
  onCreated: (result: CreateTokenResult) => void;
  onCancel: () => void;
}

export function CreateTokenForm({ onCreated, onCancel }: CreateTokenFormProps) {
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState<Set<string>>(new Set());
  const [expiresInDays, setExpiresInDays] = useState("");
  const [error, setError] = useState("");

  const mutation = useMutation({
    mutationFn: () =>
      tokenApi.createToken(
        name,
        Array.from(scopes),
        expiresInDays ? parseInt(expiresInDays, 10) : undefined,
      ),
    onSuccess: onCreated,
    onError: (err: Error) => setError(err.message),
  });

  const toggleScope = (scope: string) => {
    setScopes((prev) => {
      const next = new Set(prev);
      if (next.has(scope)) next.delete(scope);
      else next.add(scope);
      return next;
    });
  };

  return (
    <Card elevated>
      <CardHeader>
        <CardTitle>Create API token</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setError("");
            if (scopes.size === 0) {
              setError("Select at least one scope");
              return;
            }
            mutation.mutate();
          }}
          className="max-w-sm space-y-4"
        >
          {error && (
            <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="token-name">Name</Label>
            <Input
              id="token-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. CI pipeline"
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Scopes</Label>
            <div className="flex gap-3">
              {SCOPES.map((scope) => (
                <label key={scope} className="flex items-center gap-1.5 text-sm">
                  <input
                    type="checkbox"
                    checked={scopes.has(scope)}
                    onChange={() => toggleScope(scope)}
                    className="rounded"
                  />
                  {scope}
                </label>
              ))}
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="token-expires">Expires in (days, optional)</Label>
            <Input
              id="token-expires"
              type="number"
              min="1"
              value={expiresInDays}
              onChange={(e) => setExpiresInDays(e.target.value)}
              placeholder="Never"
            />
          </div>
          <div className="flex gap-2">
            <Button type="submit" loading={mutation.isPending}>
              Create
            </Button>
            <Button type="button" variant="ghost" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
