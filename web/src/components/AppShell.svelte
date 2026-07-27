<script>
  import { inertia, router } from '@inertiajs/svelte'

  /** @type {'dashboard'|'lots'|'stock'|'sales'|'cash'|'payables'|'receivables'|'config'|''} */
  export let active = ''
  export let title = ''
  export let companyName = 'AuctionHQ'
  export let showLogout = true
  export let hideBottomNav = false

  $: brand = title || companyName || 'AuctionHQ'

  function logout() {
    router.post('/logout')
  }

  const tabs = [
    { href: '/dashboard', key: 'dashboard', label: 'Dashboard', icon: 'dashboard' },
    { href: '/lots', key: 'lots', label: 'Lotes', icon: 'gavel' },
    { href: '/stock', key: 'stock', label: 'Estoque', icon: 'inventory_2' },
    { href: '/sales', key: 'sales', label: 'Vendas', icon: 'sell' },
    { href: '/cash', key: 'cash', label: 'Finanças', icon: 'analytics' },
  ]

  function isActive(key) {
    if (active === key) return true
    if (key === 'cash' && (active === 'payables' || active === 'receivables')) return true
    return false
  }

  function initials(name) {
    const parts = String(name || 'AQ')
      .trim()
      .split(/\s+/)
      .filter(Boolean)
    if (parts.length === 0) return 'AQ'
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }

  const pageTitles = {
    dashboard: 'Dashboard',
    lots: 'Lotes',
    stock: 'Estoque',
    sales: 'Vendas',
    cash: 'Caixa',
    payables: 'A pagar',
    receivables: 'A receber',
    config: 'Configurações',
  }

  $: docTitle = pageTitles[active]
    ? `${pageTitles[active]} · ${brand}`
    : brand
</script>

<svelte:head>
  <title>{docTitle}</title>
</svelte:head>

<div class="min-h-screen bg-background font-body-md text-body-md text-on-surface md:flex">
  <!-- Desktop sidebar -->
  {#if !hideBottomNav}
    <aside
      class="hidden md:flex md:flex-col md:w-56 md:shrink-0 md:fixed md:inset-y-0 md:left-0 z-50
        bg-surface-container-lowest border-r border-outline-variant"
    >
      <a href="/dashboard" use:inertia class="h-16 px-4 flex items-center gap-3 border-b border-outline-variant min-w-0">
        <div
          class="w-9 h-9 rounded-full bg-primary text-on-primary flex items-center justify-center shrink-0 font-bold text-sm"
        >
          {initials(brand)}
        </div>
        <span class="font-headline-md text-headline-md font-bold text-primary truncate">{brand}</span>
      </a>

      <nav class="flex-1 p-3 flex flex-col gap-1 overflow-y-auto" aria-label="Navegação principal">
        {#each tabs as tab}
          <a
            href={tab.href}
            use:inertia
            class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all
              {isActive(tab.key)
              ? 'bg-secondary-container text-on-secondary-container font-semibold'
              : 'text-on-surface-variant hover:bg-surface-container-high'}"
          >
            <span class="material-symbols-outlined {isActive(tab.key) ? 'fill' : ''}">{tab.icon}</span>
            <span class="font-label-md text-label-md">{tab.label}</span>
          </a>
        {/each}
      </nav>

      <div class="p-3 border-t border-outline-variant flex flex-col gap-1">
        <a
          href="/config"
          use:inertia
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all
            {active === 'config'
            ? 'bg-secondary-container text-on-secondary-container font-semibold'
            : 'text-on-surface-variant hover:bg-surface-container-high'}"
        >
          <span class="material-symbols-outlined {active === 'config' ? 'fill' : ''}">settings</span>
          <span class="font-label-md text-label-md">Config</span>
        </a>
        {#if showLogout}
          <button
            type="button"
            on:click={logout}
            class="flex items-center gap-3 px-3 py-2.5 rounded-xl text-on-surface-variant
              hover:bg-surface-container-high text-left w-full"
          >
            <span class="material-symbols-outlined">logout</span>
            <span class="font-label-md text-label-md">Sair</span>
          </button>
        {/if}
      </div>
    </aside>
  {/if}

  <!-- Mobile top bar only -->
  <header
    class="fixed top-0 left-0 right-0 z-50 h-16 px-container-margin flex items-center justify-between
      bg-surface-container-lowest border-b border-outline-variant md:hidden"
  >
    <a href="/dashboard" use:inertia class="flex items-center gap-3 min-w-0">
      <div
        class="w-9 h-9 rounded-full bg-primary text-on-primary flex items-center justify-center shrink-0 font-bold text-sm"
      >
        {initials(brand)}
      </div>
      <span class="font-headline-md text-headline-md font-bold text-primary truncate">{brand}</span>
    </a>
    <div class="flex items-center gap-1">
      <a
        href="/config"
        use:inertia
        class="w-10 h-10 flex items-center justify-center rounded-full transition-all
          {active === 'config'
          ? 'bg-secondary-container text-on-secondary-container'
          : 'text-on-surface-variant hover:bg-surface-container-low active:scale-95'}"
        title="Configurações"
        aria-label="Configurações"
      >
        <span class="material-symbols-outlined {active === 'config' ? 'fill' : ''}">settings</span>
      </a>
      {#if showLogout}
        <button
          type="button"
          on:click={logout}
          class="w-10 h-10 flex items-center justify-center rounded-full text-on-surface-variant
            hover:bg-surface-container-low active:scale-95 transition-all"
          title="Sair"
          aria-label="Sair"
        >
          <span class="material-symbols-outlined">logout</span>
        </button>
      {/if}
    </div>
  </header>

  <div class="flex-1 min-w-0 md:pl-56">
    <main
      class="pt-20 md:pt-8 {hideBottomNav ? 'pb-8' : 'pb-28 md:pb-8'} px-container-margin max-w-5xl mx-auto md:mx-0 md:max-w-none md:px-8"
    >
      <slot />
    </main>
  </div>

  <!-- Mobile bottom footer only -->
  {#if !hideBottomNav}
    <nav
      class="fixed bottom-0 left-0 right-0 z-50 h-20 px-2 pb-2 flex justify-around items-center
        bg-surface-container-lowest border-t border-outline-variant shadow-float md:hidden"
    >
      {#each tabs as tab}
        <a
          href={tab.href}
          use:inertia
          class="flex flex-col items-center justify-center min-w-[4rem] px-2 py-1.5 rounded-full transition-all
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
