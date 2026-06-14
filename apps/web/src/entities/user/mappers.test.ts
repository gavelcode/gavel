import { toUser, type UserDTO } from "./mappers";

describe("toUser", () => {
  it("maps DTO snake_case to model camelCase", () => {
    const dto: UserDTO = {
      id: "42",
      email: "admin@local",
      display_name: "Admin User",
      role: "admin",
      must_change_password: true,
    };

    const user = toUser(dto);

    expect(user).toEqual({
      id: "42",
      email: "admin@local",
      displayName: "Admin User",
      role: "admin",
      mustChangePassword: true,
    });
  });
});
