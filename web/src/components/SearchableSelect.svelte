<script>
  /**
   * Single-select combobox with typeahead search.
   * @type {{ value: string, label: string }[]}
   */
  export let options = []
  /** Selected option value (string). */
  export let value = ''
  export let placeholder = 'Selecione…'
  export let searchPlaceholder = 'Buscar…'
  export let id = 'searchable-select'
  export let emptyLabel = 'Nenhuma opção'
  /** Extra classes on the trigger button (defaults to full-width ahq-select). */
  export let buttonClass =
    'ahq-select w-full text-left flex items-center justify-between gap-2 min-h-[2.5rem]'
  /** Show "Limpar seleção" when a value is set. */
  export let allowClear = true
  /** Called when value changes: (value: string) => void */
  export let onChange = undefined

  let open = false
  let query = ''
  let rootEl

  $: selected = options.find((o) => String(o.value) === String(value))
  $: q = query.trim().toLowerCase()
  $: filtered = !q
    ? options
    : options.filter((o) => {
        const hay = `${o.value} ${o.label}`.toLowerCase()
        return hay.includes(q)
      })

  function selectOption(opt) {
    value = String(opt.value)
    open = false
    query = ''
    onChange?.(value)
  }

  function clearSelection() {
    value = ''
    open = false
    query = ''
    onChange?.('')
  }

  function onDocClick(e) {
    if (!open) return
    if (rootEl && !rootEl.contains(e.target)) {
      open = false
      query = ''
    }
  }

  function onKeydown(e) {
    if (e.key === 'Escape') {
      open = false
      query = ''
    }
  }
</script>

<svelte:window on:click={onDocClick} on:keydown={onKeydown} />

<div class="relative" bind:this={rootEl}>
  <button
    type="button"
    {id}
    class={buttonClass}
    aria-haspopup="listbox"
    aria-expanded={open}
    on:click|stopPropagation={() => {
      open = !open
      if (!open) query = ''
    }}
  >
    <span class="truncate {selected ? 'text-primary' : 'text-on-surface-variant'}">
      {selected ? selected.label : placeholder}
    </span>
    <span class="material-symbols-outlined text-[20px] text-on-surface-variant shrink-0">
      {open ? 'expand_less' : 'expand_more'}
    </span>
  </button>

  {#if open}
    <div
      class="absolute z-40 left-0 right-0 mt-1 rounded-lg border border-outline-variant
        bg-surface-container-lowest shadow-float overflow-hidden"
      role="listbox"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown|stopPropagation
    >
      <div class="p-2 border-b border-outline-variant">
        <div class="relative">
          <span
            class="material-symbols-outlined absolute left-2.5 top-1/2 -translate-y-1/2 text-[18px] text-on-surface-variant pointer-events-none"
          >
            search
          </span>
          <input
            type="search"
            class="ahq-input h-9 pl-9 w-full text-sm"
            placeholder={searchPlaceholder}
            bind:value={query}
            on:click|stopPropagation
          />
        </div>
      </div>
      <ul class="max-h-56 overflow-y-auto py-1">
        {#if filtered.length === 0}
          <li class="px-3 py-2 text-sm text-on-surface-variant">{emptyLabel}</li>
        {:else}
          {#each filtered as opt (String(opt.value))}
            <li>
              <button
                type="button"
                class="w-full text-left px-3 py-2 text-sm hover:bg-surface-container-high transition-colors
                  {String(opt.value) === String(value)
                  ? 'bg-secondary-container text-on-secondary-container font-medium'
                  : 'text-primary'}"
                role="option"
                aria-selected={String(opt.value) === String(value)}
                on:click={() => selectOption(opt)}
              >
                {opt.label}
              </button>
            </li>
          {/each}
        {/if}
      </ul>
      {#if allowClear && value !== '' && value != null}
        <div class="border-t border-outline-variant p-1">
          <button
            type="button"
            class="w-full text-sm text-on-surface-variant hover:text-error py-1.5"
            on:click={clearSelection}
          >
            Limpar seleção
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>
