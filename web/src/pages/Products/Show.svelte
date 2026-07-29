<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let product = {}
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let videoURL = ''
  let mediaBusy = false

  function isVideo(m) {
    return m?.kind === 'video'
  }

  async function copyText(text) {
    const t = String(text || '').trim()
    if (!t) return
    try {
      await navigator.clipboard.writeText(t)
    } catch {
      // ignore
    }
  }

  function addVideo() {
    const url = videoURL.trim()
    if (!url) return
    mediaBusy = true
    router.post(
      `/products/${product.id}/media`,
      { kind: 'video', url },
      {
        forceFormData: true,
        onFinish: () => {
          mediaBusy = false
        },
        onSuccess: () => {
          videoURL = ''
        },
      },
    )
  }

  function addPhotoURL(url) {
    const u = String(url || '').trim()
    if (!u) return
    mediaBusy = true
    router.post(
      `/products/${product.id}/media`,
      { kind: 'photo', url: u },
      {
        forceFormData: true,
        onFinish: () => {
          mediaBusy = false
        },
      },
    )
  }

  function onPhotoFile(e) {
    const file = e?.target?.files?.[0]
    if (!file) return
    mediaBusy = true
    router.post(
      `/products/${product.id}/media`,
      { kind: 'photo', file },
      {
        forceFormData: true,
        onFinish: () => {
          mediaBusy = false
          if (e?.target) e.target.value = ''
        },
      },
    )
  }

  function deleteMedia(m) {
    if (!confirm('Remover esta mídia?')) return
    mediaBusy = true
    router.post(`/products/${product.id}/media/${m.id}/delete`, {}, {
      onFinish: () => {
        mediaBusy = false
      },
    })
  }
</script>

<AppShell {companyName} active="products">
  <div class="mb-section-padding">
    <a href="/products" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Produtos
    </a>
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 class="font-headline-lg text-headline-lg-mobile text-primary">{product.name}</h1>
        <p class="text-on-surface-variant text-sm mt-1">
          {product.kindLabel || 'Principal'}
          · estoque {product.qtyInStock ?? 0}
          · {product.salePriceHint || 'sem preço'}
        </p>
      </div>
      <a
        href={`/products/${product.id}/edit`}
        use:inertia
        class="ahq-btn-primary h-10 px-4 text-sm inline-flex items-center"
      >
        <span class="material-symbols-outlined text-[18px] mr-1">edit</span>
        Editar
      </a>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {@const feats = product.features || {}}
  {@const featureLabels = [
    { key: 'curved', label: 'Curvo' },
    { key: 'includesBox', label: 'Inclui caixa' },
    { key: 'displayPort', label: 'Possui DisplayPort' },
    { key: 'hdr', label: 'Possui HDR' },
    { key: 'widescreen', label: 'Widescreen' },
    { key: 'includesCables', label: 'Inclui cabos' },
    { key: 'audio', label: 'Possui áudio' },
    { key: 'hdmi', label: 'Possui HDMI' },
    { key: 'ultrawide', label: 'Ultrawide' },
  ]}
  {@const activeFeatures = featureLabels.filter((f) => feats[f.key])}
  {@const hasOlx =
    product.screenType ||
    product.maxResolution ||
    product.refreshRate ||
    product.condition ||
    activeFeatures.length > 0}

  <section class="ahq-card p-4 mb-4 space-y-3">
    <div class="flex items-center justify-between gap-2 flex-wrap">
      <h2 class="font-semibold text-primary">Atributos OLX</h2>
      <a
        href={`/products/${product.id}/edit`}
        use:inertia
        class="text-xs text-secondary font-medium"
      >
        Editar atributos
      </a>
    </div>
    {#if !hasOlx}
      <p class="text-sm text-on-surface-variant">
        Ainda sem atributos. Preencha na edição para copiar no anúncio da OLX.
      </p>
    {:else}
      <dl class="grid sm:grid-cols-2 gap-x-4 gap-y-2 text-sm">
        <div>
          <dt class="text-on-surface-variant text-xs">Tipo de tela</dt>
          <dd class="font-medium text-primary">{product.screenType || '—'}</dd>
        </div>
        <div>
          <dt class="text-on-surface-variant text-xs">Resolução máxima</dt>
          <dd class="font-medium text-primary">{product.maxResolution || '—'}</dd>
        </div>
        <div>
          <dt class="text-on-surface-variant text-xs">Taxa de atualização</dt>
          <dd class="font-medium text-primary">{product.refreshRate || '—'}</dd>
        </div>
        <div>
          <dt class="text-on-surface-variant text-xs">Condição</dt>
          <dd class="font-medium text-primary">{product.condition || '—'}</dd>
        </div>
        <div class="sm:col-span-2">
          <dt class="text-on-surface-variant text-xs">Entregar grátis pela OLX</dt>
          <dd class="mt-0.5">
            {#if product.olxFreeShipping}
              <span
                class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-secondary-container text-on-secondary-container"
              >
                <span class="material-symbols-outlined text-[16px]">local_shipping</span>
                Sim — oferecer frete grátis
              </span>
            {:else}
              <span
                class="inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-full bg-surface-container-high text-on-surface-variant"
              >
                Não — só retirada / frete do comprador
              </span>
            {/if}
          </dd>
        </div>
      </dl>
      {#if activeFeatures.length}
        <div>
          <p class="text-on-surface-variant text-xs mb-1.5">Características</p>
          <ul class="flex flex-wrap gap-1.5">
            {#each activeFeatures as f (f.key)}
              <li
                class="text-xs px-2 py-1 rounded-full bg-secondary-container text-on-secondary-container"
              >
                {f.label}
              </li>
            {/each}
          </ul>
        </div>
      {/if}
    {/if}
  </section>

  <div class="grid gap-4 lg:grid-cols-2">
    <section class="ahq-card p-4 space-y-3">
      <div class="flex items-center justify-between gap-2">
        <h2 class="font-semibold text-primary">Descrição técnica</h2>
        {#if product.description}
          <button
            type="button"
            class="text-xs text-secondary font-medium"
            on:click={() => copyText(product.description)}
          >
            Copiar
          </button>
        {/if}
      </div>
      <p class="text-sm text-on-surface whitespace-pre-wrap">
        {product.description || 'Sem ficha técnica.'}
      </p>
    </section>

    <section class="ahq-card p-4 space-y-3">
      <div class="flex items-center justify-between gap-2">
        <h2 class="font-semibold text-primary">Anúncio (ML/OLX)</h2>
        {#if product.listingText}
          <button
            type="button"
            class="text-xs text-secondary font-medium"
            on:click={() => copyText(product.listingText)}
          >
            Copiar
          </button>
        {/if}
      </div>
      <p class="text-sm text-on-surface whitespace-pre-wrap">
        {product.listingText || 'Sem texto de anúncio.'}
      </p>
    </section>
  </div>

  <section class="ahq-card p-4 mt-4 space-y-4">
    <h2 class="font-semibold text-primary">Fotos e vídeos</h2>

    {#if !(product.media || []).length}
      <p class="text-sm text-on-surface-variant">Nenhuma mídia ainda.</p>
    {:else}
      <ul class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {#each product.media as m (m.id)}
          <li class="flex gap-3 items-start border border-outline-variant rounded-lg p-3">
            <div
              class="w-24 h-20 shrink-0 rounded bg-surface-container flex items-center justify-center overflow-hidden"
            >
              {#if isVideo(m)}
                <span class="material-symbols-outlined text-3xl text-secondary">movie</span>
              {:else}
                <img src={m.url} alt="" class="w-full h-full object-cover" />
              {/if}
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs uppercase text-on-surface-variant">{isVideo(m) ? 'Vídeo' : 'Foto'}</p>
              <a
                href={m.url}
                target="_blank"
                rel="noopener"
                class="text-sm text-secondary break-all hover:underline"
              >
                {m.url}
              </a>
              <button
                type="button"
                class="block mt-2 text-error text-sm font-medium"
                disabled={mediaBusy}
                on:click={() => deleteMedia(m)}
              >
                Remover
              </button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}

    <div class="grid sm:grid-cols-2 gap-3 pt-2">
      <div class="border border-dashed border-outline-variant rounded-lg p-3">
        <p class="ahq-label mb-2">Adicionar foto</p>
        <label class="ahq-btn-ghost h-10 px-3 text-sm inline-flex items-center cursor-pointer">
          <span class="material-symbols-outlined text-[18px] mr-1">upload</span>
          Enviar arquivo
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp,image/gif"
            class="hidden"
            on:change={onPhotoFile}
            disabled={mediaBusy}
          />
        </label>
        <div class="mt-2 flex gap-2">
          <input
            type="url"
            class="ahq-input h-9 text-sm flex-1"
            placeholder="ou URL…"
            id="photo-url"
          />
          <button
            type="button"
            class="ahq-btn-ghost h-9 px-3 text-sm"
            disabled={mediaBusy}
            on:click={() => {
              const el = document.getElementById('photo-url')
              addPhotoURL(el?.value)
              if (el) el.value = ''
            }}
          >
            URL
          </button>
        </div>
      </div>
      <div class="border border-dashed border-outline-variant rounded-lg p-3">
        <p class="ahq-label mb-2">Adicionar vídeo</p>
        <input
          type="url"
          class="ahq-input h-9 text-sm w-full mb-2"
          placeholder="https://youtube.com/… ou .mp4"
          bind:value={videoURL}
          disabled={mediaBusy}
        />
        <button
          type="button"
          class="ahq-btn-primary h-9 px-4 text-sm"
          disabled={mediaBusy || !videoURL.trim()}
          on:click={addVideo}
        >
          Adicionar vídeo
        </button>
      </div>
    </div>
  </section>
</AppShell>
