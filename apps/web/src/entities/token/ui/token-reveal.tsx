import { useState } from "react";
import type { CreateTokenResult } from "../model";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";
import { Copy, Check } from "lucide-react";

interface TokenRevealProps {
  result: CreateTokenResult;
  onDismiss: () => void;
}

export function TokenReveal({ result, onDismiss }: TokenRevealProps) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    await navigator.clipboard.writeText(result.token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Card className="border-success/30 bg-success/10">
      <CardContent className="pt-6">
        <p className="mb-2 text-sm font-medium text-success">
          Token created. Copy it now — it won't be shown again.
        </p>
        <div className="flex items-center gap-2">
          <code className="flex-1 rounded bg-background px-3 py-2 text-sm break-all border border-border">
            {result.token}
          </code>
          <Button variant="outline" size="icon" onClick={copy}>
            {copied ? (
              <Check className="h-4 w-4 text-success" />
            ) : (
              <Copy className="h-4 w-4" />
            )}
          </Button>
        </div>
        <Button variant="ghost" size="sm" className="mt-3" onClick={onDismiss}>
          Dismiss
        </Button>
      </CardContent>
    </Card>
  );
}
