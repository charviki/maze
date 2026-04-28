import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './dialog';
import { Panel } from './Panel';
import { DecryptText } from './DecryptText';

export interface AnimationSettings {
  canvasBackground: boolean;
  crtScanlines: boolean;
  decryptText: boolean;
  glitchEffect: boolean;
  bootSequence: boolean;
}

const DEFAULT_SETTINGS: AnimationSettings = {
  canvasBackground: true,
  crtScanlines: true,
  decryptText: true,
  glitchEffect: true,
  bootSequence: true,
};

const STORAGE_KEY = 'maze:animation-settings';

interface AnimationSettingsContextValue {
  settings: AnimationSettings;
  setSettings: React.Dispatch<React.SetStateAction<AnimationSettings>>;
}

const AnimationSettingsContext = createContext<AnimationSettingsContextValue | null>(null);

// 从 localStorage 读取设置，解析失败则回退到默认值
function loadSettings(): AnimationSettings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      return { ...DEFAULT_SETTINGS, ...JSON.parse(raw) };
    }
  } catch {
    // JSON 解析失败，忽略并使用默认值
  }
  return DEFAULT_SETTINGS;
}

export function AnimationSettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<AnimationSettings>(loadSettings);

  // 设置变更时持久化到 localStorage
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  }, [settings]);

  // CRT 特效开关：通过 body 类名控制全局伪元素显隐
  useEffect(() => {
    if (settings.crtScanlines) {
      document.body.classList.remove('no-crt-effects');
    } else {
      document.body.classList.add('no-crt-effects');
    }
  }, [settings.crtScanlines]);

  return (
    <AnimationSettingsContext.Provider value={{ settings, setSettings }}>
      {children}
    </AnimationSettingsContext.Provider>
  );
}

export function useAnimationSettings() {
  const ctx = useContext(AnimationSettingsContext);
  if (!ctx) {
    throw new Error('useAnimationSettings must be used within AnimationSettingsProvider');
  }

  const updateSetting = useCallback(
    (key: keyof AnimationSettings, value: boolean) => {
      ctx.setSettings((prev) => ({ ...prev, [key]: value }));
    },
    [ctx],
  );

  return { settings: ctx.settings, updateSetting };
}

// 五个开关项的配置
const TOGGLE_ITEMS: { key: keyof AnimationSettings; label: string; description: string }[] = [
  { key: 'canvasBackground', label: 'TERRAIN RENDER', description: 'Canvas background animations' },
  { key: 'crtScanlines', label: 'CRT SCANLINES', description: 'CRT scanline and noise overlay' },
  { key: 'decryptText', label: 'DECRYPT FX', description: 'Text decryption animation effect' },
  { key: 'glitchEffect', label: 'GLITCH FX', description: 'Offline node glitch distortion' },
  { key: 'bootSequence', label: 'BOOT SEQUENCE', description: 'System startup animation' },
];

interface AnimationSettingsPanelProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AnimationSettingsPanel({ open, onOpenChange }: AnimationSettingsPanelProps) {
  const { settings, updateSetting } = useAnimationSettings();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>
            <DecryptText text="VISUAL EFFECTS CONFIG" />
          </DialogTitle>
        </DialogHeader>

        <Panel variant="default">
          <div className="flex flex-col gap-2">
            {TOGGLE_ITEMS.map(({ key, label, description }) => (
              <label
                key={key}
                className="flex items-center justify-between p-3 border border-primary/20 hover:border-primary/40 transition-colors cursor-pointer"
                style={{
                  clipPath:
                    'polygon(0 0, calc(100% - 6px) 0, 100% 6px, 100% 100%, 6px 100%, 0 calc(100% - 6px))',
                }}
              >
                <div>
                  <div className="text-xs font-mono uppercase tracking-widest text-primary">
                    {label}
                  </div>
                  <div className="text-[10px] text-muted-foreground mt-0.5">{description}</div>
                </div>
                <button
                  type="button"
                  onClick={() => updateSetting(key, !settings[key])}
                  className={`w-10 h-5 rounded-full transition-colors relative ${
                    settings[key] ? 'bg-primary' : 'bg-muted'
                  }`}
                >
                  <div
                    className={`absolute top-0.5 w-4 h-4 rounded-full bg-background transition-transform ${
                      settings[key] ? 'translate-x-5' : 'translate-x-0.5'
                    }`}
                  />
                </button>
              </label>
            ))}
          </div>
        </Panel>
      </DialogContent>
    </Dialog>
  );
}
