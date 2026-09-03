<script>
  import { onMount } from 'svelte'
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import MonitorIcon from '@/components/MonitorIcon.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let product = {}
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let videoURL = ''
  let mediaBusy = false
  let lightboxOpen = false
  let lightboxIndex = 0

  $: mediaList = Array.isArray(product.media) ? product.media : []
  $: lightboxItem = mediaList[lightboxIndex] || null

  function isVideo(m) {
    return m?.kind === 'video'
  }

  function portal(node) {
    document.body.appendChild(node)
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node)
      },
    }
  }

  function openLightbox(index) {
    if (!mediaList.length) return
    lightboxIndex = Math.max(0, Math.min(index, mediaList.length - 1))
    lightboxOpen = true
    document.body.style.overflow = 'hidden'
  }

  function closeLightbox() {
    lightboxOpen = false
    document.body.style.overflow = ''
  }

  function slide(dir) {
    if (mediaList.length < 2) return
    lightboxIndex = (lightboxIndex + dir + mediaList.length) % mediaList.length
  }

  onMount(() => {
    const onKey = (e) => {
      if (!lightboxOpen) return
      if (e.key === 'Escape') closeLightbox()
      if (e.key === 'ArrowRight') slide(1)
      if (e.key === 'ArrowLeft') slide(-1)
    }
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
    }
  })

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

  async function deleteMedia(m) {
    const kind = isVideo(m) ? 'vídeo' : 'foto'
    const ok = await askConfirm({
      title: `Remover ${kind}`,
      message: `Tem certeza que deseja remover este ${kind}?`,
      detail: 'A mídia deixa de aparecer no produto e no catálogo.',
      confirmLabel: 'Remover',
      tone: 'danger',
      icon: 'delete',
    })
    if (!ok) return
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

  <section class="ahq-card p-4 mb-4 space-y-4">
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

    {#if hasOlx}
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
    {:else}
      <p class="text-sm text-on-surface-variant">
        Ainda sem atributos de tela. Preencha na edição para copiar no anúncio da OLX.
      </p>
    {/if}

    <!-- Flags em linha (3 colunas), abaixo dos atributos de tela -->
    <div class="grid gap-2 sm:grid-cols-3 border-t border-outline-variant pt-3">
      <div class="rounded-xl border border-outline-variant px-3 py-2.5 space-y-1">
        <p class="text-xs font-semibold text-primary">Visível no catálogo</p>
        {#if product.shopVisible}
          <span
            class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-secondary-container text-on-secondary-container"
          >
            <MonitorIcon size={16} />
            Sim — mostrar
          </span>
        {:else}
          <span
            class="inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-full bg-surface-container-high text-on-surface-variant"
          >
            Não — ocultar
          </span>
        {/if}
      </div>
      <div class="rounded-xl border border-outline-variant px-3 py-2.5 space-y-1">
        <p class="text-xs font-semibold text-primary">Já publicado na OLX</p>
        {#if product.olxPublished}
          <span
            class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-secondary-container text-on-secondary-container"
          >
            <span class="material-symbols-outlined text-[16px]">check_circle</span>
            Sim — no ar
          </span>
        {:else}
          <span
            class="inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-full bg-surface-container-high text-on-surface-variant"
          >
            Não — falta publicar
          </span>
        {/if}
      </div>
      <div class="rounded-xl border border-outline-variant px-3 py-2.5 space-y-1">
        <p class="text-xs font-semibold text-primary">Entregar grátis pela OLX</p>
        {#if product.olxFreeShipping}
          <span
            class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-secondary-container text-on-secondary-container"
          >
            <span class="material-symbols-outlined text-[16px]">local_shipping</span>
            Sim
          </span>
        {:else}
          <span
            class="inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-full bg-surface-container-high text-on-surface-variant"
          >
            Não
          </span>
        {/if}
      </div>
    </div>
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

    {#if !mediaList.length}
      <p class="text-sm text-on-surface-variant">Nenhuma mídia ainda.</p>
    {:else}
      <ul class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {#each mediaList as m, i (m.id)}
          <li class="flex gap-3 items-start border border-outline-variant rounded-lg p-3">
            <button
              type="button"
              class="w-24 h-20 shrink-0 rounded bg-surface-container flex items-center justify-center overflow-hidden
                ring-offset-2 hover:ring-2 hover:ring-secondary/50 focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary cursor-zoom-in"
              title={isVideo(m) ? 'Ampliar vídeo' : 'Ampliar foto'}
              aria-label={isVideo(m) ? 'Abrir vídeo ampliado' : 'Abrir foto ampliada'}
              on:click={() => openLightbox(i)}
            >
              {#if isVideo(m)}
                <span class="material-symbols-outlined text-3xl text-secondary">movie</span>
              {:else}
                <img src={m.url} alt="" class="w-full h-full object-cover pointer-events-none" />
              {/if}
            </button>
            <div class="min-w-0 flex-1 flex flex-col justify-between gap-2 self-stretch">
              <p class="text-xs uppercase text-on-surface-variant font-medium tracking-wide">
                {isVideo(m) ? 'Vídeo' : 'Foto'}
              </p>
              <button
                type="button"
                class="self-start text-error text-sm font-medium"
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

{#if lightboxOpen && lightboxItem}
  <div
    use:portal
    class="shop-lightbox"
    role="dialog"
    aria-modal="true"
    aria-label="Mídia ampliada"
  >
    <button
      type="button"
      class="shop-lightbox-backdrop"
      aria-label="Fechar"
      on:click={closeLightbox}
    ></button>

    <div class="shop-lightbox-bar">
      <span class="shop-lightbox-count">
        {lightboxIndex + 1} / {mediaList.length}
      </span>
      <button
        type="button"
        class="shop-lightbox-close"
        on:click={closeLightbox}
        aria-label="Fechar galeria"
      >
        <span class="material-symbols-outlined">close</span>
      </button>
    </div>

    <div class="shop-lightbox-stage" role="presentation">
      {#if mediaList.length > 1}
        <button
          type="button"
          class="shop-lightbox-nav shop-lightbox-prev"
          on:click={() => slide(-1)}
          aria-label="Anterior"
        >
          <span class="material-symbols-outlined">chevron_left</span>
        </button>
      {/if}

      <div class="shop-lightbox-frame">
        {#if isVideo(lightboxItem)}
          <video
            class="shop-lightbox-video"
            controls
            playsinline
            muted
            autoplay
            preload="metadata"
            src={lightboxItem.url}
          >
            <source src={lightboxItem.url} type="video/mp4" />
          </video>
        {:else}
          <img
            class="shop-lightbox-img"
            src={lightboxItem.url}
            alt={product.name || 'Foto do produto'}
          />
        {/if}
      </div>

      {#if mediaList.length > 1}
        <button
          type="button"
          class="shop-lightbox-nav shop-lightbox-next"
          on:click={() => slide(1)}
          aria-label="Próxima"
        >
          <span class="material-symbols-outlined">chevron_right</span>
        </button>
      {/if}
    </div>

    {#if mediaList.length > 1}
      <div class="shop-lightbox-dots" role="tablist" aria-label="Slides">
        {#each mediaList as m, i (m.id)}
          <button
            type="button"
            class="shop-lightbox-dot"
            class:is-on={lightboxIndex === i}
            class:is-video={isVideo(m)}
            on:click={() => (lightboxIndex = i)}
            aria-label={isVideo(m) ? `Vídeo ${i + 1}` : `Foto ${i + 1}`}
          ></button>
        {/each}
      </div>
    {/if}
  </div>
{/if}
