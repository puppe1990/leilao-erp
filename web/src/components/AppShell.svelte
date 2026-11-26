<script>
  import { inertia, router } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import {
    applyShopThemeToDocument,
    getShopTheme,
    setShopTheme,
  } from '@/lib/shopTheme.js'
  import ConfirmModal from '@/components/ConfirmModal.svelte'
  import PixEletroMark from '@/components/PixEletroMark.svelte'

  /** @type {'dashboard'|'lots'|'stock'|'products'|'finds'|'suppliers'|'sales'|'clients'|'cash'|'payables'|'receivables'|'config'|''} */
  export let active = ''
  export let title = ''
  export let companyName = 'Pix Eletro'
  export let showLogout = true
  export let hideBottomNav = false

  $: brand = title || companyName || 'Pix Eletro'

  /** @type {'dark'|'light'} */
  let theme = 'dark'

  onMount(() => {
    theme = getShopTheme()
    applyShopThemeToDocument(theme)
  })

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    setShopTheme(theme)
  }

  function logout() {
    router.post('/logout')
  }

  const tabs = [
    { href: '/dashboard', key: 'dashboard', label: 'Dashboard', icon: 'dashboard' },
    { href: '/lots', key: 'lots', label: 'Lotes', icon: 'gavel' },
    { href: '/stock', key: 'stock', label: 'Estoque', icon: 'inventory_2' },
    { href: '/products', key: 'products', label: 'Produtos', icon: 'category' },
    { href: '/finds', key: 'finds', label: 'Achados', icon: 'travel_explore' },
    { href: '/suppliers', key: 'suppliers', label: 'Fornecedores', icon: 'local_shipping' },
    { href: '/sales', key: 'sales', label: 'Vendas', icon: 'sell' },
    { href: '/clients', key: 'clients', label: 'Clientes', icon: 'group' },
    { href: '/cash', key: 'cash', label: 'Finanças', icon: 'analytics' },
  ]

  function isActive(key) {
    if (active === key) return true
    if (key === 'cash' && (active === 'payables' || active === 'receivables')) return true
    return false
  }

  const pageTitles = {
    dashboard: 'Dashboard',
    lots: 'Lotes',
    stock: 'Estoque',
    products: 'Produtos',
    finds: 'Achados',
    suppliers: 'Fornecedores',
    sales: 'Vendas',
    clients: 'Clientes',
    cash: 'Caixa',
    payables: 'A pagar',
    receivables: 'A receber',
    config: 'Configurações',
  }

  $: docTitle = pageTitles[active] ? `${pageTitles[active]} · ${brand}` : brand
</script>

<svelte:head>
  <title>{docTitle}</title>
</svelte:head>

<div class="min-h-screen bg-background font-body-md text-body-md text-on-surface md:flex">
  <!-- Desktop sidebar -->
  {#if !hideBottomNav}
    <aside
      class="hidden md:flex md:flex-col md:w-60 md:shrink-0 md:fixed md:inset-y-0 md:left-0 z-50
        bg-surface-container-lowest border-r border-outline-variant"
    >
      <a
        href="/dashboard"
        use:inertia
        class="h-16 px-4 flex items-center gap-3 border-b border-outline-variant min-w-0"
      >
        <div class="ahq-brand-mark">
          <PixEletroMark size={28} />
        </div>
        <div class="min-w-0">
          <div class="flex items-center gap-2 min-w-0">
            <span class="ahq-wordmark truncate">{brand}<span class="text-secondary">.</span></span>
          </div>
          <p class="text-[11px] text-on-surface-variant truncate">Admin</p>
        </div>
      </a>

      <nav class="flex-1 p-3 flex flex-col gap-1 overflow-y-auto" aria-label="Navegação principal">
        {#each tabs as tab}
          <a
            href={tab.href}
            use:inertia
            class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors
              {isActive(tab.key)
              ? 'bg-surface-container-high text-on-surface font-medium'
              : 'text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface'}"
          >
            <span class="material-symbols-outlined {isActive(tab.key) ? 'fill' : ''}">{tab.icon}</span>
            <span class="text-sm font-medium">{tab.label}</span>
          </a>
        {/each}
      </nav>

      <div class="p-3 border-t border-outline-variant flex flex-col gap-1">
        <a
          href="/"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant
            hover:bg-surface-container-high hover:text-secondary transition-colors"
          title="Abrir catálogo público"
        >
          <span class="material-symbols-outlined">storefront</span>
          <span class="text-sm font-medium">Catálogo</span>
        </a>
        <a
          href="/config"
          use:inertia
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors
            {active === 'config'
            ? 'bg-surface-container-high text-on-surface font-medium'
            : 'text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface'}"
        >
          <span class="material-symbols-outlined {active === 'config' ? 'fill' : ''}">settings</span>
          <span class="text-sm font-medium">Config</span>
        </a>
        <button
          type="button"
          on:click={toggleTheme}
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant
            hover:bg-surface-container-high hover:text-on-surface text-left w-full transition-colors"
          title={theme === 'dark' ? 'Modo claro' : 'Modo escuro'}
        >
          <span class="material-symbols-outlined"
            >{theme === 'dark' ? 'light_mode' : 'dark_mode'}</span
          >
          <span class="text-sm font-medium">{theme === 'dark' ? 'Modo claro' : 'Modo escuro'}</span>
        </button>
        {#if showLogout}
          <button
            type="button"
            on:click={logout}
            class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-on-surface-variant
              hover:bg-surface-container-high hover:text-error text-left w-full transition-colors"
          >
            <span class="material-symbols-outlined">logout</span>
            <span class="text-sm font-medium">Sair</span>
          </button>
        {/if}
      </div>
    </aside>
  {/if}

  <!-- Mobile top bar -->
  <header
    class="fixed top-0 left-0 right-0 z-50 h-16 px-container-margin flex items-center justify-between
      bg-surface-container-lowest/95 backdrop-blur-md border-b border-outline-variant md:hidden"
  >
    <a href="/dashboard" use:inertia class="flex items-center gap-2.5 min-w-0">
      <div class="ahq-brand-mark !w-8 !h-8">
        <PixEletroMark size={24} />
      </div>
      <span class="ahq-wordmark truncate">{brand}<span class="text-secondary">.</span></span>
    </a>
    <div class="flex items-center gap-1">
      <button
        type="button"
        class="ahq-theme-btn !w-10 !h-10"
        on:click={toggleTheme}
        title={theme === 'dark' ? 'Modo claro' : 'Modo escuro'}
        aria-label={theme === 'dark' ? 'Ativar modo claro' : 'Ativar modo escuro'}
      >
        <span class="material-symbols-outlined text-[20px]"
          >{theme === 'dark' ? 'light_mode' : 'dark_mode'}</span
        >
      </button>
      <a
        href="/config"
        use:inertia
        class="w-10 h-10 flex items-center justify-center rounded-lg transition-colors
          {active === 'config'
          ? 'bg-surface-container-high text-on-surface'
          : 'text-on-surface-variant hover:bg-surface-container-high'}"
        title="Configurações"
        aria-label="Configurações"
      >
        <span class="material-symbols-outlined {active === 'config' ? 'fill' : ''}">settings</span>
      </a>
      {#if showLogout}
        <button
          type="button"
          on:click={logout}
          class="w-10 h-10 flex items-center justify-center rounded-lg text-on-surface-variant
            hover:bg-surface-container-high transition-colors"
          title="Sair"
          aria-label="Sair"
        >
          <span class="material-symbols-outlined">logout</span>
        </button>
      {/if}
    </div>
  </header>

  <div class="flex-1 min-w-0 md:pl-60">
    <main
      class="pt-20 md:pt-8 {hideBottomNav
        ? 'pb-8'
        : 'pb-28 md:pb-8'} px-container-margin max-w-5xl mx-auto md:mx-0 md:max-w-none md:px-8"
    >
      <slot />
    </main>
  </div>

  <!-- Mobile bottom nav -->
  {#if !hideBottomNav}
    <nav
      class="fixed bottom-0 left-0 right-0 z-50 h-20 px-1 pb-2 flex justify-start items-center overflow-x-auto
        bg-surface-container-lowest/95 backdrop-blur-md border-t border-outline-variant shadow-float md:hidden"
    >
      {#each tabs as tab}
        <a
          href={tab.href}
          use:inertia
          class="flex flex-col items-center justify-center min-w-[3.5rem] shrink-0 px-1.5 py-1.5 rounded-lg transition-colors
            {isActive(tab.key)
            ? 'bg-surface-container-high text-on-surface'
            : 'text-on-surface-variant hover:bg-surface-container-high'}"
        >
          <span class="material-symbols-outlined {isActive(tab.key) ? 'fill' : ''}">{tab.icon}</span>
          <span class="text-[10px] font-medium mt-0.5">{tab.label}</span>
        </a>
      {/each}
    </nav>
  {/if}
</div>

<ConfirmModal />
