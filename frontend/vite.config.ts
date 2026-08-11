import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import Icons from 'unplugin-icons/vite'
import { FileSystemIconLoader } from 'unplugin-icons/loaders'

export default defineConfig({
  base: '/dashboard/',
  plugins: [
    vue(),
    UnoCSS(),
    Icons({
      compiler: 'vue3',
      // `~icons/prt/*` reads frontend/icons/. Same namespace uno.config.ts
      // gives the kit's `i-prt-*` classes, so a house icon is reachable both
      // ways and only has to be drawn once.
      customCollections: {
        prt: FileSystemIconLoader('./icons', svg =>
          svg.replace(/#fff|#ffffff|white/gi, 'currentColor'),
        ),
      },
      // Iconify emits width/height of 1em. em is not used in this codebase - it
      // resizes with whatever font-size the icon lands in, so the same icon
      // comes out different in a button and in a table cell.
      iconCustomizer(_collection, _icon, props) {
        props.width = '1rem'
        props.height = '1rem'
      },
    }),
  ],
  // vue-tsc reads the matching "paths" entry in tsconfig.json; both are needed.
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:9000',
      // Proxy S3 object paths (/{bucket}/{key}). Without the lookahead this also
      // swallows Vite's own dev paths and /dashboard, which the backend serves from
      // the embedded build.
      '^/(?!@|src|node_modules|dashboard)[^/]+/.+': {
        target: 'http://localhost:9000',
        changeOrigin: true,
      },
    }
  }
})
