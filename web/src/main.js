import { createInertiaApp } from '@inertiajs/svelte'
import { mount } from 'svelte'

createInertiaApp({
  // Cais uses double-submit cookie `cais_csrf` + header `X-CSRF-Token`
  // (Inertia defaults are XSRF-TOKEN / X-XSRF-TOKEN).
  http: {
    xsrfCookieName: 'cais_csrf',
    xsrfHeaderName: 'X-CSRF-Token',
  },
  resolve: (name) => {
    const pages = import.meta.glob('./pages/**/*.svelte', { eager: true })
    const page = pages[`./pages/${name}.svelte`]
    if (!page) {
      throw new Error(`Inertia page not found: ${name}`)
    }
    // App.svelte expects a module with `.default`
    return page
  },
  setup({ el, App, props }) {
    // Svelte 5: use mount(), not `new App(...)`
    mount(App, { target: el, props })
  },
})
