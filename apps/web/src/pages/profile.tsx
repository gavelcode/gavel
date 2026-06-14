import { useState, type FormEvent } from "react";
import { useAuth } from "@/entities/user/use-auth";
import * as userApi from "@/entities/user/api";
import { ApiErrorResponse } from "@/shared/api/client";
import { TopBar } from "@/shared/ui/top-bar";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";
import { useDensity, type Density } from "@/shared/lib/use-density";

const DENSITY_OPTIONS: { value: Density; label: string; description: string }[] = [
  { value: "comfortable", label: "Comfortable", description: "More breathing room" },
  { value: "compact", label: "Compact", description: "Default balance" },
  { value: "dense", label: "Dense", description: "Maximum information" },
];

export function ProfilePage() {
  const { user } = useAuth();
  const { density, setDensity } = useDensity();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleChangePassword = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (newPassword !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }
    if (newPassword.length < 8) {
      setError("New password must be at least 8 characters");
      return;
    }

    setSubmitting(true);
    try {
      await userApi.changePassword(currentPassword, newPassword);
      setSuccess("Password changed successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
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
      <TopBar crumbs={["Profile"]} />
      <div className="flex flex-col gap-d-gap overflow-auto p-d-page">

      <Card elevated>
        <CardHeader>
          <CardTitle>Account information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-2 text-sm">
            <span className="text-muted-foreground">Email</span>
            <span>{user?.email}</span>
            <span className="text-muted-foreground">Display name</span>
            <span>{user?.displayName}</span>
            <span className="text-muted-foreground">Role</span>
            <span className="capitalize">{user?.role}</span>
          </div>
        </CardContent>
      </Card>

      <Card elevated>
        <CardHeader>
          <CardTitle>Display density</CardTitle>
        </CardHeader>
        <CardContent>
          <fieldset className="flex gap-3">
            {DENSITY_OPTIONS.map((opt) => (
              <label
                key={opt.value}
                className={`flex flex-1 cursor-pointer flex-col items-center gap-1 rounded-lg border p-3 text-center transition-colors duration-fast ${
                  density === opt.value
                    ? "border-primary bg-primary/5"
                    : "border-border hover:bg-muted/50"
                }`}
              >
                <input
                  type="radio"
                  name="density"
                  value={opt.value}
                  checked={density === opt.value}
                  onChange={() => setDensity(opt.value)}
                  className="sr-only"
                  aria-label={opt.label}
                />
                <span className="text-sm font-medium">{opt.label}</span>
                <span className="text-2xs text-muted-foreground">{opt.description}</span>
              </label>
            ))}
          </fieldset>
        </CardContent>
      </Card>

      <Card elevated>
        <CardHeader>
          <CardTitle>Change password</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleChangePassword} className="max-w-sm space-y-4">
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
              <Label htmlFor="current">Current password</Label>
              <Input
                id="current"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new">New password</Label>
              <Input
                id="new"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm">Confirm new password</Label>
              <Input
                id="confirm"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" loading={submitting}>
              Change password
            </Button>
          </form>
        </CardContent>
      </Card>
      </div>
    </div>
  );
}
