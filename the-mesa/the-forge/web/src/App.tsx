import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AppShell } from '@maze/fabrication';
import Layout from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import KnowledgeList from '@/pages/KnowledgeList';
import KnowledgeDetail from '@/pages/KnowledgeDetail';
import KnowledgeEdit from '@/pages/KnowledgeEdit';
import TaskList from '@/pages/TaskList';
import TaskDetail from '@/pages/TaskDetail';
import TaskEdit from '@/pages/TaskEdit';

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
    <AppShell>
      <BrowserRouter basename={import.meta.env.BASE_URL.replace(/\/$/, '')}>
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/knowledge" element={<KnowledgeList />} />
            <Route path="/knowledge/new" element={<KnowledgeEdit />} />
            <Route path="/knowledge/:id" element={<KnowledgeDetail />} />
            <Route path="/knowledge/:id/edit" element={<KnowledgeEdit />} />
            <Route path="/tasks" element={<TaskList />} />
            <Route path="/tasks/new" element={<TaskEdit />} />
            <Route path="/tasks/:id" element={<TaskDetail />} />
            <Route path="/tasks/:id/edit" element={<TaskEdit />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </Layout>
      </BrowserRouter>
    </AppShell>
  );
}

export default App;
