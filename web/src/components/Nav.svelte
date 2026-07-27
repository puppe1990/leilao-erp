<script>
  /** @deprecated Prefer AppShell bottom nav. Kept for thin top secondary links. */
  import { inertia } from '@inertiajs/svelte'
  export let active = ''
  export let showLogout = false

  const links = [
    { href: '/cash', label: 'Caixa', key: 'cash' },
    { href: '/payables', label: 'A pagar', key: 'payables' },
    { href: '/receivables', label: 'A receber', key: 'receivables' },
  ]
</script>

<nav class="flex flex-wrap items-center gap-2 text-sm">
  {#each links as link}
    <a
      href={link.href}
      use:inertia
      class="px-3 py-1.5 rounded-full border transition-colors
        {active === link.key
        ? 'bg-secondary-container text-on-secondary-container border-transparent font-semibold'
        : 'bg-surface-container-lowest text-on-surface-variant border-outline-variant hover:bg-surface-container-low'}"
    >
      {link.label}
    </a>
  {/each}
  {#if showLogout}
    <form method="post" action="/logout" use:inertia class="inline">
      <button type="submit" class="px-3 py-1.5 text-on-surface-variant underline">Sair</button>
    </form>
  {/if}
</nav>
