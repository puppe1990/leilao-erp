<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import { onMount } from 'svelte'
  import PasswordInput from '@/components/PasswordInput.svelte'
  import {
    applyShopThemeToDocument,
    getShopTheme,
    setShopTheme,
  } from '@/lib/shopTheme.js'

  export let errors = {}
  export let token = ''
  let form = useForm({ token, password: '', password_confirmation: '' })
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
    form.post('/reset-password')
  }
</script>

<svelte:head>
  <title>Redefinir senha · Admin</title>
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
      <h1 class="font-headline-lg text-headline-lg text-on-surface mb-1">Redefinir senha</h1>
      <p class="text-on-surface-variant text-body-md mb-6">Escolha uma nova senha para sua conta.</p>
      {#if errors.token}<p class="text-error text-sm mb-3">{errors.token}</p>{/if}
      <form on:submit|preventDefault={submit} class="ahq-card p-6 space-y-4 shadow-float">
        <input type="hidden" bind:value={form.token} />
        <div>
          <label class="ahq-label block mb-1.5" for="password">Nova senha</label>
          <PasswordInput id="password" bind:value={form.password} autocomplete="new-password" />
          {#if errors.password}<p class="text-error text-sm mt-1">{errors.password}</p>{/if}
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="password_confirmation">Confirmar senha</label>
          <PasswordInput
            id="password_confirmation"
            bind:value={form.password_confirmation}
            autocomplete="new-password"
          />
          {#if errors.password_confirmation}
            <p class="text-error text-sm mt-1">{errors.password_confirmation}</p>
          {/if}
        </div>
        <button type="submit" class="ahq-btn-primary w-full" disabled={form.processing}>
          Salvar senha
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
