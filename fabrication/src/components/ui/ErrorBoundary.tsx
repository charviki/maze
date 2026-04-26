import { Component, type ReactNode, type ErrorInfo } from 'react';
import { Panel } from './Panel';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  // 捕获错误详情用于调试日志，不暴露给用户
  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('[MAZE] SYSTEM FAULT:', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="fixed inset-0 bg-background flex items-center justify-center p-8">
          <Panel variant="destructive" className="max-w-lg w-full" cornerSize={16}>
            <div className="text-center space-y-4">
              <div className="text-destructive text-2xl font-mono font-bold tracking-widest uppercase">
                SYSTEM FAULT
              </div>
              <div className="text-[10px] text-destructive/60 font-mono tracking-widest">
                // CRITICAL ERROR DETECTED //
              </div>
              <div className="bg-destructive/10 border border-destructive/30 p-3 text-xs font-mono text-destructive/80 max-h-32 overflow-y-auto text-left break-all">
                {this.state.error?.message || 'UNKNOWN ERROR'}
              </div>
              <button
                onClick={this.handleRetry}
                className="text-xs font-mono tracking-widest uppercase px-4 py-2 border border-primary/50 text-primary hover:bg-primary/10 transition-colors"
              >
                [ RETRY ]
              </button>
            </div>
          </Panel>
        </div>
      );
    }

    return this.props.children;
  }
}
