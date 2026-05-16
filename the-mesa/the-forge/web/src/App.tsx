import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AppShell, ErrorBoundary } from '@maze/fabrication';
import Layout from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import DocDetail from '@/pages/DocDetail';
import DocEdit from '@/pages/DocEdit';

function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center h-[60vh] text-center">
      <h1 className="text-foreground font-mono font-bold text-4xl mb-4">404</h1>
      <p className="text-muted-foreground font-mono text-sm">Page not found</p>
    </div>
  );
}

function App() {
  return (
    <AppShell requireAuth loginUrl="/arrival-gate/">
      <BrowserRouter basename={import.meta.env.BASE_URL.replace(/\/$/, '')}>
        <ErrorBoundary>
          <Layout>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/docs/:archiveId" element={<DocDetail />} />
              <Route path="/docs/:archiveId/new" element={<DocEdit />} />
              <Route path="/docs/:archiveId/:parentId/new" element={<DocEdit />} />
              <Route path="/docs/:archiveId/:docId" element={<DocDetail />} />
              <Route path="/docs/:archiveId/:docId/edit" element={<DocEdit />} />
              <Route path="*" element={<NotFound />} />
            </Routes>
          </Layout>
        </ErrorBoundary>
      </BrowserRouter>
    </AppShell>
  );
}

export default App;
