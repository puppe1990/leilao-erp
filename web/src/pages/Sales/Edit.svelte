<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'

  export let sale = {}
  export let channels = []
  export let errors = {}
  export let companyName = 'AuctionHQ'
  export let site = {}

  let form = useForm({
    sold_at: sale.soldAt || '',
    channel: sale.channel || 'direct',
    gross: sale.gross || '',
    fee: sale.fee || '0',
    shipping: sale.shipping || '0',
    due_on: sale.dueOn || '',
  })

  $: channelOptions = (Array.isArray(channels) ? channels : []).map((ch) => ({
    value: String(ch.value),
    label: ch.label,
  }))

  function submit() {
    form.post(`/sales/${sale.id}`)
  }
</script>

<AppShell {companyName} active="sales">
  <div class="mb-section-padding">
    <a href={`/sales/${sale.id}`} use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Voltar
    </a>
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Editar venda</h1>
    <p class="text-on-surface-variant text-sm mt-1">{sale.itemTitle}</p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <form on:submit|preventDefault={submit} class="ahq-card p-5 space-y-4">
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="ahq-label block mb-1.5" for="sold_at">Data</label>
        <input id="sold_at" type="date" class="ahq-input font-mono" bind:value={form.sold_at} />
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="channel">Canal</label>
        <SearchableSelect
          id="channel"
          bind:value={form.channel}
          options={channelOptions}
          placeholder="Canal"
          searchPlaceholder="Buscar canal…"
          allowClear={false}
        />
      </div>
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="gross">Bruto (R$)</label>
      <input id="gross" class="ahq-input font-mono" bind:value={form.gross} />
      {#if errors.gross}<p class="text-error text-sm mt-1">{errors.gross}</p>{/if}
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="ahq-label block mb-1.5" for="fee">Taxa (R$)</label>
        <input id="fee" class="ahq-input font-mono" bind:value={form.fee} />
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="shipping">Frete (R$)</label>
        <input id="shipping" class="ahq-input font-mono" bind:value={form.shipping} />
      </div>
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="due_on">Vencimento (a receber)</label>
      <input id="due_on" type="date" class="ahq-input font-mono" bind:value={form.due_on} />
    </div>
    <div class="flex gap-3">
      <button type="submit" class="ahq-btn-primary flex-1" disabled={form.processing}>Salvar</button>
      <a href={`/sales/${sale.id}`} use:inertia class="ahq-btn-ghost flex-1 text-center">Cancelar</a>
    </div>
  </form>
</AppShell>
