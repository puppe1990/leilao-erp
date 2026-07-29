<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'

  export let product = {}
  export let olxOptions = {}
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let name = product.name || ''
  let sale = product.salePriceRaw || ''
  let description = product.description || ''
  let listingText = product.listingText || ''
  let screenType = product.screenType || ''
  let maxResolution = product.maxResolution || ''
  let refreshRate = product.refreshRate || ''
  let condition = product.condition || ''
  let olxFreeShipping = !!product.olxFreeShipping
  let shopVisible = !!product.shopVisible
  let features = {
    curved: false,
    includesBox: false,
    displayPort: false,
    hdr: false,
    widescreen: false,
    includesCables: false,
    audio: false,
    hdmi: false,
    ultrawide: false,
    ...(product.features || {}),
  }
  let busy = false

  $: screenTypeOptions = (olxOptions.screenTypes || []).map((o) => ({
    value: o.value,
    label: o.label,
  }))
  $: resolutionOptions = (olxOptions.resolutions || []).map((o) => ({
    value: o.value,
    label: o.label,
  }))
  $: refreshOptions = (olxOptions.refreshRates || []).map((o) => ({
    value: o.value,
    label: o.label,
  }))
  $: conditionOptions = (olxOptions.conditions || []).map((o) => ({
    value: o.value,
    label: o.label,
  }))
  $: featureDefs = olxOptions.features || [
    { key: 'curved', label: 'Curvo' },
    { key: 'includesBox', label: 'Inclui caixa' },
    { key: 'displayPort', label: 'Possui DisplayPort' },
    { key: 'hdr', label: 'Possui HDR' },
    { key: 'widescreen', label: 'Widescreen' },
    { key: 'includesCables', label: 'Inclui cabos' },
    { key: 'audio', label: 'Possui áudio' },
    { key: 'hdmi', label: 'Possui HDMI' },
    { key: 'ultrawide', label: 'Ultrawide' },
  ]

  const selectBtn =
    'ahq-select h-10 w-full text-left flex items-center justify-between gap-2'

  function submit() {
    busy = true
    const payload = {
      name,
      sale_price_hint: sale,
      description,
      listing_text: listingText,
      save_descriptions: '1',
      save_olx: '1',
      screen_type: screenType,
      max_resolution: maxResolution,
      refresh_rate: refreshRate,
      condition,
      olx_free_shipping: olxFreeShipping ? '1' : '0',
      save_shop_visible: '1',
      shop_visible: shopVisible ? '1' : '0',
      return_to: `/products/${product.id}`,
    }
    for (const f of featureDefs) {
      payload[`feat_${f.key}`] = features[f.key] ? '1' : '0'
    }
    router.post(`/products/${product.id}`, payload, {
      onFinish: () => {
        busy = false
      },
    })
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
      Nome, preço e atributos OLX para agilizar o anúncio.
    </p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <form on:submit|preventDefault={submit} class="space-y-4 max-w-2xl">
    <section class="ahq-card p-5 space-y-4">
      <h2 class="font-semibold text-primary">Cadastro</h2>
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
    </section>

    <section class="ahq-card p-5 space-y-4">
      <div>
        <h2 class="font-semibold text-primary">Atributos OLX</h2>
        <p class="text-xs text-on-surface-variant mt-0.5">
          Mesmas opções do formulário de monitores na OLX.
        </p>
      </div>

      <div>
        <label class="ahq-label block mb-1.5" for="screen_type">Tipo de tela</label>
        <SearchableSelect
          id="screen_type"
          bind:value={screenType}
          options={screenTypeOptions}
          placeholder="Selecione"
          searchPlaceholder="Buscar tipo…"
          allowClear={true}
          buttonClass={selectBtn}
        />
      </div>

      <div>
        <label class="ahq-label block mb-1.5" for="max_resolution">Resolução máxima</label>
        <SearchableSelect
          id="max_resolution"
          bind:value={maxResolution}
          options={resolutionOptions}
          placeholder="Selecione"
          searchPlaceholder="Buscar resolução…"
          allowClear={true}
          buttonClass={selectBtn}
        />
      </div>

      <div>
        <label class="ahq-label block mb-1.5" for="refresh_rate">Taxa de atualização</label>
        <SearchableSelect
          id="refresh_rate"
          bind:value={refreshRate}
          options={refreshOptions}
          placeholder="Selecione"
          searchPlaceholder="Buscar Hz…"
          allowClear={true}
          buttonClass={selectBtn}
        />
      </div>

      <div>
        <label class="ahq-label block mb-1.5" for="condition">
          Condição <span class="text-error">*</span>
        </label>
        <SearchableSelect
          id="condition"
          bind:value={condition}
          options={conditionOptions}
          placeholder="Selecione"
          searchPlaceholder="Buscar condição…"
          allowClear={true}
          buttonClass={selectBtn}
        />
      </div>

      <div>
        <p class="ahq-label mb-2">Características</p>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          {#each featureDefs as f (f.key)}
            <label
              class="flex items-center gap-2 text-sm text-on-surface cursor-pointer rounded-lg border border-outline-variant px-3 py-2 hover:bg-surface-container-low"
            >
              <input
                type="checkbox"
                class="rounded border-outline-variant text-secondary focus:ring-secondary"
                bind:checked={features[f.key]}
              />
              {f.label}
            </label>
          {/each}
        </div>
      </div>

      <div class="rounded-xl border-2 border-outline-variant p-4 space-y-2">
        <p class="font-semibold text-primary text-sm">Visível no catálogo (ecommerce)</p>
        <p class="text-xs text-on-surface-variant">
          Além desta flag, o produto só aparece se tiver foto e estoque.
        </p>
        <div class="flex flex-wrap gap-2 pt-1">
          <button
            type="button"
            class="flex-1 min-w-[8rem] rounded-lg border px-3 py-2.5 text-sm
              {shopVisible
              ? 'border-secondary bg-secondary-container/40 text-on-secondary-container font-semibold'
              : 'border-outline-variant'}"
            on:click={() => (shopVisible = true)}
          >
            Sim — mostrar
          </button>
          <button
            type="button"
            class="flex-1 min-w-[8rem] rounded-lg border px-3 py-2.5 text-sm
              {!shopVisible
              ? 'border-outline-variant bg-surface-container-high font-semibold'
              : 'border-outline-variant'}"
            on:click={() => (shopVisible = false)}
          >
            Não — ocultar
          </button>
        </div>
      </div>

      <div class="rounded-xl border-2 border-outline-variant p-4 space-y-2">
        <p class="font-semibold text-primary text-sm">Entregar grátis pela OLX</p>
        <p class="text-xs text-on-surface-variant">
          Flag de decisão: Sim se for oferecer frete grátis nesse anúncio (só paga se vender).
        </p>
        <div class="flex flex-wrap gap-2 pt-1">
          <button
            type="button"
            class="flex-1 min-w-[8rem] rounded-lg border px-3 py-2.5 text-sm
              {olxFreeShipping
              ? 'border-secondary bg-secondary-container/40 text-on-secondary-container font-semibold'
              : 'border-outline-variant'}"
            on:click={() => (olxFreeShipping = true)}
          >
            Sim
          </button>
          <button
            type="button"
            class="flex-1 min-w-[8rem] rounded-lg border px-3 py-2.5 text-sm
              {!olxFreeShipping
              ? 'border-outline-variant bg-surface-container-high font-semibold'
              : 'border-outline-variant'}"
            on:click={() => (olxFreeShipping = false)}
          >
            Não
          </button>
        </div>
      </div>
    </section>

    <div class="flex flex-col sm:flex-row gap-3">
      <button type="submit" class="ahq-btn-primary flex-1 h-11" disabled={busy}>Salvar</button>
      <a
        href={`/products/${product.id}`}
        use:inertia
        class="ahq-btn-ghost flex-1 text-center h-11 inline-flex items-center justify-center"
      >
        Cancelar
      </a>
    </div>
  </form>
</AppShell>
