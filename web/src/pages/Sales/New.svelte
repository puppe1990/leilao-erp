<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let errors = {}
  export let items = []
  export let cashAccounts = []
  export let channels = []
  export let site = {}
  export let companyName = 'AuctionHQ'

  $: mains = items.filter((it) => !it.isAccessory)
  $: accessories = items.filter((it) => it.isAccessory)
  $: mainOptions = mains.length > 0 ? mains : items

  let form = useForm({
    item_id: mainOptions[0]?.id?.toString() || items[0]?.id?.toString() || '',
    accessory_ids: [],
    channel: 'direct',
    gross: '',
    fee: '0',
    shipping: '0',
    payment_status: 'received',
    cash_account_id: cashAccounts[0]?.id?.toString() || '',
    due_on: '',
    sold_at: new Date().toISOString().slice(0, 10),
  })

  $: selectedMain = items.find((it) => String(it.id) === String(form.item_id))
  $: selectedAccessories = accessories.filter((it) =>
    form.accessory_ids.map(String).includes(String(it.id)),
  )
  $: totalCostRaw =
    (selectedMain?.unitCostRaw || 0) +
    selectedAccessories.reduce((sum, it) => sum + (it.unitCostRaw || 0), 0)

  let lastPrefillItem = ''
  $: if (selectedMain && String(selectedMain.id) !== lastPrefillItem) {
    lastPrefillItem = String(selectedMain.id)
    if (selectedMain.salePriceRaw) {
      form.gross = formatCentsInput(selectedMain.salePriceRaw)
    }
  }

  function formatCents(cents) {
    const n = Number(cents) || 0
    const neg = n < 0
    const abs = Math.abs(n)
    const s = `${Math.floor(abs / 100)},${String(abs % 100).padStart(2, '0')}`
    return neg ? `-R$ ${s}` : `R$ ${s}`
  }

  function formatCentsInput(cents) {
    const n = Number(cents) || 0
    const abs = Math.abs(n)
    return `${Math.floor(abs / 100)},${String(abs % 100).padStart(2, '0')}`
  }

  function toggleAccessory(id) {
    const sid = String(id)
    const cur = form.accessory_ids.map(String)
    if (cur.includes(sid)) {
      form.accessory_ids = form.accessory_ids.filter((x) => String(x) !== sid)
    } else {
      form.accessory_ids = [...form.accessory_ids, id]
    }
  }

  function isChecked(id) {
    return form.accessory_ids.map(String).includes(String(id))
  }

  /** Pick first free cabo de força + VGA + HDMI (kit completo). */
  function selectCompleteKit() {
    const picked = []
    const usedTitles = new Set()
    const want = [
      (t) => t.includes('força') || t.includes('forca') || t.includes('power'),
      (t) => t.includes('vga'),
      (t) => t.includes('hdmi'),
    ]
    for (const match of want) {
      const found = accessories.find((it) => {
        const t = (it.title || '').toLowerCase()
        return match(t) && !usedTitles.has(it.title) && !picked.includes(it.id)
      })
      if (found) {
        picked.push(found.id)
        usedTitles.add(found.title)
      }
    }
    form.accessory_ids = picked
  }

  function clearAccessories() {
    form.accessory_ids = []
  }

  function submit() {
    form.post('/sales')
  }
</script>

<AppShell {companyName} active="sales">
  <div class="mb-section-padding">
    <a href="/sales" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Vendas
    </a>
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Nova venda</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      Monitor sozinho ou kit completo com cabos do estoque.
    </p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if items.length === 0}
    <div class="ahq-card p-6 border-dashed text-on-surface-variant text-sm">
      <p class="mb-2">Nenhum item em estoque para vender.</p>
      <a href="/lots" use:inertia class="text-secondary font-medium">Ver lotes</a>
    </div>
  {:else}
    <form on:submit|preventDefault={submit} class="ahq-card p-5 space-y-4">
      <div>
        <label class="ahq-label block mb-1.5" for="item_id">Item principal</label>
        <select id="item_id" bind:value={form.item_id} class="ahq-select">
          <option value="">Selecione…</option>
          {#each mainOptions as it}
            <option value={String(it.id)}>
              #{it.id} — {it.title} (custo {it.unitCost}{it.salePriceHint ? ` · venda ${it.salePriceHint}` : ''})
            </option>
          {/each}
        </select>
        {#if selectedMain?.salePriceHint}
          <p class="text-on-surface-variant text-xs mt-1">
            Preço sugerido: <span class="font-mono font-semibold text-secondary">{selectedMain.salePriceHint}</span>
            (preenchido no bruto — pode editar)
          </p>
        {/if}
        {#if errors.item_id}<p class="text-error text-sm mt-1">{errors.item_id}</p>{/if}
      </div>

      {#if accessories.length > 0}
        <fieldset class="border border-outline-variant rounded-lg p-4 space-y-3">
          <legend class="ahq-label px-1">Acessórios (cabos)</legend>
          <div class="flex flex-wrap gap-2 mb-2">
            <button
              type="button"
              class="ahq-btn-primary h-9 px-3 text-xs"
              on:click={selectCompleteKit}
            >
              <span class="material-symbols-outlined text-[16px] mr-1">package_2</span>
              Kit completo
            </button>
            <button type="button" class="ahq-btn-ghost h-9 px-3 text-xs" on:click={clearAccessories}>
              Só o item
            </button>
          </div>
          <p class="text-on-surface-variant text-xs mb-2">
            Kit completo = 1 cabo de força + 1 VGA + 1 HDMI (se houver em estoque).
          </p>
          <div class="flex flex-col gap-2 max-h-48 overflow-y-auto">
            {#each accessories as acc}
              <label class="flex items-center gap-2 text-body-md cursor-pointer">
                <input
                  type="checkbox"
                  checked={isChecked(acc.id)}
                  on:change={() => toggleAccessory(acc.id)}
                  class="text-secondary rounded"
                />
                <span class="flex-1">#{acc.id} — {acc.title}</span>
                <span class="font-mono text-sm text-on-surface-variant">{acc.unitCost}</span>
              </label>
            {/each}
          </div>
          {#if errors.accessory_ids}
            <p class="text-error text-sm mt-1">{errors.accessory_ids}</p>
          {/if}
        </fieldset>
      {/if}

      <div class="bg-surface-container-low rounded-lg p-3 text-sm">
        <span class="ahq-label text-[10px]">Custo total da composição</span>
        <p class="font-mono font-semibold text-primary">{formatCents(totalCostRaw)}</p>
        {#if selectedAccessories.length > 0}
          <p class="text-on-surface-variant text-xs mt-1">
            {selectedMain?.title || 'Item'} + {selectedAccessories.length} acessório(s)
          </p>
        {/if}
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="ahq-label block mb-1.5" for="channel">Canal</label>
          <select id="channel" bind:value={form.channel} class="ahq-select">
            {#each channels as ch}
              <option value={ch.value}>{ch.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="sold_at">Data</label>
          <input id="sold_at" type="date" bind:value={form.sold_at} class="ahq-input font-mono" />
        </div>
      </div>

      <div>
        <label class="ahq-label block mb-1.5" for="gross">Valor bruto (R$)</label>
        <input
          id="gross"
          type="text"
          bind:value={form.gross}
          class="ahq-input font-mono"
          placeholder="150,00"
        />
        {#if errors.gross}<p class="text-error text-sm mt-1">{errors.gross}</p>{/if}
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="ahq-label block mb-1.5" for="fee">Taxa (R$)</label>
          <input id="fee" type="text" bind:value={form.fee} class="ahq-input font-mono" placeholder="0,00" />
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="shipping">Frete (R$)</label>
          <input
            id="shipping"
            type="text"
            bind:value={form.shipping}
            class="ahq-input font-mono"
            placeholder="0,00"
          />
        </div>
      </div>

      <fieldset class="border border-outline-variant rounded-lg p-4 space-y-3">
        <legend class="ahq-label px-1">Pagamento</legend>
        <label class="flex items-center gap-2 text-body-md">
          <input type="radio" bind:group={form.payment_status} value="received" class="text-secondary" />
          Recebi agora
        </label>
        <label class="flex items-center gap-2 text-body-md">
          <input type="radio" bind:group={form.payment_status} value="pending" class="text-secondary" />
          A receber
        </label>

        {#if form.payment_status === 'received'}
          <div>
            <label class="ahq-label block mb-1.5" for="cash_account_id">Conta de caixa</label>
            <select id="cash_account_id" bind:value={form.cash_account_id} class="ahq-select">
              <option value="">Selecione…</option>
              {#each cashAccounts as acc}
                <option value={String(acc.id)}>{acc.name}</option>
              {/each}
            </select>
            {#if errors.cash_account_id}
              <p class="text-error text-sm mt-1">{errors.cash_account_id}</p>
            {/if}
          </div>
        {/if}

        {#if form.payment_status === 'pending'}
          <div>
            <label class="ahq-label block mb-1.5" for="due_on">Vencimento</label>
            <input id="due_on" type="date" bind:value={form.due_on} class="ahq-input font-mono" />
            {#if errors.due_on}<p class="text-error text-sm mt-1">{errors.due_on}</p>{/if}
          </div>
        {/if}
      </fieldset>

      <div class="flex flex-col sm:flex-row gap-3 pt-2">
        <button type="submit" class="ahq-btn-primary flex-1" disabled={form.processing}>Salvar venda</button>
        <a href="/sales" use:inertia class="ahq-btn-ghost flex-1 text-center">Cancelar</a>
      </div>
    </form>
  {/if}
</AppShell>
