<script>
  import { inertia } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import ShopCart from '@/components/ShopCart.svelte'
  import MonitorIcon from '@/components/MonitorIcon.svelte'
  import PixEletroMark from '@/components/PixEletroMark.svelte'
  import {
    applyShopThemeToDocument,
    getShopTheme,
    setShopTheme,
    shopRootClass,
  } from '@/lib/shopTheme.js'

  export let lotOffer = {}
  export let shopShowQty = true
  export let companyName = 'Pix Eletro'
  export let whatsappSet = false
  export let whatsappDigits = ''
  export let site = {}
  /** @type {{ title?: string, description?: string, image?: string, url?: string }} */
  export let og = {}

  $: photos = Array.isArray(lotOffer.photos) ? lotOffer.photos : []
  $: videos = Array.isArray(lotOffer.videos) ? lotOffer.videos : []
  $: units = Array.isArray(lotOffer.items) ? lotOffer.items : []
  $: mediaItems = [
    ...photos.map((p) => ({
      key: `photo-${p.id}`,
      kind: 'photo',
      url: p.url,
      thumb: p.thumbUrl || p.url,
      poster: p.thumbUrl || p.url,
      label: p.productName || '',
    })),
    ...videos.map((v) => {
      const cover = photos.find((p) => p.productId === v.productId) || photos[0] || {}
      const poster = cover.thumbUrl || cover.url || ''
      return {
        key: `video-${v.id}`,
        kind: 'video',
        url: v.url,
        thumb: poster,
        poster,
        label: v.productName || '',
      }
    }),
  ]
  $: firstMedia = mediaItems[0] || null
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

  /** @type {import('@/components/ShopCart.svelte').default | undefined} */
  let cart
  /** @type {'dark'|'light'} */
  let theme = 'dark'
  $: rootClass = shopRootClass(theme)
  let lightboxOpen = false

  onMount(() => {
    if (!document.querySelector('link[data-shop-css]')) {
      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.href = '/static/css/shop.css?v=18'
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

  function addLot() {
    cart?.addProduct(
      {
        ...lotOffer,
        id: 'lot',
        isLot: true,
        fixedPix: true,
        qtyInStock: 1,
        thumbUrl: lotOffer.thumbUrl || photos[0]?.url || '',
      },
      1,
    )
  }
</script>

<svelte:head>
  <title>{og.title || `${lotOffer.name || 'Lote'} — ${companyName}`}</title>
  {#if og.description}
    <meta name="description" content={og.description} />
  {/if}
  <meta property="og:type" content="product" />
  <meta property="og:site_name" content={companyName} />
  <meta property="og:title" content={og.title || `${lotOffer.name || 'Lote'} — ${companyName}`} />
  {#if og.description}
    <meta property="og:description" content={og.description} />
  {/if}
  {#if og.url}<meta property="og:url" content={og.url} />{/if}
  {#if og.image}
    <meta property="og:image" content={og.image} />
  {/if}
  <link rel="stylesheet" href="/static/css/shop.css?v=18" data-shop-css="1" />
</svelte:head>

<div class={rootClass}>
  <div class="shop-banner">Lote fechado · preço único · fotos e vídeos de cada unidade</div>

  <header class="shop-header">
    <div class="shop-header-inner" style="min-height:3.5rem">
      <a href="/" use:inertia class="shop-back">
        <span class="material-symbols-outlined" style="font-size:18px">arrow_back</span>
        Catálogo
      </a>
      <div style="display:flex;align-items:center;gap:10px">
        <PixEletroMark size={28} />
        <span class="shop-logo-title" style="font-size:1.15rem">Pix Eletro<span>.</span></span>
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
        <ShopCart bind:this={cart} {whatsappDigits} {whatsappSet} {companyName} {theme} />
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
              preload="none"
              poster={activeMedia.poster || undefined}
              src={activeMedia.url}
            >
              <source src={activeMedia.url} type="video/mp4" />
            </video>
          {:else if activeMedia?.url}
            <button
              type="button"
              class="shop-detail-main-hit"
              on:click={() => openLightbox(activeMedia.key)}
              aria-label="Ampliar foto"
            >
              <img src={activeMedia.url} alt={activeMedia.label || lotOffer.name} />
            </button>
          {:else}
            <div class="shop-card-media-fallback">
              <MonitorIcon size={72} />
            </div>
          {/if}
          <span class="shop-badge shop-badge-ok" style="top:1rem;left:1rem">Lote</span>
        </div>
        {#if mediaItems.length > 1}
          <div class="shop-thumbs shop-lot-thumbs" role="listbox" aria-label="Mídia do lote">
            {#each mediaItems as m (m.key)}
              <button
                type="button"
                class:is-on={activeMedia?.key === m.key}
                class:is-video={m.kind === 'video'}
                on:click={() => selectMedia(m.key)}
                on:dblclick={() => openLightbox(m.key)}
                title={m.label || (m.kind === 'video' ? 'Vídeo' : 'Foto')}
              >
                {#if m.kind === 'video'}
                  {#if m.thumb || m.poster}
                    <img src={m.thumb || m.poster} alt="" loading="lazy" decoding="async" />
                  {:else}
                    <span class="shop-thumb-video-fallback"></span>
                  {/if}
                  <span class="shop-thumb-video-badge" aria-hidden="true">
                    <span class="material-symbols-outlined">play_circle</span>
                  </span>
                {:else}
                  <img src={m.thumb || m.url} alt="" loading="lazy" decoding="async" />
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <div>
        <p class="shop-brand">Lote fechado</p>
        <h1 class="shop-detail-title">{lotOffer.name || 'Lote de monitores'}</h1>
        <p class="shop-detail-meta">
          {#if shopShowQty}{lotOffer.qty || 0} unidades · {/if}{lotOffer.photoCount || 0} fotos · {lotOffer.videoCount || 0} vídeos
        </p>
        <div class="shop-detail-price">
          <div>
            <span class="shop-pix">{lotOffer.pixPrice || lotOffer.price}</span>
            <span class="shop-pix-tag">preço do lote</span>
          </div>
          {#if lotOffer.retailPrice}
            <p class="shop-card-price" style="margin-top:6px">
              Unidades no catálogo somam {lotOffer.retailPrice}
            </p>
          {/if}
        </div>
        {#if lotOffer.description}
          <div class="shop-desc">
            <h2>O lote</h2>
            <p>{lotOffer.description}</p>
          </div>
        {/if}

        <button type="button" class="shop-cta shop-cta-cart" on:click={addLot}>
          <span class="material-symbols-outlined" style="font-size:20px;vertical-align:-4px">
            add_shopping_cart
          </span>
          Quero o lote no carrinho
        </button>
        {#if whatsappSet && lotOffer.whatsappUrl}
          <a
            href={lotOffer.whatsappUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="shop-cta shop-cta-secondary"
          >
            Pedir o lote no WhatsApp
          </a>
        {/if}
      </div>
    </div>

    {#if units.length}
      <div class="shop-lot-units">
        <h2>Monitores inclusos</h2>
        <ul>
          {#each units as u (u.productId)}
            <li>
              <a href={u.href || `/produto/${u.slug}`} use:inertia class="shop-lot-unit">
                <div class="shop-lot-unit-thumb">
                  {#if u.thumbUrl}
                    <img src={u.thumbUrl} alt="" />
                  {:else}
                    <MonitorIcon size={28} />
                  {/if}
                </div>
                <div>
                  <strong>{u.qty}× {u.name}</strong>
                  {#if u.price && u.price !== 'R$ 0,00'}
                    <span>{u.price} un. no catálogo</span>
                  {/if}
                </div>
              </a>
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
</div>

{#if lightboxOpen && activeMedia}
  <div
    use:portal
    class="shop-lightbox"
    role="dialog"
    aria-modal="true"
    aria-label="Galeria do lote"
    data-theme={theme === 'light' ? 'light' : 'dark'}
  >
    <button type="button" class="shop-lightbox-backdrop" aria-label="Fechar" on:click={closeLightbox}
    ></button>
    <div class="shop-lightbox-bar">
      <span class="shop-lightbox-count">
        {activeIndex + 1} / {mediaItems.length}
        {#if activeMedia.label}
          · {activeMedia.label}
        {/if}
      </span>
      <button type="button" class="shop-lightbox-close" on:click={closeLightbox} aria-label="Fechar">
        <span class="material-symbols-outlined">close</span>
      </button>
    </div>
    <div class="shop-lightbox-stage">
      {#if mediaItems.length > 1}
        <button type="button" class="shop-lightbox-nav shop-lightbox-prev" on:click={() => slide(-1)}>
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
            poster={activeMedia.poster || undefined}
            src={activeMedia.url}
          >
            <source src={activeMedia.url} type="video/mp4" />
          </video>
        {:else}
          <img class="shop-lightbox-img" src={activeMedia.url} alt={activeMedia.label || 'Foto'} />
        {/if}
      </div>
      {#if mediaItems.length > 1}
        <button type="button" class="shop-lightbox-nav shop-lightbox-next" on:click={() => slide(1)}>
          <span class="material-symbols-outlined">chevron_right</span>
        </button>
      {/if}
    </div>
  </div>
{/if}
