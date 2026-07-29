<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import {
    applyShopThemeToDocument,
    getShopTheme,
    setShopTheme,
  } from '@/lib/shopTheme.js'

  export let errors = {}
  let form = useForm({ email: '' })
  /** @type {'dark'|'light'} */
  let theme = 'dark'

  onMount(() => {
    theme = getShopTheme()
    applyShopThemeToDocument(theme)
  })

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    setShopTheme(theme)
  }

  function submit() {
    form.post('/forgot-password')
  }
</script>

<svelte:head>
  <title>Esqueci a senha · Admin</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col">
  <header
    class="h-16 px-container-margin flex items-center justify-between border-b border-outline-variant bg-surface-container-lowest"
  >
    <a href="/login" use:inertia class="flex items-center gap-3">
      <div class="ahq-brand-mark">
        <span class="material-symbols-outlined text-[20px]">storefront</span>
      </div>
      <span class="font-headline-md text-headline-md font-black text-on-surface"
        >Admin<span class="text-secondary">.</span></span
      >
    </a>
    <button
      type="button"
      class="ahq-theme-btn"
      on:click={toggleTheme}
      title={theme === 'dark' ? 'Modo claro' : 'Modo escuro'}
      aria-label={theme === 'dark' ? 'Ativar modo claro' : 'Ativar modo escuro'}
    >
      <span class="material-symbols-outlined"
        >{theme === 'dark' ? 'light_mode' : 'dark_mode'}</span
      >
    </button>
  </header>

  <main class="flex-1 flex items-center justify-center px-container-margin py-10">
    <div class="w-full max-w-sm">
      <h1 class="font-headline-lg text-headline-lg text-on-surface mb-1">Esqueci a senha</h1>
      <p class="text-on-surface-variant text-body-md mb-6">
        Informe seu e-mail e enviaremos um link para redefinir a senha.
      </p>
      <form on:submit|preventDefault={submit} class="ahq-card p-6 space-y-4 shadow-float">
        <div>
          <label class="ahq-label block mb-1.5" for="email">E-mail</label>
          <input
            id="email"
            type="email"
            bind:value={form.email}
            class="ahq-input"
            autocomplete="username"
          />
          {#if errors.email}<p class="text-error text-xs mt-1">{errors.email}</p>{/if}
        </div>
        <button type="submit" class="ahq-btn-primary w-full" disabled={form.processing}>
          Enviar link
        </button>
      </form>
      <p class="mt-4 text-sm text-center">
        <a href="/login" use:inertia class="text-secondary font-semibold hover:underline"
          >Voltar ao login</a
        >
      </p>
    </div>
  </main>
</div>
