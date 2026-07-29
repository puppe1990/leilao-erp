<script>
  import { inertia } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import ShopCart from '@/components/ShopCart.svelte'
  import {
    applyShopThemeToDocument,
    getShopTheme,
    setShopTheme,
    shopRootClass,
  } from '@/lib/shopTheme.js'

  export let product = {}
  export let companyName = 'Puppe Eletrônicos'
  export let whatsappSet = false
  export let whatsappDigits = ''
  export let site = {}
  /** @type {{ title?: string, description?: string, image?: string, url?: string }} */
  export let og = {}

  $: photos = Array.isArray(product.photos) ? product.photos : []
  $: videos = Array.isArray(product.videos) ? product.videos : []
  /** Photos first, videos last — unified gallery. */
  $: mediaItems = [
    ...photos.map((p) => ({
      key: `photo-${p.id}`,
      kind: 'photo',
      url: p.url,
      poster: p.url,
    })),
    ...videos.map((v) => ({
      key: `video-${v.id}`,
      kind: 'video',
      url: v.url,
      poster: photos[0]?.url || product.thumbUrl || '',
    })),
  ]
  $: firstMedia = mediaItems[0] || null
  /** @type {string} */
  let activeKey = ''
  $: if (firstMedia && !activeKey) activeKey = firstMedia.key
  $: if (activeKey && !mediaItems.some((m) => m.key === activeKey) && firstMedia) {
    activeKey = firstMedia.key
  }
  $: activeMedia = mediaItems.find((m) => m.key === activeKey) || firstMedia
  $: activeIndex = Math.max(
    0,
    mediaItems.findIndex((m) => m.key === (activeMedia?.key || '')),
  )
  $: mainPhoto = photos[0]?.url || product.thumbUrl || ''
  /** @type {import('@/components/ShopCart.svelte').default | undefined} */
  let cart
  /** @type {'dark'|'light'} */
  let theme = 'dark'
  $: rootClass = shopRootClass(theme)

  let lightboxOpen = false
  /** touch swipe */
  let touchStartX = 0
  let touchDeltaX = 0

  onMount(() => {
    if (!document.querySelector('link[data-shop-css]')) {
      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.href = '/static/css/shop.css?v=6'
      link.setAttribute('data-shop-css', '1')
      document.head.appendChild(link)
    }
    theme = getShopTheme()
    applyShopThemeToDocument(theme)

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

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    setShopTheme(theme)
  }

  function selectMedia(key) {
    activeKey = key
  }

  function openLightbox(key) {
    if (key) activeKey = key
    if (!activeMedia) return
    lightboxOpen = true
    document.body.style.overflow = 'hidden'
  }

  function closeLightbox() {
    lightboxOpen = false
    document.body.style.overflow = ''
  }

  function slide(dir) {
    if (mediaItems.length < 2) return
    const next = (activeIndex + dir + mediaItems.length) % mediaItems.length
    activeKey = mediaItems[next].key
  }

  function portal(node) {
    document.body.appendChild(node)
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node)
      },
    }
  }

  function onTouchStart(e) {
    touchStartX = e.changedTouches?.[0]?.clientX || 0
    touchDeltaX = 0
  }

  function onTouchMove(e) {
    const x = e.changedTouches?.[0]?.clientX || 0
    touchDeltaX = x - touchStartX
  }

  function onTouchEnd() {
    if (Math.abs(touchDeltaX) < 50) return
    if (touchDeltaX < 0) slide(1)
    else slide(-1)
    touchDeltaX = 0
  }

  function addToCart() {
    cart?.addProduct(
      {
        ...product,
        thumbUrl: product.thumbUrl || photos[0]?.url || '',
      },
      1,
    )
  }
</script>

<svelte:head>
  <title>{og.title || `${product.name || 'Produto'} — ${companyName}`}</title>
  {#if og.description}
    <meta name="description" content={og.description} />
  {/if}
  <meta property="og:type" content="product" />
  <meta property="og:site_name" content={companyName} />
  <meta property="og:title" content={og.title || `${product.name || 'Produto'} — ${companyName}`} />
  {#if og.description}
    <meta property="og:description" content={og.description} />
  {/if}
  {#if og.url}<meta property="og:url" content={og.url} />{/if}
  {#if og.image || product.thumbUrl || photos[0]?.url}
    <meta
      property="og:image"
      content={og.image || product.thumbUrl || photos[0]?.url}
    />
  {/if}
  <meta name="twitter:card" content="summary_large_image" />
  <meta
    name="twitter:title"
    content={og.title || `${product.name || 'Produto'} — ${companyName}`}
  />
  {#if og.description}
    <meta name="twitter:description" content={og.description} />
  {/if}
  {#if og.image || product.thumbUrl || photos[0]?.url}
    <meta name="twitter:image" content={og.image || product.thumbUrl || photos[0]?.url} />
  {/if}
  <link rel="stylesheet" href="/static/css/shop.css?v=6" data-shop-css="1" />
</svelte:head>

<div class={rootClass}>
  <div class="shop-banner">Monte o carrinho · 10% OFF no PIX · pedido no WhatsApp</div>

  <header class="shop-header">
    <div class="shop-header-inner" style="min-height:3.5rem">
      <a href="/" use:inertia class="shop-back">
        <span class="material-symbols-outlined" style="font-size:18px">arrow_back</span>
        Catálogo
      </a>
      <div style="display:flex;align-items:center;gap:10px">
        <span class="shop-logo-title" style="font-size:1.15rem">Puppe<span>.</span></span>
        <button
          type="button"
          class="shop-theme-btn"
          on:click={toggleTheme}
          title={theme === 'dark' ? 'Modo claro' : 'Modo escuro'}
          aria-label={theme === 'dark' ? 'Ativar modo claro' : 'Ativar modo escuro'}
        >
          <span class="material-symbols-outlined">
            {theme === 'dark' ? 'light_mode' : 'dark_mode'}
          </span>
        </button>
        <ShopCart
          bind:this={cart}
          {whatsappDigits}
          {whatsappSet}
          {companyName}
          {theme}
        />
      </div>
    </div>
  </header>

  <div class="shop-detail">
    <div class="shop-detail-grid">
      <div>
        <div class="shop-detail-main">
          {#if activeMedia?.kind === 'video'}
            <video
              class="shop-detail-video"
              controls
              playsinline
              muted
              preload="metadata"
              poster={activeMedia.poster || mainPhoto || undefined}
              src={activeMedia.url}
              on:click|stopPropagation
            >
              <source src={activeMedia.url} type="video/mp4" />
            </video>
            <button
              type="button"
              class="shop-detail-zoom"
              on:click={() => openLightbox(activeMedia.key)}
              title="Ampliar"
              aria-label="Ampliar mídia"
            >
              <span class="material-symbols-outlined">open_in_full</span>
            </button>
          {:else if activeMedia?.url}
            <button
              type="button"
              class="shop-detail-main-hit"
              on:click={() => openLightbox(activeMedia.key)}
              aria-label="Ampliar foto"
            >
              <img src={activeMedia.url} alt={product.name} />
              <span class="shop-detail-zoom-hint" aria-hidden="true">
                <span class="material-symbols-outlined">zoom_in</span>
              </span>
            </button>
          {/if}
          {#if product.badge}
            <span
              class="shop-badge {product.badge === 'DEFEITO' ? 'shop-badge-bad' : 'shop-badge-ok'}"
              style="top:1rem;left:1rem"
            >
              {product.badge}
            </span>
          {/if}
        </div>
        {#if mediaItems.length > 1}
          <div class="shop-thumbs" role="listbox" aria-label="Mídia do produto">
            {#each mediaItems as m (m.key)}
              <button
                type="button"
                class:is-on={activeMedia?.key === m.key}
                class:is-video={m.kind === 'video'}
                on:click={() => selectMedia(m.key)}
                on:dblclick={() => openLightbox(m.key)}
                title={m.kind === 'video' ? 'Vídeo' : 'Foto'}
                aria-label={m.kind === 'video' ? 'Ver vídeo' : 'Ver foto'}
              >
                {#if m.kind === 'video'}
                  {#if m.poster}
                    <img src={m.poster} alt="" />
                  {:else}
                    <span class="shop-thumb-video-fallback"></span>
                  {/if}
                  <span class="shop-thumb-video-badge" aria-hidden="true">
                    <span class="material-symbols-outlined">play_circle</span>
                  </span>
                {:else}
                  <img src={m.url} alt="" />
                {/if}
              </button>
            {/each}
          </div>
        {/if}

        <div class="shop-assurances">
          <div>
            <span class="material-symbols-outlined">verified_user</span>
            Fotos e vídeos reais da unidade
          </div>
          <div>
            <span class="material-symbols-outlined">local_shipping</span>
            Entrega ou retirada a combinar no WhatsApp
          </div>
        </div>
      </div>

      <div>
        <p class="shop-brand">{product.brand || 'MONITOR'}</p>
        <h1 class="shop-detail-title">{product.name}</h1>
        <p class="shop-detail-meta">
          {product.condition || 'Usado'} · {product.qtyInStock || 0} em estoque
        </p>

        <div class="shop-detail-price">
          {#if product.pixPrice}
            <div>
              <span class="shop-pix">{product.pixPrice}</span>
              <span class="shop-pix-tag">PIX · 10% OFF</span>
            </div>
            <p class="shop-card-price" style="margin-top:6px">{product.price} no cartão</p>
          {:else}
            <span class="shop-pix">{product.price}</span>
          {/if}
        </div>

        {#if product.screenType || product.maxResolution || product.refreshRate}
          <dl class="shop-specs">
            {#if product.screenType}
              <div class="shop-spec">
                <dt>Tela</dt>
                <dd>{product.screenType}</dd>
              </div>
            {/if}
            {#if product.maxResolution}
              <div class="shop-spec">
                <dt>Resolução</dt>
                <dd>{product.maxResolution}</dd>
              </div>
            {/if}
            {#if product.refreshRate}
              <div class="shop-spec">
                <dt>Taxa</dt>
                <dd>{product.refreshRate}</dd>
              </div>
            {/if}
          </dl>
        {/if}

        {#if product.description}
          <div class="shop-desc">
            <h2>Detalhes</h2>
            <p>{product.description}</p>
          </div>
        {/if}

        <button type="button" class="shop-cta shop-cta-cart" on:click={addToCart}>
          <span class="material-symbols-outlined" style="font-size:20px;vertical-align:-4px">
            add_shopping_cart
          </span>
          Adicionar ao carrinho
        </button>
        {#if whatsappSet && product.whatsappUrl}
          <a
            href={product.whatsappUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="shop-cta shop-cta-secondary"
          >
            Pedir só este no WhatsApp
          </a>
        {:else if !whatsappSet}
          <p style="text-align:center;color:#737373;font-size:14px;margin-top:0.75rem">
            WhatsApp da loja não configurado.
          </p>
        {/if}
      </div>
    </div>
  </div>
</div>

{#if lightboxOpen && activeMedia}
  <div
    use:portal
    class="shop-lightbox"
    role="dialog"
    aria-modal="true"
    aria-label="Galeria ampliada"
    data-theme={theme === 'light' ? 'light' : 'dark'}
  >
    <button
      type="button"
      class="shop-lightbox-backdrop"
      aria-label="Fechar"
      on:click={closeLightbox}
    ></button>

    <div class="shop-lightbox-bar">
      <span class="shop-lightbox-count">
        {activeIndex + 1} / {mediaItems.length}
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

    <div
      class="shop-lightbox-stage"
      on:touchstart={onTouchStart}
      on:touchmove={onTouchMove}
      on:touchend={onTouchEnd}
      role="presentation"
    >
      {#if mediaItems.length > 1}
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
        {#if activeMedia.kind === 'video'}
          <video
            class="shop-lightbox-video"
            controls
            playsinline
            muted
            autoplay
            preload="metadata"
            poster={activeMedia.poster || mainPhoto || undefined}
            src={activeMedia.url}
          >
            <source src={activeMedia.url} type="video/mp4" />
          </video>
        {:else}
          <img class="shop-lightbox-img" src={activeMedia.url} alt={product.name || 'Foto'} />
        {/if}
      </div>

      {#if mediaItems.length > 1}
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

    {#if mediaItems.length > 1}
      <div class="shop-lightbox-dots" role="tablist" aria-label="Slides">
        {#each mediaItems as m, i (m.key)}
          <button
            type="button"
            class="shop-lightbox-dot"
            class:is-on={activeMedia.key === m.key}
            class:is-video={m.kind === 'video'}
            on:click={() => selectMedia(m.key)}
            aria-label={m.kind === 'video' ? `Vídeo ${i + 1}` : `Foto ${i + 1}`}
          ></button>
        {/each}
      </div>
    {/if}
  </div>
{/if}
