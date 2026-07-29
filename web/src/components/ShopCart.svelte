<script>
  import { onMount } from 'svelte'
  import { shopCart, cartCount, cartSummary, formatCartBRL } from '@/lib/shopCart.js'
  import { buildWhatsAppCartURL } from '@/lib/shopWhatsApp.js'

  export let whatsappDigits = ''
  export let whatsappSet = false
  export let companyName = ''
  /** @type {'dark'|'light'} */
  export let theme = 'dark'

  let open = false
  let toast = ''
  let toastTimer

  onMount(() => {
    shopCart.hydrate()
    const onKey = (e) => {
      if (e.key === 'Escape' && open) open = false
    }
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
    }
  })

  $: items = $shopCart
  $: count = $cartCount
  $: summary = $cartSummary
  $: orderURL =
    whatsappSet && items.length > 0
      ? buildWhatsAppCartURL(whatsappDigits, items, { companyName })
      : ''

  // Lock page scroll while drawer is open (portal lives on body)
  $: if (typeof document !== 'undefined') {
    document.body.style.overflow = open ? 'hidden' : ''
  }

  /**
   * Move node to document.body so position:fixed is not trapped by sticky
   * header (backdrop-filter creates a containing block).
   * @param {HTMLElement} node
   */
  function portal(node) {
    document.body.appendChild(node)
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node)
      },
    }
  }

  function showToast(msg) {
    toast = msg
    clearTimeout(toastTimer)
    toastTimer = setTimeout(() => {
      toast = ''
    }, 2200)
  }

  /** @param {Record<string, any>} product @param {number} [qty] */
  export function addProduct(product, qty = 1) {
    const ok = shopCart.add(product, qty)
    if (ok) {
      showToast('Adicionado ao carrinho')
      open = true
    } else {
      showToast('Quantidade máxima no estoque')
    }
    return ok
  }

  function bump(id, delta) {
    const it = items.find((x) => String(x.id) === String(id))
    if (!it) return
    shopCart.setQty(id, it.qty + delta)
  }

  function sendOrder() {
    if (!orderURL) return
    window.open(orderURL, '_blank', 'noopener,noreferrer')
  }

  function close() {
    open = false
  }
</script>

<button
  type="button"
  class="shop-cart-btn"
  on:click={() => (open = true)}
  aria-label="Abrir carrinho"
  title="Carrinho"
>
  <span class="material-symbols-outlined">shopping_cart</span>
  {#if count > 0}
    <span class="shop-cart-badge">{count > 99 ? '99+' : count}</span>
  {/if}
</button>

{#if open}
  <!-- Portal to <body>: escapes sticky header + backdrop-filter containing block -->
  <div
    use:portal
    class="shop-cart-layer"
    data-theme={theme === 'light' ? 'light' : 'dark'}
  >
    <div
      class="shop-cart-backdrop"
      role="presentation"
      on:click={close}
      on:keydown={(e) => e.key === 'Escape' && close()}
    ></div>
    <div class="shop-cart-drawer" role="dialog" aria-label="Carrinho" aria-modal="true">
      <div class="shop-cart-head">
        <div>
          <h2>Carrinho</h2>
          <p>{count === 0 ? 'Vazio' : `${count} ${count === 1 ? 'item' : 'itens'}`}</p>
        </div>
        <button type="button" class="shop-cart-close" on:click={close} aria-label="Fechar">
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="shop-cart-body">
        {#if items.length === 0}
          <div class="shop-cart-empty">
            <span class="material-symbols-outlined">shopping_bag</span>
            <p>Seu carrinho está vazio.</p>
            <p class="shop-cart-empty-hint">Toque em “Adicionar” nos monitores do catálogo.</p>
          </div>
        {:else}
          <ul class="shop-cart-list">
            {#each items as it (it.id)}
              <li class="shop-cart-line">
                <div class="shop-cart-thumb">
                  {#if it.thumbUrl}
                    <img src={it.thumbUrl} alt="" />
                  {:else}
                    <span class="material-symbols-outlined">monitor</span>
                  {/if}
                </div>
                <div class="shop-cart-line-main">
                  <div class="shop-cart-line-name">{it.name}</div>
                  <div class="shop-cart-line-price">
                    {#if it.priceCents > 0}
                      <span class="shop-cart-pix"
                        >{formatCartBRL(Math.floor((it.priceCents * 90) / 100))}</span
                      >
                      <span class="shop-cart-pix-tag">PIX un.</span>
                    {:else}
                      <span class="shop-cart-pix">Consulte</span>
                    {/if}
                  </div>
                  <div class="shop-cart-qty">
                    <button type="button" on:click={() => bump(it.id, -1)} aria-label="Diminuir"
                      >−</button
                    >
                    <span>{it.qty}</span>
                    <button
                      type="button"
                      on:click={() => bump(it.id, 1)}
                      aria-label="Aumentar"
                      disabled={it.qty >= it.maxQty}
                    >
                      +
                    </button>
                    <button
                      type="button"
                      class="shop-cart-remove"
                      on:click={() => shopCart.remove(it.id)}
                    >
                      Remover
                    </button>
                  </div>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </div>

      {#if items.length > 0}
        <div class="shop-cart-foot">
          {#if summary.priced > 0}
            <div class="shop-cart-totals">
              <div>
                <span>Total PIX</span>
                <strong>{formatCartBRL(summary.pixCents)}</strong>
              </div>
              <div class="shop-cart-totals-muted">
                <span>Cartão</span>
                <span>{formatCartBRL(summary.cardCents)}</span>
              </div>
            </div>
          {/if}
          {#if whatsappSet && orderURL}
            <button type="button" class="shop-btn-wa shop-cart-checkout" on:click={sendOrder}>
              <span class="material-symbols-outlined" style="font-size:18px">chat</span>
              Enviar pedido no WhatsApp
            </button>
          {:else}
            <p class="shop-cart-wa-off">WhatsApp da loja não configurado.</p>
          {/if}
          <button type="button" class="shop-cart-clear" on:click={() => shopCart.clear()}>
            Esvaziar carrinho
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}

{#if toast}
  <div use:portal class="shop-cart-toast" role="status">{toast}</div>
{/if}
