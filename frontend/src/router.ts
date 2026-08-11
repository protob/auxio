import { createRouter, createWebHistory } from 'vue-router'
import BucketsView from './views/BucketsView.vue'
import ObjectsView from './views/ObjectsView.vue'

const router = createRouter({
  history: createWebHistory('/dashboard/'),
  routes: [
    { path: '/', name: 'buckets', component: BucketsView },
    // Lazy: Settings is read-only and rarely opened; 404 only fires when
    // nothing else matches, which is rare.
    { path: '/settings', name: 'settings', component: () => import('./views/SettingsView.vue') },
    { path: '/:bucket', name: 'browser', component: ObjectsView, props: true },
    // Only reachable for multi-segment paths - '/:bucket' claims every
    // single-segment one, and a non-existent bucket is the browser's own empty
    // state rather than a routing error.
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('./views/NotFoundView.vue') },
  ],
})

export default router
