export type UserRole = "viewer" | "maintainer" | "admin";

export interface User {
  id: string;
  email: string;
  displayName: string;
  role: UserRole;
  mustChangePassword: boolean;
}
