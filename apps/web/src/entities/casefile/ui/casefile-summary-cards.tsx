import type { CaseFile } from "../model";
import { Card, CardHeader, CardTitle, CardContent } from "@/shared/ui/card";

interface CaseFileSummaryCardsProps {
  caseFile: CaseFile;
}

export function CaseFileSummaryCards({ caseFile }: CaseFileSummaryCardsProps) {
  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
      <Card elevated>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">Total</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold">{caseFile.totalFindings}</p>
        </CardContent>
      </Card>
      <Card elevated>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">New</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold text-danger">{caseFile.newFindings}</p>
        </CardContent>
      </Card>
      <Card elevated>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">Existing</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold">{caseFile.existingFindings}</p>
        </CardContent>
      </Card>
      <Card elevated>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">Resolved</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl font-bold text-success">{caseFile.resolvedFindings}</p>
        </CardContent>
      </Card>
    </div>
  );
}
