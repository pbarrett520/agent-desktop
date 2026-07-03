import { useState, useRef, useEffect } from 'react';

export interface DropdownOption {
  value: string;
  label: string;
}

interface DropdownProps {
  value: string;
  options: DropdownOption[];
  onChange: (value: string) => void;
  disabled?: boolean;
  className?: string;
  placeholder?: string;
}

// A hand-rolled dropdown. Native <select>/<option> popups are rendered by
// the OS/webview toolkit (GTK on Linux via WebKitGTK) and largely ignore
// page CSS, which is why they can show up as unreadable white-on-white in
// dark mode. Rendering the option list ourselves keeps it fully themeable.
export default function Dropdown({ value, options, onChange, disabled, className = '', placeholder }: DropdownProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const selected = options.find(o => o.value === value);

  return (
    <div ref={rootRef} className={`relative ${className}`}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen(prev => !prev)}
        className="w-full flex items-center justify-between gap-2 input-field text-xs disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <span className={`truncate ${selected ? 'text-white' : 'text-white/40'}`}>
          {selected?.label || placeholder || 'Select…'}
        </span>
        <svg className={`w-3.5 h-3.5 flex-shrink-0 text-white/40 transition-transform ${open ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {open && (
        <div className="absolute z-50 mt-1 w-full max-h-56 overflow-y-auto rounded-lg border border-brand-border bg-brand-panel shadow-lg">
          {options.map(opt => (
            <button
              key={opt.value}
              type="button"
              onClick={() => {
                onChange(opt.value);
                setOpen(false);
              }}
              className={`w-full text-left px-3 py-2 text-xs truncate transition-colors ${
                opt.value === value
                  ? 'bg-brand-cyan/15 text-brand-cyan'
                  : 'text-white/80 hover:bg-white/5 hover:text-white'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
