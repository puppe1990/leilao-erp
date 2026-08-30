<script>
  import { inertia, router, useForm } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let sale = {}
  export let errors = {}
  export let companyName = 'AuctionHQ'
  export let site = {}
  export let clients = []
  export let companyCnpj = ''
  export let receiptSellerName = ''

  const clientList = Array.isArray(clients) ? clients : []

  let attachForm = useForm({
    client_id: sale.clientId ? String(sale.clientId) : '',
  })

  $: clientOptions = clientList.map((c) => ({
    value: String(c.id),
    label: c.document ? `${c.name} · ${c.document}` : c.name,
  }))

  async function destroy() {
    const ok = await askConfirm({
      title: 'Excluir venda',
      message: 'Excluir/cancelar esta venda pendente?',
      detail: 'Os itens voltam ao estoque. Essa ação não pode ser desfeita.',
      confirmLabel: 'Excluir venda',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/sales/${sale.id}/delete`)
  }
</script>

<AppShell {companyName} active="sales">
  <div class="mb-section-padding">
    <a href="/sales" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Vendas
    </a>
    <div class="flex items-start justify-between gap-3">
      <div>
        <h1 class="font-headline-lg text-headline-lg-mobile text-primary">{sale.itemTitle || `Venda #${sale.id}`}</h1>
        <p class="text-on-surface-variant text-body-md mt-1">
          {sale.soldAt?.slice?.(0, 10) || sale.soldAt} · {sale.channelLabel}
        </p>
      </div>
      <span class="ahq-badge-pending">{sale.paymentLabel}</span>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <div class="ahq-card p-5 mb-section-padding grid grid-cols-2 gap-4">
    <div>
      <span class="ahq-label">Bruto</span>
      <p class="ahq-value">{sale.gross}</p>
    </div>
    <div>
      <span class="ahq-label">Taxa</span>
      <p class="ahq-value">{sale.fee}</p>
    </div>
    <div>
      <span class="ahq-label">Frete</span>
      <p class="ahq-value">{sale.shipping}</p>
    </div>
    <div>
      <span class="ahq-label">Líquido</span>
      <p class="ahq-value text-secondary">{sale.net}</p>
    </div>
    <div>
      <span class="ahq-label">Custo total (composição)</span>
      <p class="font-mono font-semibold">{sale.unitCost}</p>
    </div>
    <div>
      <span class="ahq-label">Margem</span>
      <p class="font-mono font-semibold text-secondary">{sale.margin || '—'}</p>
    </div>
  </div>

  {#if sale.lines?.length}
    <div class="ahq-card p-5 mb-section-padding">
      <h2 class="font-semibold text-primary mb-3">Itens da venda</h2>
      <ul class="divide-y divide-outline-variant">
        {#each sale.lines as line}
          <li class="py-2 flex justify-between gap-3 text-sm">
            <div>
              <span class="font-medium">{line.title}</span>
              <span class="text-on-surface-variant ml-2 text-xs">{line.roleLabel}</span>
            </div>
            <span class="font-mono shrink-0">{line.unitCost}</span>
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  <section class="ahq-card p-5 mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-1">Cliente</h2>
    {#if sale.clientId}
      <p class="font-semibold text-primary">{sale.clientName}</p>
      {#if sale.clientDocument}
        <p class="font-mono text-sm text-on-surface-variant mt-1">CPF/CNPJ {sale.clientDocument}</p>
      {/if}
      {#if sale.clientPhone}
        <p class="text-sm text-on-surface-variant mt-0.5">Telefone {sale.clientPhone}</p>
      {/if}
      {#if sale.clientEmail}
        <p class="text-sm text-on-surface-variant mt-0.5">{sale.clientEmail}</p>
      {/if}
    {:else if sale.canAttachClient}
      <p class="text-on-surface-variant text-sm mb-3">
        Esta venda ainda não tem comprador. Vincule um cliente cadastrado.
      </p>
      <form
        class="space-y-3"
        on:submit|preventDefault={() => attachForm.post(`/sales/${sale.id}/client`)}
      >
        <SearchableSelect
          id="attach-client"
          options={clientOptions}
          bind:value={attachForm.client_id}
          placeholder="Selecionar cliente…"
          emptyLabel="Nenhum cliente"
        />
        {#if errors.client_id}
          <p class="text-error text-sm">{errors.client_id}</p>
        {/if}
        <p class="text-[11px] text-on-surface-variant">
          <a href="/clients" use:inertia class="text-secondary underline">Cadastrar cliente</a>
        </p>
        <button type="submit" class="ahq-btn-primary" disabled={!attachForm.client_id}>
          Salvar cliente na venda
        </button>
      </form>
    {:else}
      <p class="text-on-surface-variant text-sm">Sem cliente nesta venda.</p>
    {/if}
  </section>

  <section class="ahq-card p-5 mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-1">Recibo de venda</h2>
    <p class="text-on-surface-variant text-sm mb-4">
      PDF com CNPJ da empresa e dados do cliente desta venda. Não é documento fiscal.
    </p>

    {#if sale.paymentStatus === 'cancelled'}
      <p class="text-error text-sm">Venda cancelada não emite recibo.</p>
    {:else if !sale.clientId}
      <p class="text-on-surface-variant text-sm">Vincule um cliente acima para emitir o recibo.</p>
    {:else if !companyCnpj}
      <p class="text-on-surface-variant text-sm">
        Salve o CNPJ em
        <a href="/config" use:inertia class="text-secondary underline">Configurações</a>
        para emitir o recibo.
      </p>
    {:else}
      <a href={`/sales/${sale.id}/recibo.pdf`} class="ahq-btn-primary" target="_blank" rel="noopener">
        Baixar PDF
      </a>
    {/if}
  </section>

  <div class="flex flex-wrap gap-3">
    {#if sale.canEdit}
      <a href={`/sales/${sale.id}/edit`} use:inertia class="ahq-btn-primary">Editar</a>
    {/if}
    {#if sale.canDelete}
      <button type="button" class="ahq-btn-ghost text-error border-error" on:click={destroy}>Excluir</button>
    {/if}
    <a href="/sales" use:inertia class="ahq-btn-ghost">Voltar</a>
  </div>
</AppShell>
