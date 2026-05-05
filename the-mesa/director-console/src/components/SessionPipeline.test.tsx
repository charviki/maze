import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SessionPipeline } from '@maze/fabrication';
import type { PipelineStep } from '@maze/fabrication';

const systemStep: PipelineStep = {
  id: 'sys-cd',
  type: 'cd',
  phase: 'system',
  order: 0,
  key: '/home/agent',
  value: '',
};

const systemEnvStep: PipelineStep = {
  id: 'sys-env-FOO',
  type: 'env',
  phase: 'system',
  order: 1,
  key: 'FOO',
  value: 'bar',
};

const templateStep: PipelineStep = {
  id: 'tpl-command',
  type: 'command',
  phase: 'template',
  order: 2,
  key: '',
  value: 'claude --dangerously-skip-permissions',
};

const userStep: PipelineStep = {
  id: 'usr-cmd-1',
  type: 'command',
  phase: 'user',
  order: 3,
  key: '',
  value: 'echo hello',
};

const allSteps: PipelineStep[] = [systemStep, systemEnvStep, templateStep, userStep];

describe('SessionPipeline', () => {
  describe('三段管线渲染', () => {
    it('渲染 system 步骤（只读，显示锁图标）', () => {
      render(<SessionPipeline steps={[systemStep]} onChange={() => {}} />);

      expect(screen.getByText('切换目录')).toBeInTheDocument();
      expect(screen.getByText(/cd \/home\/agent/)).toBeInTheDocument();
      // system 层标题
      expect(screen.getByText('系统步骤')).toBeInTheDocument();
      // 只读标记
      expect(screen.getByText('只读')).toBeInTheDocument();
    });

    it('渲染 template 步骤（可编辑）', () => {
      render(<SessionPipeline steps={[templateStep]} onChange={() => {}} />);

      expect(screen.getByText('执行命令')).toBeInTheDocument();
      expect(screen.getByText('claude --dangerously-skip-permissions')).toBeInTheDocument();
      expect(screen.getByText('模板步骤')).toBeInTheDocument();
    });

    it('渲染 user 步骤（可添加和删除）', () => {
      render(<SessionPipeline steps={[userStep]} onChange={() => {}} />);

      expect(screen.getByText('用户步骤')).toBeInTheDocument();
      expect(screen.getByText('echo hello')).toBeInTheDocument();
    });

    it('readOnly 模式不显示添加按钮', () => {
      render(<SessionPipeline steps={allSteps} onChange={() => {}} readOnly />);

      expect(screen.queryByPlaceholderText('输入自定义 shell 命令...')).not.toBeInTheDocument();
    });

    it('非 readOnly 模式显示添加命令输入框', () => {
      render(<SessionPipeline steps={allSteps} onChange={() => {}} />);

      expect(screen.getByPlaceholderText('输入自定义 shell 命令...')).toBeInTheDocument();
      expect(screen.getByText('添加命令')).toBeInTheDocument();
    });
  });

  describe('添加用户命令', () => {
    it('输入命令并点击添加按钮，触发 onChange 新增 user 步骤', () => {
      const onChange = vi.fn();
      render(<SessionPipeline steps={[systemStep, templateStep]} onChange={onChange} />);

      const input = screen.getByPlaceholderText('输入自定义 shell 命令...');
      fireEvent.change(input, { target: { value: 'ls -la' } });

      const btn = screen.getByText('添加命令');
      fireEvent.click(btn);

      expect(onChange).toHaveBeenCalledTimes(1);
      const newSteps = onChange.mock.calls[0][0];
      // 新增了一个 user 层的 command 步骤
      const added = newSteps.find((s: PipelineStep) => s.phase === 'user' && s.value === 'ls -la');
      expect(added).toBeDefined();
      expect(added.type).toBe('command');
    });

    it('空命令不触发 onChange', () => {
      const onChange = vi.fn();
      render(<SessionPipeline steps={[systemStep]} onChange={onChange} />);

      const btn = screen.getByText('添加命令');
      fireEvent.click(btn);

      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe('模板命令编辑', () => {
    it('点击模板命令文本进入编辑模式', () => {
      render(<SessionPipeline steps={[templateStep]} onChange={() => {}} />);

      // 点击命令文本
      const cmdText = screen.getByText('claude --dangerously-skip-permissions');
      fireEvent.click(cmdText);

      // 应出现编辑输入框
      const input = screen.getByDisplayValue('claude --dangerously-skip-permissions');
      expect(input).toBeInTheDocument();
    });

    it('编辑模板命令后确认，触发 onChange 更新值', () => {
      const onChange = vi.fn();
      render(<SessionPipeline steps={[templateStep]} onChange={onChange} />);

      // 点击进入编辑
      const cmdText = screen.getByText('claude --dangerously-skip-permissions');
      fireEvent.click(cmdText);

      // 修改值
      const input = screen.getByDisplayValue('claude --dangerously-skip-permissions');
      fireEvent.change(input, { target: { value: 'claude --resume abc' } });

      // 点击确认
      const confirmBtn = screen.getByText('确认');
      fireEvent.click(confirmBtn);

      expect(onChange).toHaveBeenCalledTimes(1);
      const updated = onChange.mock.calls[0][0];
      const tplStep = updated.find((s: PipelineStep) => s.id === 'tpl-command');
      expect(tplStep.value).toBe('claude --resume abc');
    });
  });
});
