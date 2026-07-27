<script>
  import { inertia } from '@inertiajs/svelte'

  /** @type {'dashboard'|'lots'|'sales'|'cash'|'payables'|'receivables'|''} */
  export let active = ''
  export let title = 'AuctionHQ'
  export let showLogout = true
  export let hideBottomNav = false

  const tabs = [
    { href: '/dashboard', key: 'dashboard', label: 'Dashboard', icon: 'dashboard' },
    { href: '/lots', key: 'lots', label: 'Lotes', icon: 'gavel' },
    { href: '/sales', key: 'sales', label: 'Vendas', icon: 'sell' },
    { href: '/cash', key: 'cash', label: 'Finanças', icon: 'analytics' },
  ]

  function isActive(key) {
    if (active === key) return true
    if (key === 'cash' && (active === 'payables' || active === 'receivables')) return true
    return false
  }
</script>

<div class="min-h-screen bg-background font-body-md text-body-md text-on-surface">
  <!-- Top bar -->
  <header
    class="fixed top-0 left-0 right-0 z-50 h-16 px-container-margin flex items-center justify-between
      bg-surface-container-lowest border-b border-outline-variant"
  >
    <div class="flex items-center gap-3 min-w-0">
      <div
        class="w-9 h-9 rounded-full bg-primary text-on-primary flex items-center justify-center shrink-0 font-bold text-sm"
      >
        AQ
      </div>
      <span class="font-headline-md text-headline-md font-bold text-primary truncate">{title}</span>
    </div>
    <div class="flex items-center gap-1">
      {#if showLogout}
        <form method="post" action="/logout" use:inertia>
          <button
            type="submit"
            class="w-10 h-10 flex items-center justify-center rounded-full text-on-surface-variant
              hover:bg-surface-container-low active:scale-95 transition-all"
            title="Sair"
            aria-label="Sair"
          >
            <span class="material-symbols-outlined">logout</span>
          </button>
        </form>
      {/if}
    </div>
  </header>

  <main class="pt-20 {hideBottomNav ? 'pb-8' : 'pb-28'} px-container-margin max-w-5xl mx-auto">
    <slot />
  </main>

  {#if !hideBottomNav}
    <nav
      class="fixed bottom-0 left-0 right-0 z-50 h-20 px-2 pb-2 flex justify-around items-center
        bg-surface-container-lowest border-t border-outline-variant shadow-float"
    >
      {#each tabs as tab}
        <a
          href={tab.href}
          use:inertia
          class="flex flex-col items-center justify-center min-w-[4.5rem] px-4 py-1.5 rounded-full transition-all
            {isActive(tab.key)
            ? 'bg-secondary-container text-on-secondary-container'
            : 'text-on-surface-variant hover:bg-surface-container-high active:scale-90'}"
        >
          <span class="material-symbols-outlined {isActive(tab.key) ? 'fill' : ''}">{tab.icon}</span>
          <span class="font-label-md text-label-md mt-0.5">{tab.label}</span>
        </a>
      {/each}
    </nav>
  {/if}
</div>
