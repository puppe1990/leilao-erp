<script>
  /** Text value (two-way) */
  export let value = ''
  export let id = ''
  export let name = ''
  export let autocomplete = 'current-password'
  /** Extra classes on the <input> */
  export let inputClass = 'ahq-input'
  export let disabled = false
  export let required = false
  export let placeholder = ''

  let show = false

  function toggle() {
    show = !show
  }
</script>

<div class="relative w-full">
  <input
    {id}
    {name}
    type={show ? 'text' : 'password'}
    bind:value
    class="{inputClass} !pr-11"
    {autocomplete}
    {disabled}
    {required}
    {placeholder}
    spellcheck="false"
  />
  <!-- Always on the RIGHT: explicit inset so it never falls to static left if Tailwind is stale -->
  <button
    type="button"
    class="absolute inset-y-0 right-0 z-10 flex w-11 items-center justify-center
      text-on-surface-variant hover:text-primary transition-colors
      focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary rounded-r"
    style="right: 0; left: auto;"
    on:click={toggle}
    aria-label={show ? 'Ocultar senha' : 'Mostrar senha'}
    aria-pressed={show}
  >
    <span class="material-symbols-outlined text-[20px]" aria-hidden="true">
      {show ? 'visibility_off' : 'visibility'}
    </span>
  </button>
</div>
