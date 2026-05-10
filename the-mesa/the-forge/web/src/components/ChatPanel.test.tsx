import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import ChatPanel from './ChatPanel';

function renderChatPanel() {
  return render(
    <MemoryRouter>
      <ChatPanel />
    </MemoryRouter>,
  );
}

describe('ChatPanel', () => {
  it('renders chat panel with input area', () => {
    renderChatPanel();

    expect(screen.getByText('ORACLE')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Ask the Oracle...')).toBeInTheDocument();
  });

  it('renders send button', () => {
    renderChatPanel();

    const sendButton = screen.getByRole('button');
    expect(sendButton).toBeInTheDocument();
  });

  it('send button is disabled when input is empty', () => {
    renderChatPanel();

    const sendButton = screen.getByRole('button');
    expect(sendButton).toBeDisabled();
  });

  it('send button becomes enabled when input has text', async () => {
    const user = userEvent.setup();
    renderChatPanel();

    const input = screen.getByPlaceholderText('Ask the Oracle...');
    await user.type(input, 'Hello');

    const sendButton = screen.getByRole('button');
    expect(sendButton).not.toBeDisabled();
  });

  it('renders empty state message', () => {
    renderChatPanel();

    expect(screen.getByText('Ask the Oracle about your knowledge base')).toBeInTheDocument();
  });
});
