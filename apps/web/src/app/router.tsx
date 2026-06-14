import { lazy, Suspense } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { AppLayout } from "./layout/sidebar";
import { ProtectedRoute } from "@/shared/auth/protected-route";
import { LoginPage } from "@/pages/login";
import { Spinner } from "@/shared/ui/spinner";

const HomePage = lazy(() => import("@/pages/home").then(m => ({ default: m.HomePage })));
const ProjectDetailPage = lazy(() => import("@/pages/project-detail").then(m => ({ default: m.ProjectDetailPage })));
const GavelspacesPage = lazy(() => import("@/pages/gavelspaces").then(m => ({ default: m.GavelspacesPage })));
const GavelspaceDetailPage = lazy(() => import("@/pages/gavelspace-detail").then(m => ({ default: m.GavelspaceDetailPage })));
const GavelspaceOverviewTab = lazy(() => import("@/pages/gavelspace/overview").then(m => ({ default: m.OverviewTab })));
const GavelspaceProjectsTab = lazy(() => import("@/pages/gavelspace/projects").then(m => ({ default: m.ProjectsTab })));
const GavelspaceFindingsTab = lazy(() => import("@/pages/gavelspace/findings").then(m => ({ default: m.FindingsTab })));
const GavelspacePRChecksTab = lazy(() => import("@/pages/gavelspace/pr-checks").then(m => ({ default: m.PRChecksTab })));
const GavelspaceCaseFilesTab = lazy(() => import("@/pages/gavelspace/case-files").then(m => ({ default: m.CaseFilesTab })));
const FindingDetailPage = lazy(() => import("@/pages/gavelspace/finding-detail").then(m => ({ default: m.FindingDetailPage })));
const PRCheckDetailPage = lazy(() => import("@/pages/pr-check-detail").then(m => ({ default: m.PRCheckDetailPage })));
const ProfilePage = lazy(() => import("@/pages/profile").then(m => ({ default: m.ProfilePage })));
const CaseFileDetailPage = lazy(() => import("@/pages/casefile-detail").then(m => ({ default: m.CaseFileDetailPage })));
const CaseFileDiffPage = lazy(() => import("@/pages/casefile-diff").then(m => ({ default: m.CaseFileDiffPage })));
const TokensPage = lazy(() => import("@/pages/tokens").then(m => ({ default: m.TokensPage })));
const AdminUsersPage = lazy(() => import("@/pages/admin-users").then(m => ({ default: m.AdminUsersPage })));
const DesignSystemPage = lazy(() => import("@/pages/design-system").then(m => ({ default: m.DesignSystemPage })));
const NotFoundPage = lazy(() => import("@/pages/not-found").then(m => ({ default: m.NotFoundPage })));

function LazyFallback() {
  return <div className="flex flex-1 items-center justify-center p-6"><Spinner /></div>;
}

export function Router() {
  return (
    <BrowserRouter>
      <Suspense fallback={<LazyFallback />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<AppLayout />}>
              <Route path="/" element={<HomePage />} />
              <Route path="/projects/:key" element={<ProjectDetailPage />} />
              <Route path="/gavelspaces" element={<GavelspacesPage />} />
              <Route path="/gavelspaces/:name" element={<GavelspaceDetailPage />}>
                <Route index element={<GavelspaceOverviewTab />} />
                <Route path="projects" element={<GavelspaceProjectsTab />} />
                <Route path="findings" element={<GavelspaceFindingsTab />} />
                <Route path="findings/:findingId" element={<FindingDetailPage />} />
                <Route path="pr-checks" element={<GavelspacePRChecksTab />} />
                <Route path="case-files" element={<GavelspaceCaseFilesTab />} />
              </Route>
              <Route path="/pr-checks/:id" element={<PRCheckDetailPage />} />
              <Route path="/profile" element={<ProfilePage />} />
              <Route path="/casefiles/:id" element={<CaseFileDetailPage />} />
              <Route path="/casefiles/:id/diff/:compareId" element={<CaseFileDiffPage />} />
              <Route path="/tokens" element={<TokensPage />} />
              <Route path="/admin/users" element={<AdminUsersPage />} />
              <Route path="/design-system" element={<DesignSystemPage />} />
              <Route path="*" element={<NotFoundPage />} />
            </Route>
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
