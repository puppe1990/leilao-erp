<script>
  import { inertia } from '@inertiajs/svelte'
  import PixEletroMark from '@/components/PixEletroMark.svelte'
  export let title = 'Pix Eletro'
  export let site = {}
</script>

<svelte:head>
  <title>{title || 'Pix Eletro'}</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col items-center justify-center px-container-margin text-center">
  <div class="ahq-brand-mark !w-14 !h-14 mx-auto mb-4">
    <PixEletroMark size={48} />
  </div>
  <h1 class="ahq-wordmark text-2xl mb-2">Pix Eletro<span class="text-secondary">.</span></h1>
  <p class="text-on-surface-variant mb-8 max-w-sm">Estoque, lotes e vendas da bancada.</p>
  <div class="flex flex-wrap gap-3 justify-center">
    <a href="/" use:inertia class="ahq-btn-primary px-8">Ver catálogo</a>
    <a href="/login" use:inertia class="ahq-btn-ghost px-8">Entrar</a>
  </div>
</div>
