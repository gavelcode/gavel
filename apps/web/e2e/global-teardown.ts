export default async function globalTeardown(): Promise<void> {
  console.log("E2E global teardown: done (PostgreSQL container kept for reuse).");
}
