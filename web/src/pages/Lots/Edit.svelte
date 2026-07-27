<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let lot = {}
  export let errors = {}
  export let companyName = 'AuctionHQ'
  export let site = {}

  let form = useForm({
    name: lot.name || '',
    purchased_at: lot.purchasedAt || '',
    auction_source: lot.auctionSource || '',
    notes: lot.notes || '',
  })

  function submit() {
    form.post(`/lots/${lot.id}`)
  }
</script>

<AppShell {companyName} active="lots">
  <div class="mb-section-padding">
    <a href={`/lots/${lot.id}`} use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Voltar
    </a>
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Editar lote</h1>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <form on:submit|preventDefault={submit} class="ahq-card p-5 space-y-4">
    <div>
      <label class="ahq-label block mb-1.5" for="name">Nome</label>
      <input id="name" class="ahq-input" bind:value={form.name} />
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="purchased_at">Data da compra</label>
      <input id="purchased_at" type="date" class="ahq-input font-mono" bind:value={form.purchased_at} />
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="auction_source">Origem / leilão</label>
      <input id="auction_source" class="ahq-input" bind:value={form.auction_source} placeholder="Opcional" />
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="notes">Notas</label>
      <textarea id="notes" class="ahq-input h-24 py-2" bind:value={form.notes} placeholder="Opcional"></textarea>
    </div>
    <div class="flex gap-3">
      <button type="submit" class="ahq-btn-primary flex-1" disabled={form.processing}>Salvar</button>
      <a href={`/lots/${lot.id}`} use:inertia class="ahq-btn-ghost flex-1 text-center">Cancelar</a>
    </div>
  </form>
</AppShell>
