import { useState, type FormEvent } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/entities/user/use-auth";
import * as userApi from "@/entities/user/api";
import { ApiErrorResponse } from "@/shared/api/client";
import { TopBar } from "@/shared/ui/top-bar";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/shared/ui/card";

const ROLES = ["viewer", "maintainer", "admin"] as const;

export function AdminUsersPage() {
  const { user } = useAuth();
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<string>("viewer");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [submitting, setSubmitting] = useState(false);

  if (user?.role !== "admin") {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    setSubmitting(true);
    try {
      const created = await userApi.createUser(email, displayName, password, role);
      setSuccess(`User ${created.email} created with role ${created.role}`);
      setEmail("");
      setDisplayName("");
      setPassword("");
      setRole("viewer");
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
    <div className="flex flex-col">
      <TopBar crumbs={["Users"]} />
      <div className="flex flex-col gap-d-gap overflow-auto p-d-page">

      <Card elevated>
        <CardHeader>
          <CardTitle>Create user</CardTitle>
          <CardDescription>
            New users must change their password on first login.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="max-w-sm space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
            {success && (
              <div className="rounded-md bg-success/10 px-3 py-2 text-sm text-success">
                {success}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="user-email">Email</Label>
              <Input
                id="user-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-name">Display name</Label>
              <Input
                id="user-name"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-password">Password</Label>
              <Input
                id="user-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="user-role">Role</Label>
              <select
                id="user-role"
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-raised focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                {ROLES.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </div>
            <Button type="submit" loading={submitting}>
              Create user
            </Button>
          </form>
        </CardContent>
      </Card>
      </div>
    </div>
  );
}
