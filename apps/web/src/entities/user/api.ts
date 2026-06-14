import { request } from "@/shared/api/client";
import { v1Path } from "@/shared/api/types";
import type { User } from "./model";
import { type UserDTO, type CreateUserDTO, toUser } from "./mappers";

export async function login(email: string, password: string): Promise<User> {
  const dto = await request<UserDTO>(v1Path("/sessions"), {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  return toUser(dto);
}

export async function logout(): Promise<void> {
  await request<void>(v1Path("/sessions/current"), { method: "DELETE" });
}

export async function me(): Promise<User> {
  const dto = await request<UserDTO>(v1Path("/me"));
  return toUser(dto);
}

export async function changePassword(
  currentPassword: string,
  newPassword: string,
): Promise<void> {
  await request<void>(v1Path("/me/password"), {
    method: "POST",
    body: JSON.stringify({
      current_password: currentPassword,
      new_password: newPassword,
    }),
  });
}

export async function createUser(
  email: string,
  displayName: string,
  password: string,
  role: string,
): Promise<CreateUserDTO> {
  return request<CreateUserDTO>(v1Path("/admin/users"), {
    method: "POST",
    body: JSON.stringify({
      email,
      display_name: displayName,
      password,
      role,
    }),
  });
}
