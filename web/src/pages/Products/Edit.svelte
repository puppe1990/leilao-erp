<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let product = {}
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let name = product.name || ''
  let sale = product.salePriceRaw || ''
  let description = product.description || ''
  let listingText = product.listingText || ''
  let busy = false

  function submit() {
    busy = true
    router.post(
      `/products/${product.id}`,
      {
        name,
        sale_price_hint: sale,
        description,
        listing_text: listingText,
        save_descriptions: '1',
        return_to: `/products/${product.id}`,
      },
      {
        onFinish: () => {
          busy = false
        },
      },
    )
  }
</script>

<AppShell {companyName} active="products">
  <div class="mb-section-padding">
    <a
      href={`/products/${product.id}`}
      use:inertia
      class="text-sm text-secondary flex items-center gap-1 mb-3"
    >
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Voltar ao produto
    </a>
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Editar produto</h1>
    <p class="text-on-surface-variant text-sm mt-1">
      Nome e preço atualizam as unidades em estoque deste catálogo.
    </p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <form on:submit|preventDefault={submit} class="ahq-card p-5 space-y-4 max-w-2xl">
    <div>
      <label class="ahq-label block mb-1.5" for="name">Nome</label>
      <input id="name" class="ahq-input" bind:value={name} required />
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="sale">Preço de venda (R$)</label>
      <input
        id="sale"
        class="ahq-input font-mono"
        bind:value={sale}
        placeholder="0,00"
      />
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="description">Descrição técnica</label>
      <textarea
        id="description"
        class="ahq-input h-28 py-2"
        bind:value={description}
        placeholder="Specs…"
      ></textarea>
    </div>
    <div>
      <label class="ahq-label block mb-1.5" for="listing">Texto anúncio (ML/OLX)</label>
      <textarea
        id="listing"
        class="ahq-input h-32 py-2"
        bind:value={listingText}
        placeholder="Texto para marketplace…"
      ></textarea>
    </div>
    <div class="flex flex-col sm:flex-row gap-3 pt-2">
      <button type="submit" class="ahq-btn-primary flex-1" disabled={busy}>Salvar</button>
      <a
        href={`/products/${product.id}`}
        use:inertia
        class="ahq-btn-ghost flex-1 text-center"
      >
        Cancelar
      </a>
    </div>
  </form>
</AppShell>
