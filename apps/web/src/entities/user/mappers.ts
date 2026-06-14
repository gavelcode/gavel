import type { components } from "@/shared/api/v1.gen";
import type { User } from "./model";

export type UserDTO = components["schemas"]["Me"];
export type CreateUserDTO = components["schemas"]["CreatedUser"];

export function toUser(dto: UserDTO): User {
  return {
    id: dto.id,
    email: dto.email,
    displayName: dto.display_name,
    role: dto.role,
    mustChangePassword: dto.must_change_password,
  };
}
