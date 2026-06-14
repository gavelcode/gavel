import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from "@tanstack/react-table";
import type { Finding } from "../model";
import { FindingBadge } from "./finding-badge";
import { Spinner } from "@/shared/ui/spinner";
import { EmptyState } from "@/shared/ui/empty-state";

const columnHelper = createColumnHelper<Finding>();

const columns = [
  columnHelper.accessor("status", {
    header: "Status",
    cell: (info) => <FindingBadge label={info.getValue()} kind="status" />,
  }),
  columnHelper.accessor("severity", {
    header: "Severity",
    cell: (info) => <FindingBadge label={info.getValue()} kind="severity" />,
  }),
  columnHelper.accessor("tool", { header: "Tool" }),
  columnHelper.accessor("ruleId", { header: "Rule" }),
  columnHelper.accessor("filePath", {
    header: "File",
    cell: (info) => {
      const row = info.row.original;
      const line = row.line ? `:${row.line}` : "";
      return (
        <code className="text-xs">
          {info.getValue()}{line}
        </code>
      );
    },
  }),
  columnHelper.accessor("message", {
    header: "Message",
    cell: (info) => (
      <span className="text-xs text-muted-foreground line-clamp-2">{info.getValue()}</span>
    ),
  }),
];

interface FindingsTableProps {
  findings: Finding[];
  isLoading?: boolean;
}

export function FindingsTable({ findings, isLoading }: FindingsTableProps) {
  const table = useReactTable({
    data: findings,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  if (isLoading) {
    return <Spinner />;
  }

  if (findings.length === 0) {
    return <EmptyState title="No findings" description="No findings to display." />;
  }

  return (
    <table className="w-full text-sm">
      <thead>
        {table.getHeaderGroups().map((hg) => (
          <tr key={hg.id} className="border-b">
            {hg.headers.map((h) => (
              <th
                key={h.id}
                className="px-4 py-3 text-left font-medium text-muted-foreground"
              >
                {flexRender(h.column.columnDef.header, h.getContext())}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.map((row) => (
          <tr key={row.id} className="border-b last:border-0">
            {row.getVisibleCells().map((cell) => (
              <td key={cell.id} className="px-4 py-3">
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
