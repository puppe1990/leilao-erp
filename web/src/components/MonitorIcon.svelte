<script>
  /** Inline monitor glyph (Lucide, ISC). Uses currentColor so parent can set green. */
  export let size = 20
  export let className = ''
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="2"
  stroke-linecap="round"
  stroke-linejoin="round"
  class="monitor-icon {className}"
  aria-hidden="true"
  focusable="false"
>
  <!-- Lucide Icons — monitor — https://lucide.dev/icons/monitor -->
  <rect width="20" height="14" x="2" y="3" rx="2" />
  <line x1="8" x2="16" y1="21" y2="21" />
  <line x1="12" x2="12" y1="17" y2="21" />
</svg>

<style>
  .monitor-icon {
    display: block;
    flex-shrink: 0;
  }
</style>
