<script>
  import { inertia } from '@inertiajs/svelte'

  export let lotOffer = null
  export let showQty = true
  /** @type {(offer: Record<string, any>) => void} */
  export let onAdd = () => {}

  $: thumbs = Array.isArray(lotOffer?.thumbs) ? lotOffer.thumbs.filter(Boolean) : []
  $: mosaic = thumbs.slice(0, 4)
  $: href = lotOffer?.href || '/produto/lote-monitores'
</script>

{#if lotOffer}
  <section class="shop-lot-banner" id="lote">
    <div class="shop-lot-banner-inner">
      <a href={href} use:inertia class="shop-lot-mosaic" aria-label="Ver lote">
        {#each mosaic as src (src)}
          <img src={src} alt="" loading="lazy" decoding="async" />
        {/each}
        {#if mosaic.length === 0}
          <span class="shop-lot-mosaic-empty">
            <span class="material-symbols-outlined">inventory_2</span>
          </span>
        {/if}
      </a>
      <div class="shop-lot-copy">
        <span class="shop-brand">Lote fechado</span>
        <h2>{lotOffer.name}</h2>
        <p>
          {lotOffer.description ||
            `Leve os ${lotOffer.qty} monitores restantes por ${lotOffer.price}. Preço fechado.`}
        </p>
        <div class="shop-lot-price">
          <span class="shop-pix">{lotOffer.pixPrice || lotOffer.price}</span>
          <span class="shop-pix-tag">preço do lote</span>
          {#if lotOffer.retailPrice}
            <div class="shop-lot-retail">
              Unidades no catálogo somam <s>{lotOffer.retailPrice}</s>
            </div>
          {/if}
        </div>
        <div class="shop-lot-actions">
          <a href={href} use:inertia class="shop-btn-wa">
            <span class="material-symbols-outlined" style="font-size:18px">photo_library</span>
            Ver fotos e vídeos
          </a>
          <button type="button" class="shop-btn-ghost" on:click={() => onAdd(lotOffer)}>
            <span class="material-symbols-outlined" style="font-size:18px">add_shopping_cart</span>
            Quero o lote
          </button>
        </div>
        <p class="shop-lot-meta">
          {lotOffer.photoCount || 0} fotos · {lotOffer.videoCount || 0} vídeos{#if showQty}
            · {lotOffer.qty} unidades{/if}
        </p>
      </div>
    </div>
  </section>
{/if}
