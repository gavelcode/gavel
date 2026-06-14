import { ErrorBoundary } from "./error-boundary";
import { Providers } from "./providers";
import { Router } from "./router";

export default function App() {
  return (
    <ErrorBoundary>
      <Providers>
        <Router />
      </Providers>
    </ErrorBoundary>
  );
}
