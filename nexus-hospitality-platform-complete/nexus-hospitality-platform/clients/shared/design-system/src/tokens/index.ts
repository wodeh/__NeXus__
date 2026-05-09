// clients/shared/design-system/src/tokens/index.ts
// ============================================================
// NEXUS DESIGN SYSTEM — Platform-wide design tokens
// Supports: web, mobile, TV, staff console, executive dashboards
// ============================================================

export const colors = {
  // Brand palette
  brand: {
    primary: '#0A2540',
    secondary: '#00D4AA',
    accent: '#FF6B6B',
    highlight: '#FFD93D',
  },
  // Semantic colors
  semantic: {
    success: '#00C853',
    warning: '#FFAB00',
    error: '#FF1744',
    info: '#2979FF',
    neutral: '#78909C',
  },
  // Hospitality-specific
  hospitality: {
    vip: '#D4AF37',
    premium: '#8B5CF6',
    standard: '#64748B',
    maintenance: '#F97316',
    occupied: '#EF4444',
    vacant: '#22C55E',
    cleaning: '#3B82F6',
  },
  // Dark mode (TV/room displays)
  dark: {
    background: '#0F172A',
    surface: '#1E293B',
    elevated: '#334155',
    text: '#F8FAFC',
    textMuted: '#94A3B8',
    border: '#475569',
  },
  // Light mode (staff/web)
  light: {
    background: '#FFFFFF',
    surface: '#F1F5F9',
    elevated: '#E2E8F0',
    text: '#0F172A',
    textMuted: '#64748B',
    border: '#CBD5E1',
  },
} as const;

export const typography = {
  // Scale based on 1.25 major third
  scale: {
    xs: '0.75rem',    // 12px
    sm: '0.875rem',   // 14px
    base: '1rem',     // 16px
    lg: '1.125rem',   // 18px
    xl: '1.25rem',    // 20px
    '2xl': '1.5rem',  // 24px
    '3xl': '1.875rem',// 30px
    '4xl': '2.25rem', // 36px
    '5xl': '3rem',    // 48px
    '6xl': '3.75rem', // 60px
  },
  // TV-optimized (larger, higher contrast)
  tv: {
    xs: '1rem',
    sm: '1.125rem',
    base: '1.25rem',
    lg: '1.5rem',
    xl: '1.875rem',
    '2xl': '2.25rem',
    '3xl': '3rem',
    '4xl': '3.75rem',
  },
  fontFamily: {
    sans: '"Inter", "SF Pro Display", -apple-system, sans-serif',
    mono: '"JetBrains Mono", "Fira Code", monospace',
    display: '"Playfair Display", serif',
  },
  weight: {
    light: 300,
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
  },
} as const;

export const spacing = {
  // 4px base grid
  0: '0',
  1: '0.25rem',   // 4px
  2: '0.5rem',    // 8px
  3: '0.75rem',   // 12px
  4: '1rem',      // 16px
  5: '1.25rem',   // 20px
  6: '1.5rem',    // 24px
  8: '2rem',      // 32px
  10: '2.5rem',   // 40px
  12: '3rem',     // 48px
  16: '4rem',     // 64px
  20: '5rem',     // 80px
  24: '6rem',     // 96px
} as const;

export const breakpoints = {
  // Mobile-first breakpoints
  sm: '640px',
  md: '768px',
  lg: '1024px',
  xl: '1280px',
  '2xl': '1536px',
  // TV-specific
  tv: '1920px',
  'tv-4k': '3840px',
} as const;

export const shadows = {
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  DEFAULT: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)',
  // TV glow effects
  tv: '0 0 20px rgba(0, 212, 170, 0.3)',
  'tv-lg': '0 0 40px rgba(0, 212, 170, 0.4)',
} as const;

export const animations = {
  duration: {
    fast: '150ms',
    normal: '250ms',
    slow: '350ms',
    slower: '500ms',
  },
  easing: {
    default: 'cubic-bezier(0.4, 0, 0.2, 1)',
    in: 'cubic-bezier(0.4, 0, 1, 1)',
    out: 'cubic-bezier(0, 0, 0.2, 1)',
    bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
  },
} as const;

export const zIndex = {
  base: 0,
  dropdown: 1000,
  sticky: 1020,
  fixed: 1030,
  modalBackdrop: 1040,
  modal: 1050,
  popover: 1060,
  tooltip: 1070,
  toast: 1080,
  overlay: 1090,
} as const;

// Accessibility
export const a11y = {
  focusRing: {
    width: '2px',
    offset: '2px',
    color: colors.brand.secondary,
  },
  minTouchTarget: '44px',
  reducedMotion: '@media (prefers-reduced-motion: reduce)',
} as const;

export type ColorToken = keyof typeof colors;
export type SpacingToken = keyof typeof spacing;
