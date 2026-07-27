<script>
  import { inertia } from '@inertiajs/svelte'
  /** Current section highlight: dashboard | lots | sales | cash | payables | receivables */
  export let active = ''
  export let showLogout = false

  const links = [
    { href: '/dashboard', label: 'Dashboard', key: 'dashboard' },
    { href: '/lots', label: 'Lotes', key: 'lots' },
    { href: '/sales', label: 'Vendas', key: 'sales' },
    { href: '/cash', label: 'Caixa', key: 'cash' },
    { href: '/payables', label: 'A pagar', key: 'payables' },
    { href: '/receivables', label: 'A receber', key: 'receivables' },
  ]
</script>

<nav class="mt-6 text-sm flex flex-wrap items-center gap-x-4 gap-y-2">
  {#each links as link}
    <a
      href={link.href}
      use:inertia
      class="underline {active === link.key ? 'font-semibold text-stone-900' : 'text-stone-700'}"
    >
      {link.label}
    </a>
  {/each}
  {#if showLogout}
    <form method="post" action="/logout" use:inertia class="inline">
      <button type="submit" class="underline text-stone-600">Sair</button>
    </form>
  {/if}
</nav>
