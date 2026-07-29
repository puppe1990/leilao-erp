<script>
  import { inertia } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import ShopCart from '@/components/ShopCart.svelte'
  import {
    applyShopThemeToDocument,
    clearShopThemeFromDocument,
    getShopTheme,
    setShopTheme,
    shopRootClass,
  } from '@/lib/shopTheme.js'

  export let product = {}
  export let companyName = 'Puppe Eletrônicos'
  export let whatsappSet = false
  export let whatsappDigits = ''
  export let site = {}

  $: photos = product.photos || []
  $: videos = product.videos || []
  $: mainPhoto = photos[0]?.url || ''
  let activePhoto = ''
  $: if (mainPhoto && !activePhoto) activePhoto = mainPhoto
  $: displayPhoto = activePhoto || mainPhoto
  /** @type {import('@/components/ShopCart.svelte').default | undefined} */
  let cart
  /** @type {'dark'|'light'} */
  let theme = 'dark'
  $: rootClass = shopRootClass(theme)

  onMount(() => {
    if (!document.querySelector('link[data-shop-css]')) {
      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.href = '/static/css/shop.css?v=4'
      link.setAttribute('data-shop-css', '1')
      document.head.appendChild(link)
    }
    theme = getShopTheme()
    applyShopThemeToDocument(theme)
    return () => clearShopThemeFromDocument()
  })

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    setShopTheme(theme)
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
  <title>{product.name || 'Produto'} — {companyName}</title>
  <link rel="stylesheet" href="/static/css/shop.css?v=4" data-shop-css="1" />
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
          {#if displayPhoto}
            <img src={displayPhoto} alt={product.name} />
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
        {#if photos.length > 1}
          <div class="shop-thumbs">
            {#each photos as ph (ph.id)}
              <button
                type="button"
                class:is-on={displayPhoto === ph.url}
                on:click={() => (activePhoto = ph.url)}
              >
                <img src={ph.url} alt="" />
              </button>
            {/each}
          </div>
        {/if}

        {#if videos.length > 0}
          <div style="margin-top:1rem">
            <h2
              style="font-size:11px;font-weight:900;text-transform:uppercase;letter-spacing:.1em;color:#25D366;margin:0 0 .5rem;display:flex;align-items:center;gap:4px"
            >
              <span class="material-symbols-outlined" style="font-size:16px">movie</span>
              Vídeos de teste
            </h2>
            {#each videos as v (v.id)}
              <div class="shop-video">
                <video controls playsinline preload="metadata" poster={mainPhoto || undefined}>
                  <source src={v.url} type="video/mp4" />
                </video>
              </div>
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
