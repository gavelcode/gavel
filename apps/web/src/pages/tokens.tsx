import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  createColumnHelper,
} from "@tanstack/react-table";
import type { Token, CreateTokenResult } from "@/entities/token/model";
import * as tokenApi from "@/entities/token/api";
import { CreateTokenForm } from "@/entities/token/ui/create-token-form";
import { TokenReveal } from "@/entities/token/ui/token-reveal";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { TopBar } from "@/shared/ui/top-bar";
import { Trash2, Plus } from "lucide-react";

const columnHelper = createColumnHelper<Token>();

const SCOPE_COLORS: Record<string, string> = {
  ingest: "bg-primary/15 text-primary",
  read: "bg-success/15 text-success",
  admin: "bg-danger/15 text-danger",
};

function formatDate(iso: string | undefined) {
  if (!iso) return "-";
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function TokensPage() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [newToken, setNewToken] = useState<CreateTokenResult | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const { data: tokens = [], isLoading } = useQuery({
    queryKey: ["tokens"],
    queryFn: () => tokenApi.listTokens(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tokenApi.deleteToken(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tokens"] });
      setDeleteConfirm(null);
    },
  });

  const columns = [
    columnHelper.accessor("name", { header: "Name" }),
    columnHelper.accessor("prefix", { header: "Prefix" }),
    columnHelper.accessor("scopes", {
      header: "Scopes",
      cell: (info) => (
        <div className="flex flex-wrap gap-1">
          {info.getValue().map((scope) => (
            <span key={scope} className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SCOPE_COLORS[scope] ?? ""}`}>
              {scope}
            </span>
          ))}
        </div>
      ),
    }),
    columnHelper.accessor("createdAt", {
      header: "Created",
      cell: (info) => formatDate(info.getValue()),
    }),
    columnHelper.accessor("lastUsedAt", {
      header: "Last used",
      cell: (info) => {
        const v = info.getValue();
        return v ? formatDate(v) : <span className="text-muted-foreground">never</span>;
      },
    }),
    columnHelper.accessor("expiresAt", {
      header: "Expires",
      cell: (info) => formatDate(info.getValue()),
    }),
    columnHelper.display({
      id: "actions",
      cell: (info) => {
        const id = info.row.original.id;
        if (deleteConfirm === id) {
          return (
            <div className="flex items-center gap-2">
              <Button
                variant="destructive"
                size="sm"
                onClick={() => deleteMutation.mutate(id)}
                disabled={deleteMutation.isPending}
              >
                Confirm
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setDeleteConfirm(null)}
              >
                Cancel
              </Button>
            </div>
          );
        }
        return (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setDeleteConfirm(id)}
          >
            <Trash2 className="h-4 w-4 text-muted-foreground" />
          </Button>
        );
      },
    }),
  ];

  const table = useReactTable({
    data: tokens,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div className="flex flex-col">
      <TopBar
        crumbs={["API Tokens"]}
        action={
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create token
          </Button>
        }
      />
      <div className="flex flex-col gap-d-gap overflow-auto p-d-page">

      {showCreate && (
        <CreateTokenForm
          onCreated={(result) => {
            setNewToken(result);
            setShowCreate(false);
            void queryClient.invalidateQueries({ queryKey: ["tokens"] });
          }}
          onCancel={() => setShowCreate(false)}
        />
      )}

      {newToken && <TokenReveal result={newToken} onDismiss={() => setNewToken(null)} />}

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <Spinner />
          ) : tokens.length === 0 ? (
            <p className="p-6 text-muted-foreground">
              No API tokens yet. Create one to get started.
            </p>
          ) : (
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
          )}
        </CardContent>
      </Card>
      </div>
    </div>
  );
}
