import { http, HttpResponse } from "msw";
import { server } from "@/test/msw-server";
import { fetchSource, fetchSourceWithContext, SourceNotFoundError } from "./api";

describe("fetchSource", () => {
  it("returns source text on success", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        new HttpResponse("package main\n", { status: 200 }),
      ),
    );
    const result = await fetchSource("payment", "abc1234", "main.go");
    expect(result).toBe("package main\n");
  });

  it("throws SourceNotFoundError on 404", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        new HttpResponse(null, { status: 404 }),
      ),
    );
    await expect(fetchSource("payment", "abc", "missing.go")).rejects.toThrow(SourceNotFoundError);
  });

  it("throws generic error on non-OK status", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        new HttpResponse(null, { status: 500 }),
      ),
    );
    await expect(fetchSource("payment", "abc", "file.go")).rejects.toThrow("fetch source failed: 500");
  });
});

describe("fetchSourceWithContext", () => {
  it("returns content and coverage map", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        HttpResponse.json({
          content: "line1\nline2\nline3\n",
          coverage: {
            covered_lines: [1, 3],
            uncovered_lines: [2],
          },
        }),
      ),
    );
    const result = await fetchSourceWithContext("payment", "abc", "main.go", "cf-1");
    expect(result.content).toBe("line1\nline2\nline3\n");
    expect(result.coverage).toBeDefined();
    expect(result.coverage!.get(1)).toBe("covered");
    expect(result.coverage!.get(2)).toBe("uncovered");
    expect(result.coverage!.get(3)).toBe("covered");
  });

  it("returns undefined coverage when not present", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        HttpResponse.json({ content: "hello\n" }),
      ),
    );
    const result = await fetchSourceWithContext("payment", "abc", "main.go", "cf-1");
    expect(result.content).toBe("hello\n");
    expect(result.coverage).toBeUndefined();
  });

  it("throws SourceNotFoundError on 404", async () => {
    server.use(
      http.get("/api/v1/projects/:key/source", () =>
        new HttpResponse(null, { status: 404 }),
      ),
    );
    await expect(fetchSourceWithContext("payment", "abc", "x.go", "cf-1")).rejects.toThrow(SourceNotFoundError);
  });
});
