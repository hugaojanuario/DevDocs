import { useEffect } from "react";
import {
  Routes,
  Route,
  Navigate,
  useLocation,
  useNavigate,
  Outlet,
} from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { session, ApiError, setSession } from "./services/http";
import { AppShell } from "./components/layout/AppShell";
import { ResultSkeleton } from "./components/documentation/ResultSkeleton";
import { Notice } from "./components/documentation/Notice";
import { Button } from "./components/ui/button";
import { Home } from "./pages/Home";
import { SearchPage } from "./pages/Search";
import { SavedPage } from "./pages/Saved";
import { SavedDetail } from "./pages/SavedDetail";
import { TechnologiesPage } from "./pages/Technologies";
import { SettingsPage } from "./pages/Settings";
import { LoginPage } from "./pages/Login";
function Authenticated() {
  const result = useQuery({
    queryKey: ["session"],
    queryFn: ({ signal }) => session(signal),
    retry: false,
    staleTime: 60_000,
  });
  const location = useLocation();
  if (result.isPending)
    return (
      <main className="dd-auth-loading">
        <ResultSkeleton />
      </main>
    );
  if (result.error) {
    if (result.error instanceof ApiError && result.error.status === 401)
      return (
        <Navigate
          to={
            "/login?" +
            new URLSearchParams({ next: location.pathname + location.search })
          }
          replace
        />
      );
    return (
      <main className="dd-auth-loading">
        <Notice
          tone="offline"
          title="Não foi possível conectar ao DevDocs"
          action={
            <Button onClick={() => void result.refetch()}>
              Tentar novamente
            </Button>
          }
        >
          {result.error.message}
        </Notice>
      </main>
    );
  }
  return <Outlet />;
}
export default function App() {
  const navigate = useNavigate();
  const location = useLocation();
  const client = useQueryClient();
  useEffect(() => {
    const handler = () => {
      setSession(null);
      void client.cancelQueries();
      client.clear();
      navigate(
        "/login?" +
          new URLSearchParams({ next: location.pathname + location.search }),
        { replace: true },
      );
    };
    window.addEventListener("devdocs:unauthorized", handler);
    return () => window.removeEventListener("devdocs:unauthorized", handler);
  }, [navigate, location, client]);
  useEffect(() => {
    document.title =
      (location.pathname === "/saved"
        ? "Consultas salvas"
        : location.pathname === "/technologies"
          ? "Tecnologias"
          : location.pathname === "/settings"
            ? "Configurações"
            : "Documentação técnica") + " — DevDocs";
  }, [location.pathname]);
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<Authenticated />}>
        <Route element={<AppShell />}>
          <Route index element={<Home />} />
          <Route path="search" element={<SearchPage />} />
          <Route path="saved" element={<SavedPage />} />
          <Route path="saved/:id" element={<SavedDetail />} />
          <Route path="technologies" element={<TechnologiesPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Route>
    </Routes>
  );
}
