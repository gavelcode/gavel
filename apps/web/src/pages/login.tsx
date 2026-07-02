import { useState, type FormEvent } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { useAuth } from "@/entities/user/use-auth";
import { ApiErrorResponse } from "@/shared/api/client";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/shared/ui/card";

export function LoginPage() {
  const { user, loading, login } = useAuth();
  const navigate = useNavigate();
  const isDev = (import.meta as unknown as { env?: { DEV?: boolean } }).env?.DEV ?? false;
  const [email, setEmail] = useState(isDev ? "admin@local" : "");
  const [password, setPassword] = useState(isDev ? "admin123!" : "");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  if (loading) return null;
  if (user) return <Navigate to="/" replace />;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await login(email, password);
      void navigate("/");
    } catch (err) {
      if (err instanceof ApiErrorResponse) {
        setError(err.message);
      } else {
        setError("Connection error");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card elevated className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Gavel</CardTitle>
          <CardDescription>Sign in to continue</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="w-full" loading={submitting}>
              Sign in
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
