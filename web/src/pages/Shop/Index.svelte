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

  export let products = []
  export let companyName = 'Puppe Eletrônicos'
  export let whatsappSet = false
  export let whatsappDigits = ''
  export let whatsappHint = ''
  export let site = {}

  let searchQuery = ''
  let category = 'all'
  let sortBy = 'featured'
  let waOpen = false
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
  })

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    setShopTheme(theme)
  }

  function addToCart(p) {
    cart?.addProduct(p, 1)
  }

  $: list = Array.isArray(products) ? products : []

  $: categories = [
    { id: 'all', label: 'TODOS', count: list.length },
    {
      id: 'com-base',
      label: 'COM BASE',
      count: list.filter((p) => p.category === 'com-base').length,
    },
    {
      id: 'sem-base',
      label: 'SEM BASE',
      count: list.filter((p) => p.category === 'sem-base').length,
    },
    {
      id: 'defeito',
      label: 'DEFEITO / PEÇAS',
      count: list.filter((p) => p.category === 'defeito').length,
    },
  ]

  $: filtered = list
    .filter((p) => {
      if (category !== 'all' && p.category !== category) return false
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase()
        return (
          (p.name || '').toLowerCase().includes(q) ||
          (p.brand || '').toLowerCase().includes(q) ||
          (p.condition || '').toLowerCase().includes(q)
        )
      }
      return true
    })
    .slice()
    .sort((a, b) => {
      if (sortBy === 'price-asc') return (a.priceCents || 0) - (b.priceCents || 0)
      if (sortBy === 'price-desc') return (b.priceCents || 0) - (a.priceCents || 0)
      if (a.badge === 'DEFEITO' && b.badge !== 'DEFEITO') return 1
      if (b.badge === 'DEFEITO' && a.badge !== 'DEFEITO') return -1
      return (b.priceCents || 0) - (a.priceCents || 0)
    })

  function generalWhatsApp(text) {
    if (!whatsappDigits) return
    const t =
      text ||
      `Olá ${companyName}! Gostaria de tirar dúvidas sobre os monitores do catálogo.`
    window.open(`https://wa.me/${whatsappDigits}?text=${encodeURIComponent(t)}`, '_blank')
  }

  function scrollCatalog() {
    document.getElementById('catalog')?.scrollIntoView({ behavior: 'smooth' })
  }
</script>

<svelte:head>
  <title>{companyName} — Catálogo</title>
  <meta
    name="description"
    content="Monitores usados testados. Pedido direto no WhatsApp. 10% OFF no PIX."
  />
  <link rel="stylesheet" href="/static/css/shop.css?v=4" data-shop-css="1" />
</svelte:head>

<div class={rootClass}>
  <div class="shop-banner">
    <span class="material-symbols-outlined" style="font-size:16px">bolt</span>
    Monte o carrinho · envie o pedido no WhatsApp · 10% OFF no PIX
  </div>

  <header class="shop-header">
    <div class="shop-header-inner">
      <a href="/" use:inertia class="shop-logo">
        <div class="shop-logo-mark">
          <span class="material-symbols-outlined">storefront</span>
        </div>
        <div>
          <div style="display:flex;align-items:center;gap:8px">
            <div class="shop-logo-title">Puppe<span>.</span></div>
            <span class="shop-pill">Catálogo</span>
          </div>
          <p class="shop-sub">Monitores & Tech · WhatsApp Orders</p>
        </div>
      </a>

      <div class="shop-search shop-search-desktop">
        <span class="material-symbols-outlined shop-search-icon">search</span>
        <input type="search" bind:value={searchQuery} placeholder="Buscar monitor Dell, Samsung, LG…" />
      </div>

      <div style="display:flex;align-items:center;gap:8px">
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
        {#if whatsappSet}
          <button type="button" class="shop-btn-wa shop-btn-wa-header" on:click={() => generalWhatsApp()}>
            <span class="material-symbols-outlined" style="font-size:18px">chat</span>
            WhatsApp
          </button>
        {/if}
      </div>

      <div class="shop-search shop-search-mobile">
        <span class="material-symbols-outlined shop-search-icon">search</span>
        <input type="search" bind:value={searchQuery} placeholder="Buscar no catálogo…" />
      </div>
    </div>
  </header>

  <section class="shop-hero">
    <div class="shop-hero-badge">
      <span class="material-symbols-outlined" style="font-size:16px">chat</span>
      Fechamento direto no WhatsApp
    </div>
    <h1>Monitores <em>testados</em> com preço de catálogo</h1>
    <p>
      Monte o carrinho com os monitores que quiser, veja fotos e vídeos reais e envie o pedido no
      WhatsApp. PIX com 10% off ou cartão.
    </p>
    <div class="shop-hero-actions">
      <button type="button" class="shop-btn-wa" on:click={scrollCatalog}>
        <span class="material-symbols-outlined" style="font-size:18px">shopping_bag</span>
        Ver catálogo
      </button>
      {#if whatsappSet}
        <button type="button" class="shop-btn-ghost" on:click={() => generalWhatsApp()}>
          <span class="material-symbols-outlined" style="font-size:18px;color:#25D366">chat</span>
          Falar no WhatsApp
        </button>
      {/if}
    </div>
    <div class="shop-props">
      <div class="shop-prop">
        <div class="shop-prop-icon"><span class="material-symbols-outlined">bolt</span></div>
        <div>
          <strong>10% OFF PIX</strong>
          <span>Desconto direto</span>
        </div>
      </div>
      <div class="shop-prop">
        <div class="shop-prop-icon"><span class="material-symbols-outlined">local_shipping</span></div>
        <div>
          <strong>Entrega / retirada</strong>
          <span>A combinar</span>
        </div>
      </div>
      <div class="shop-prop">
        <div class="shop-prop-icon"><span class="material-symbols-outlined">verified</span></div>
        <div>
          <strong>Fotos reais</strong>
          <span>Com vídeo de teste</span>
        </div>
      </div>
    </div>
  </section>

  <div class="shop-cats">
    <div class="shop-cats-inner">
      {#each categories as cat (cat.id)}
        <button
          type="button"
          class="shop-cat"
          class:is-on={category === cat.id}
          on:click={() => (category = cat.id)}
        >
          {cat.label}
          <span class="shop-cat-count">{cat.count}</span>
        </button>
      {/each}
    </div>
  </div>

  <main id="catalog" class="shop-main">
    {#if !whatsappSet}
      <div class="shop-warn">WhatsApp da loja ainda não configurado (Configurações no ERP).</div>
    {/if}

    <div class="shop-main-head">
      <h2>
        Catálogo
        <span class="shop-count">{filtered.length} itens</span>
      </h2>
      <div class="shop-sort">
        <span class="material-symbols-outlined" style="font-size:16px">swap_vert</span>
        <select bind:value={sortBy}>
          <option value="featured">Destaques</option>
          <option value="price-asc">Menor preço</option>
          <option value="price-desc">Maior preço</option>
        </select>
      </div>
    </div>

    {#if filtered.length === 0}
      <div class="shop-empty">
        <span class="material-symbols-outlined" style="font-size:48px">search_off</span>
        <h3>Nenhum produto</h3>
        <p>Ajuste a busca ou o filtro de categoria.</p>
        <button
          type="button"
          class="shop-btn-wa"
          style="margin-top:1rem"
          on:click={() => {
            searchQuery = ''
            category = 'all'
          }}
        >
          Limpar filtros
        </button>
      </div>
    {:else}
      <div class="shop-grid">
        {#each filtered as p (p.id)}
          <article class="shop-card">
            <a href={`/produto/${p.id}`} use:inertia class="shop-card-media">
              {#if p.thumbUrl}
                <img src={p.thumbUrl} alt={p.name} loading="lazy" />
              {/if}
              {#if p.badge}
                <span class="shop-badge {p.badge === 'DEFEITO' ? 'shop-badge-bad' : 'shop-badge-ok'}">
                  {p.badge}
                </span>
              {/if}
              {#if p.videoCount > 0}
                <span class="shop-vid">
                  <span class="material-symbols-outlined" style="font-size:14px">movie</span>
                  {p.videoCount}
                </span>
              {/if}
            </a>
            <div class="shop-card-body">
              <div style="display:flex;justify-content:space-between;align-items:center">
                <span class="shop-brand">{p.brand || 'MONITOR'}</span>
                <span style="font-size:10px;color:#737373;font-weight:700">{p.qtyInStock || 0} un.</span>
              </div>
              <a href={`/produto/${p.id}`} use:inertia class="shop-card-title">{p.name}</a>
              <div class="shop-price-box">
                {#if p.pixPrice}
                  <div>
                    <span class="shop-pix">{p.pixPrice}</span>
                    <span class="shop-pix-tag">PIX</span>
                  </div>
                  <div class="shop-card-price">{p.price} no cartão</div>
                {:else}
                  <span class="shop-pix">{p.price}</span>
                {/if}
              </div>
              <div class="shop-card-actions">
                <a href={`/produto/${p.id}`} use:inertia class="shop-btn-eye" title="Ver">
                  <span class="material-symbols-outlined" style="font-size:18px">visibility</span>
                </a>
                <button
                  type="button"
                  class="shop-btn-wa"
                  on:click={() => addToCart(p)}
                  title="Adicionar ao carrinho"
                >
                  <span class="material-symbols-outlined" style="font-size:18px">add_shopping_cart</span>
                  Adicionar
                </button>
              </div>
            </div>
          </article>
        {/each}
      </div>
    {/if}
  </main>

  <footer class="shop-footer">
    <div class="shop-footer-props">
      <div class="shop-footer-prop">
        <div class="ico"><span class="material-symbols-outlined">chat</span></div>
        <div>
          <h4>Pedido no WhatsApp</h4>
          <p>Sem cadastro. Monte o carrinho e envie o pedido completo.</p>
        </div>
      </div>
      <div class="shop-footer-prop">
        <div class="ico"><span class="material-symbols-outlined">qr_code_2</span></div>
        <div>
          <h4>10% OFF no PIX</h4>
          <p>Melhor preço no catálogo de monitores usados testados.</p>
        </div>
      </div>
      <div class="shop-footer-prop">
        <div class="ico" style="background:#262626;color:#25D366;border:1px solid #404040">
          <span class="material-symbols-outlined">verified_user</span>
        </div>
        <div>
          <h4>Fotos e vídeos reais</h4>
          <p>Cada unidade testada. Defeitos sempre informados.</p>
        </div>
      </div>
    </div>
    <div class="shop-footer-bottom">
      <div>
        <strong style="color:#fff;font-weight:900;letter-spacing:-0.03em">
          PUPPE<span style="color:#25D366">.</span>
        </strong>
        <span style="margin-left:8px">{companyName}</span>
      </div>
      <p>© {new Date().getFullYear()} · WhatsApp direto</p>
    </div>
  </footer>

  {#if whatsappSet}
    <div class="shop-float">
      {#if waOpen}
        <div class="shop-float-panel">
          <div class="shop-float-head">
            <div>
              <h3>{companyName}</h3>
              <p>● Online no WhatsApp</p>
            </div>
            <button
              type="button"
              on:click={() => (waOpen = false)}
              style="background:transparent;border:0;color:#fff;cursor:pointer"
            >
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <div class="shop-float-body">
            <button
              type="button"
              on:click={() =>
                generalWhatsApp('Olá! Gostaria de falar com um atendente sobre os monitores.')}
            >
              <span class="material-symbols-outlined" style="color:#25D366;font-size:18px">support_agent</span>
              Falar com atendente
            </button>
            <button
              type="button"
              on:click={() =>
                generalWhatsApp('Olá! Quero combinar retirada ou entrega de um monitor do catálogo.')}
            >
              <span class="material-symbols-outlined" style="color:#fbbf24;font-size:18px">location_on</span>
              Retirada / entrega
            </button>
          </div>
        </div>
      {/if}
      <button type="button" class="shop-float-btn" on:click={() => (waOpen = !waOpen)} aria-label="WhatsApp">
        <span class="material-symbols-outlined" style="font-size:28px">chat</span>
      </button>
    </div>
  {/if}
</div>
