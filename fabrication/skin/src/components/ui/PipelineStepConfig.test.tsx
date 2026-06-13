import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { PipelineStepConfig, type PipelineConfigStep } from './PipelineStepConfig';

describe('PipelineStepConfig', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
  });

  it('renders empty state with add row', () => {
    render(<PipelineStepConfig steps={[]} onChange={mockOnChange} />);
    expect(screen.getByPlaceholderText('输入 prompt 内容...')).toBeInTheDocument();
    expect(screen.getByText('添加')).toBeInTheDocument();
  });

  it('adds a prompt step via button click', () => {
    render(<PipelineStepConfig steps={[]} onChange={mockOnChange} />);
    const input = screen.getByPlaceholderText('输入 prompt 内容...');
    fireEvent.change(input, { target: { value: 'hello world' } });
    const addButton = screen.getByText('添加');
    fireEvent.click(addButton);
    expect(mockOnChange).toHaveBeenCalledTimes(1);
    expect(mockOnChange).toHaveBeenCalledWith([
      expect.objectContaining({
        type: 'CONFIG_ITEM_TYPE_PROMPT',
        value: 'hello world',
      }),
    ]);
  });

  it('adds a prompt step via Enter key', () => {
    render(<PipelineStepConfig steps={[]} onChange={mockOnChange} />);
    const input = screen.getByPlaceholderText('输入 prompt 内容...');
    fireEvent.change(input, { target: { value: 'test prompt' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(mockOnChange).toHaveBeenCalledWith([
      expect.objectContaining({
        type: 'CONFIG_ITEM_TYPE_PROMPT',
        value: 'test prompt',
      }),
    ]);
  });

  it('renders existing steps', () => {
    const steps: PipelineConfigStep[] = [
      { id: '1', type: 'CONFIG_ITEM_TYPE_PROMPT', value: 'implement feature X' },
      { id: '2', type: 'CONFIG_ITEM_TYPE_PROMPT', value: 'run tests' },
    ];
    render(<PipelineStepConfig steps={steps} onChange={mockOnChange} />);
    expect(screen.getByText('implement feature X')).toBeInTheDocument();
    expect(screen.getByText('run tests')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('button is disabled when input is empty', () => {
    render(<PipelineStepConfig steps={[]} onChange={mockOnChange} />);
    const addButton = screen.getByText('添加');
    expect(addButton.closest('button')).toBeDisabled();
  });

  it('button is enabled when input has value', () => {
    render(<PipelineStepConfig steps={[]} onChange={mockOnChange} />);
    const input = screen.getByPlaceholderText('输入 prompt 内容...');
    fireEvent.change(input, { target: { value: 'something' } });
    const addButton = screen.getByText('添加');
    expect(addButton.closest('button')).not.toBeDisabled();
  });

  it('removes a step', () => {
    const steps: PipelineConfigStep[] = [
      { id: '1', type: 'CONFIG_ITEM_TYPE_PROMPT', value: 'test' },
    ];
    render(<PipelineStepConfig steps={steps} onChange={mockOnChange} />);
    // Find the delete button (has Trash2 icon)
    const allButtons = screen.getAllByRole('button');
    // The delete button is the one with the trash icon (not the add button)
    const deleteButton = allButtons.find(
      (b) => b.innerHTML.includes('trash-2') || b.querySelector('.lucide-trash-2'),
    );
    expect(deleteButton).toBeTruthy();
    fireEvent.click(deleteButton!);
    expect(mockOnChange).toHaveBeenCalledWith([]);
  });

  it('clears input after adding', () => {
    render(<PipelineStepConfig steps={[]} onChange={mockOnChange} />);
    const input = screen.getByPlaceholderText('输入 prompt 内容...');
    fireEvent.change(input, { target: { value: 'test' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect((input as HTMLInputElement).value).toBe('');
  });
});
