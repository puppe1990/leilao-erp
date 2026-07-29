<script>
  import { inertia } from '@inertiajs/svelte'
  export let title = 'AuctionHQ'
  export let site = {}
</script>

<svelte:head>
  <title>{title || 'AuctionHQ'}</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col items-center justify-center px-container-margin text-center">
  <div class="w-14 h-14 rounded-full bg-primary text-on-primary flex items-center justify-center font-bold text-lg mb-4">
    AQ
  </div>
  <h1 class="font-headline-lg text-headline-lg text-primary mb-2">AuctionHQ</h1>
  <p class="text-on-surface-variant mb-8 max-w-sm">ERP de gestão de leilões — financeiro, lotes e vendas.</p>
  <div class="flex flex-wrap gap-3 justify-center">
    <a href="/" use:inertia class="ahq-btn-primary px-8">Ver catálogo</a>
    <a href="/login" use:inertia class="ahq-btn-ghost px-8">Entrar</a>
  </div>
</div>
