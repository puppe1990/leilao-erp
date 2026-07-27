<script>
  import { inertia } from '@inertiajs/svelte'
  export let title = 'Início'
  export let site = {}
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-6 py-14 text-center">
  <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">
    leilao-erp
  </h1>
  <p class="mt-3 max-w-md text-lg text-stone-600">
    Mini ERP para revenda de leilões — lotes, vendas, caixa e financeiro.
  </p>
  <nav class="mt-8 space-x-4 text-sm">
    <a href="/login" use:inertia class="px-4 py-2 bg-stone-800 text-white rounded">Entrar</a>
    <a href="/contact" use:inertia class="underline text-stone-600">Contato</a>
  </nav>
  <p class="mt-6 text-xs text-stone-500">Inertia: {title}</p>
</div>
