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
  export let site = {}

  let form = useForm({ email: '', password: '' })
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
    form.post('/login')
  }
</script>

<svelte:head>
  <title>Entrar · Admin</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col">
  <header
    class="h-16 px-container-margin flex items-center justify-between border-b border-outline-variant bg-surface-container-lowest"
  >
    <div class="flex items-center gap-3">
      <div class="ahq-brand-mark">
        <span class="material-symbols-outlined text-[20px]">storefront</span>
      </div>
      <div>
        <div class="font-headline-md text-headline-md font-black text-on-surface">
          Admin<span class="text-secondary">.</span>
        </div>
        <p class="text-[10px] font-bold uppercase tracking-wider text-on-surface-variant">
          ERP · leilões
        </p>
      </div>
    </div>
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
      <div
        class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-secondary/15 text-secondary text-[11px] font-bold uppercase tracking-wider mb-4"
      >
        <span class="material-symbols-outlined text-[16px]">lock</span>
        Acesso restrito
      </div>
      <h1 class="font-headline-lg text-headline-lg text-on-surface mb-1">Entrar</h1>
      <p class="text-on-surface-variant text-body-md mb-6">
        Acesse o ERP de gestão de leilões e estoque.
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
          {#if errors.email}<div class="text-error text-xs mt-1">{errors.email}</div>{/if}
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="password">Senha</label>
          <PasswordInput id="password" bind:value={form.password} autocomplete="current-password" />
        </div>
        <button type="submit" class="ahq-btn-primary w-full" disabled={form.processing}>
          Entrar
        </button>
      </form>

      <p class="mt-4 text-center text-sm">
        <a href="/forgot-password" use:inertia class="text-secondary font-semibold hover:underline"
          >Esqueci a senha</a
        >
      </p>
      <p class="mt-6 text-center text-xs text-on-surface-variant">
        <a href="/" class="hover:text-secondary transition-colors">← Voltar ao catálogo</a>
      </p>
    </div>
  </main>
</div>
