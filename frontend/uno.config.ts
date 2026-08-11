import { FileSystemIconLoader } from '@iconify/utils/lib/loader/node-loaders'
import { defineConfig, presetIcons, presetWind3 } from 'unocss'

/**
 * Protobiont UI — UnoCSS config.
 * Utilities map onto the CSS custom properties defined in tokens.css.
 * There is NO safelist in this system, by design: every class a component can
 * emit exists as a complete literal string in source. If styles go missing,
 * fix the component — never add a safelist.
 */
export default defineConfig({
  presets: [
    presetWind3(),
    presetIcons({
      // 1rem, matching `w-4 h-4` - what call sites use for a non-default size.
      scale: 1,
      unit: 'rem',
      // icons resolve from local @iconify-json/* packages — no CDN
      // width/height don't apply to inline elements, so a bare icon span
      // renders at zero size without display: inline-block
      extraProperties: {
        display: 'inline-block',
        'vertical-align': 'middle',
      },
      collections: {
        // i-prt-* icons: drop an SVG in ./icons/, FileSystemIconLoader resolves
        // i-prt-<name> at build time and rewrites #fff/white → currentColor so
        // masks tint via text-*/seeds.
        prt: FileSystemIconLoader('./icons', (svg) =>
          svg.replace(/#fff|#ffffff|white/gi, 'currentColor'),
        ),
      },
    }),
  ],

  theme: {
    colors: {
      surface: {
        0: 'var(--surface-0)',
        1: 'var(--surface-1)',
        2: 'var(--surface-2)',
        3: 'var(--surface-3)',
      },
      ink: {
        DEFAULT: 'var(--ink)',
        muted: 'var(--ink-muted)',
        faint: 'var(--ink-faint)',
      },
      accent: {
        DEFAULT: 'var(--accent)',
        ink: 'var(--accent-ink)',
      },
      danger: {
        DEFAULT: 'var(--danger)',
        ink: 'var(--danger-ink)',
      },
      warning: {
        DEFAULT: 'var(--warning)',
        ink: 'var(--warning-ink)',
      },
      info: {
        DEFAULT: 'var(--info)',
        ink: 'var(--info-ink)',
      },
      edge: {
        DEFAULT: 'var(--edge)',
        strong: 'var(--edge-strong)',
      },
      wash: 'var(--wash)',
      // for labeling categories in data/tags, not a semantic color - never use on a button or input
      cat: {
        teal: 'var(--p-cat-teal)',
        purple: 'var(--p-cat-purple)',
        magenta: 'var(--p-cat-magenta)',
      },
    },
    fontFamily: {
      sans: 'var(--font-sans)',
      mono: 'var(--font-mono)',
    },
    borderRadius: {
      control: 'var(--radius-control)',
      surface: 'var(--radius-surface)',
      mark: 'var(--radius-mark)',
    },
    boxShadow: {
      raise: 'var(--shadow-raise)',
      float: 'var(--shadow-float)',
      overlay: 'var(--shadow-overlay)',
    },
  },

  shortcuts: {
    // the one motion curve — use this, never ad-hoc durations
    'prt-motion':
      'transition-all duration-[var(--motion-duration)] ease-[var(--motion-ease)]',
    'prt-motion-colors':
      'transition-colors duration-[var(--motion-duration)] ease-[var(--motion-ease)]',
    // the one focus ring
    'prt-ring':
      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-surface-0',
  },

  content: {
    pipeline: {
      include: ['**/*.{vue,ts,tsx,html}'],
    },
  },
})
