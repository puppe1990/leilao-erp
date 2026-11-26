<script>
  /** 3×3 pixel grid; bottom-right cell is the PIX period. */
  export let size = 40
  export let className = ''
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 64 64"
  class="pixeletro-mark {className}"
  aria-hidden="true"
  focusable="false"
>
  <rect class="cell" x="10" y="10" width="12" height="12" rx="2.5" />
  <rect class="cell" x="26" y="10" width="12" height="12" rx="2.5" />
  <rect class="cell" x="42" y="10" width="12" height="12" rx="2.5" />
  <rect class="cell" x="10" y="26" width="12" height="12" rx="2.5" />
  <rect class="cell" x="26" y="26" width="12" height="12" rx="2.5" />
  <rect class="cell" x="42" y="26" width="12" height="12" rx="2.5" />
  <rect class="cell" x="10" y="42" width="12" height="12" rx="2.5" />
  <rect class="cell" x="26" y="42" width="12" height="12" rx="2.5" />
  <rect class="cell-pix" x="42" y="42" width="12" height="12" rx="2.5" />
</svg>

<style>
  .pixeletro-mark {
    display: block;
    flex-shrink: 0;
  }
  .cell {
    fill: none;
    stroke: currentColor;
    stroke-width: 2.4;
  }
  .cell-pix {
    fill: var(--shop-green, #25d366);
    stroke: none;
  }
</style>
