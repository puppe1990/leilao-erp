<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let products = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let query = ''
  let editingId = null
  let editName = ''
  let editSale = ''

  /** @type {number|null} product id with media panel open */
  let mediaOpenId = null
  let videoURL = ''
  /** @type {HTMLInputElement|null} */
  let photoInput = null
  let mediaBusy = false

  $: q = query.trim().toLowerCase()
  $: filtered = (products || []).filter((p) => {
    if (!q) return true
    return String(p.name || '')
      .toLowerCase()
      .includes(q)
  })

  function startEdit(p) {
    editingId = p.id
    editName = p.name || ''
    editSale = p.salePriceRaw || ''
  }

  function cancelEdit() {
    editingId = null
  }

  function save(p) {
    router.post(
      `/products/${p.id}`,
      {
        name: editName || p.name,
        sale_price_hint: editSale,
        return_to: '/products',
      },
      {
        onSuccess: () => {
          editingId = null
        },
      },
    )
  }

  function toggleMedia(p) {
    if (mediaOpenId === p.id) {
      mediaOpenId = null
      return
    }
    mediaOpenId = p.id
    videoURL = ''
  }

  function addVideo(p) {
    const url = videoURL.trim()
    if (!url) return
    mediaBusy = true
    router.post(
      `/products/${p.id}/media`,
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

  function addPhotoURL(p, url) {
    const u = String(url || '').trim()
    if (!u) return
    mediaBusy = true
    router.post(
      `/products/${p.id}/media`,
      { kind: 'photo', url: u },
      {
        forceFormData: true,
        onFinish: () => {
          mediaBusy = false
        },
      },
    )
  }

  function onPhotoFile(p, e) {
    const file = e?.target?.files?.[0]
    if (!file) return
    mediaBusy = true
    router.post(
      `/products/${p.id}/media`,
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

  function deleteMedia(p, m) {
    if (!confirm('Remover esta mídia?')) return
    mediaBusy = true
    router.post(`/products/${p.id}/media/${m.id}/delete`, {}, {
      onFinish: () => {
        mediaBusy = false
      },
    })
  }

  function isVideo(m) {
    return m?.kind === 'video'
  }

  function isYouTube(url) {
    return /youtu\.?be|youtube\.com/i.test(String(url || ''))
  }
</script>

<AppShell {companyName} active="stock">
  <div class="flex items-start justify-between gap-3 mb-section-padding flex-wrap">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Produtos</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Catálogo com fotos e vídeos. Editar nome/preço atualiza o estoque.
      </p>
    </div>
    <div class="flex flex-wrap gap-2">
      <a href="/stock" use:inertia class="ahq-btn-ghost h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">inventory_2</span>
        Estoque
      </a>
      <a href="/lots/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">add</span>
        Novo lote
      </a>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if products.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">category</span>
      <p class="text-on-surface-variant mb-4">
        Nenhum produto ainda. Ao cadastrar lotes, os nomes viram produtos automaticamente.
      </p>
      <a href="/lots/new" use:inertia class="ahq-btn-primary">Registrar lote</a>
    </div>
  {:else}
    <div class="ahq-card overflow-hidden">
      <div class="p-3 border-b border-outline-variant flex flex-col sm:flex-row gap-2 sm:items-center sm:justify-between">
        <div class="relative flex-1 max-w-md">
          <span
            class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant text-[20px]"
          >
            search
          </span>
          <input
            type="search"
            class="ahq-input h-10 pl-10 w-full"
            placeholder="Buscar nome…"
            bind:value={query}
            aria-label="Buscar produtos"
          />
        </div>
        <p class="text-sm text-on-surface-variant">
          {filtered.length}
          {filtered.length === 1 ? 'produto' : 'produtos'}
        </p>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left min-w-[720px]">
          <thead>
            <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide">Nome</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide w-28">Tipo</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-20">
                Estoque
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-center w-24">
                Mídia
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-32">
                Preço venda
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-36">
                Ações
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline-variant">
            {#each filtered as p (p.id)}
              {#if editingId === p.id}
                <tr class="bg-secondary-container/20">
                  <td class="px-3 py-2">
                    <input class="ahq-input h-9 text-sm w-full" bind:value={editName} />
                  </td>
                  <td class="px-3 py-2 text-on-surface-variant">{p.kindLabel}</td>
                  <td class="px-3 py-2 font-mono text-right">{p.qtyInStock}</td>
                  <td class="px-3 py-2 text-center text-on-surface-variant text-xs font-mono">
                    {p.photoCount || 0}f · {p.videoCount || 0}v
                  </td>
                  <td class="px-3 py-2">
                    <input
                      class="ahq-input h-9 text-sm font-mono w-full text-right"
                      placeholder="0,00"
                      bind:value={editSale}
                    />
                  </td>
                  <td class="px-3 py-2 text-right whitespace-nowrap">
                    <button
                      type="button"
                      class="text-secondary font-medium text-sm mr-2"
                      on:click={() => save(p)}
                    >
                      Salvar
                    </button>
                    <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>
                      Cancelar
                    </button>
                  </td>
                </tr>
              {:else}
                <tr class="hover:bg-surface-container-low/80">
                  <td class="px-3 py-2.5 font-medium text-primary">{p.name}</td>
                  <td class="px-3 py-2.5">
                    {#if p.kind === 'accessory'}
                      <span class="ahq-badge-sold text-[10px]">Acessório</span>
                    {:else}
                      <span class="ahq-badge-live text-[10px]">Principal</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2.5 font-mono text-right">
                    {#if p.qtyInStock > 0}
                      <span
                        class="inline-flex min-w-[2rem] justify-center px-2 py-0.5 rounded-full
                          bg-secondary-container text-on-secondary-container font-semibold"
                      >
                        {p.qtyInStock}
                      </span>
                    {:else}
                      <span class="text-on-surface-variant">0</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2.5 text-center">
                    <span class="text-xs text-on-surface-variant font-mono">
                      <span class="material-symbols-outlined text-[16px] align-middle">image</span>
                      {p.photoCount || 0}
                      <span class="material-symbols-outlined text-[16px] align-middle ml-1">movie</span>
                      {p.videoCount || 0}
                    </span>
                  </td>
                  <td class="px-3 py-2.5 font-mono text-right font-semibold">
                    {p.salePriceHint || '—'}
                  </td>
                  <td class="px-3 py-2.5 text-right whitespace-nowrap">
                    <button
                      type="button"
                      class="text-on-surface-variant hover:text-secondary text-sm font-medium mr-2"
                      on:click={() => startEdit(p)}
                    >
                      Editar
                    </button>
                    <button
                      type="button"
                      class="text-secondary text-sm font-medium"
                      on:click={() => toggleMedia(p)}
                    >
                      {mediaOpenId === p.id ? 'Fechar mídia' : 'Fotos/vídeos'}
                    </button>
                  </td>
                </tr>
                {#if mediaOpenId === p.id}
                  <tr class="bg-surface-container-low/50">
                    <td colspan="6" class="px-3 py-4">
                      <div class="space-y-4 max-w-3xl">
                        <p class="text-sm font-semibold text-primary">Mídia — {p.name}</p>

                        <!-- existing -->
                        {#if (p.media || []).length === 0}
                          <p class="text-sm text-on-surface-variant">Nenhuma foto ou vídeo ainda.</p>
                        {:else}
                          <ul class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                            {#each p.media as m (m.id)}
                              <li
                                class="ahq-card p-3 flex gap-3 items-start border border-outline-variant"
                              >
                                <div
                                  class="w-20 h-20 shrink-0 rounded bg-surface-container flex items-center justify-center overflow-hidden"
                                >
                                  {#if isVideo(m)}
                                    <span class="material-symbols-outlined text-3xl text-secondary"
                                      >movie</span
                                    >
                                  {:else}
                                    <img src={m.url} alt="" class="w-full h-full object-cover" />
                                  {/if}
                                </div>
                                <div class="min-w-0 flex-1">
                                  <p class="text-xs uppercase tracking-wide text-on-surface-variant">
                                    {isVideo(m) ? 'Vídeo' : 'Foto'}
                                  </p>
                                  <a
                                    href={m.url}
                                    target="_blank"
                                    rel="noopener"
                                    class="text-sm text-secondary break-all hover:underline"
                                  >
                                    {m.url}
                                  </a>
                                  {#if isVideo(m) && isYouTube(m.url)}
                                    <p class="text-[10px] text-on-surface-variant mt-1">YouTube</p>
                                  {/if}
                                  <button
                                    type="button"
                                    class="mt-2 text-error text-sm font-medium"
                                    disabled={mediaBusy}
                                    on:click={() => deleteMedia(p, m)}
                                  >
                                    Remover
                                  </button>
                                </div>
                              </li>
                            {/each}
                          </ul>
                        {/if}

                        <!-- add photo -->
                        <div class="grid sm:grid-cols-2 gap-3">
                          <div class="ahq-card p-3 border border-dashed border-outline-variant">
                            <p class="ahq-label mb-2">Adicionar foto</p>
                            <label class="ahq-btn-ghost h-10 px-3 text-sm inline-flex items-center cursor-pointer">
                              <span class="material-symbols-outlined text-[18px] mr-1">upload</span>
                              Enviar arquivo
                              <input
                                type="file"
                                accept="image/jpeg,image/png,image/webp,image/gif"
                                class="hidden"
                                bind:this={photoInput}
                                on:change={(e) => onPhotoFile(p, e)}
                                disabled={mediaBusy}
                              />
                            </label>
                            <p class="text-[10px] text-on-surface-variant mt-2">JPG, PNG, WebP ou GIF</p>
                            <div class="mt-2 flex gap-2">
                              <input
                                type="url"
                                class="ahq-input h-9 text-sm flex-1"
                                placeholder="ou URL /static/… ou https://…"
                                id={`photo-url-${p.id}`}
                              />
                              <button
                                type="button"
                                class="ahq-btn-ghost h-9 px-3 text-sm"
                                disabled={mediaBusy}
                                on:click={() => {
                                  const el = document.getElementById(`photo-url-${p.id}`)
                                  addPhotoURL(p, el?.value)
                                  if (el) el.value = ''
                                }}
                              >
                                URL
                              </button>
                            </div>
                          </div>

                          <div class="ahq-card p-3 border border-dashed border-outline-variant">
                            <p class="ahq-label mb-2">Adicionar vídeo</p>
                            <input
                              type="url"
                              class="ahq-input h-9 text-sm w-full mb-2"
                              placeholder="https://youtube.com/… ou link .mp4"
                              bind:value={videoURL}
                              disabled={mediaBusy}
                            />
                            <button
                              type="button"
                              class="ahq-btn-primary h-9 px-4 text-sm"
                              disabled={mediaBusy || !videoURL.trim()}
                              on:click={() => addVideo(p)}
                            >
                              Adicionar vídeo
                            </button>
                            <p class="text-[10px] text-on-surface-variant mt-2">
                              YouTube, Vimeo ou URL direta (https)
                            </p>
                          </div>
                        </div>
                      </div>
                    </td>
                  </tr>
                {/if}
              {/if}
            {:else}
              <tr>
                <td colspan="6" class="px-3 py-10 text-center text-on-surface-variant">
                  Nenhum produto com essa busca.
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</AppShell>
