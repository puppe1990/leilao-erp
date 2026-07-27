<script>
  import { useForm, inertia } from '@inertiajs/svelte'

  export let errors = {}
  export let site = {}

  let form = useForm({ email: '', password: '' })

  function submit() {
    form.post('/login')
  }
</script>

<div class="min-h-screen bg-background flex flex-col">
  <header class="h-16 px-container-margin flex items-center border-b border-outline-variant bg-surface-container-lowest">
    <div class="flex items-center gap-3">
      <div class="w-9 h-9 rounded-full bg-primary text-on-primary flex items-center justify-center font-bold text-sm">
        AQ
      </div>
      <span class="font-headline-md text-headline-md font-bold text-primary">AuctionHQ</span>
    </div>
  </header>

  <main class="flex-1 flex items-center justify-center px-container-margin py-10">
    <div class="w-full max-w-sm">
      <h1 class="font-headline-lg text-headline-lg text-primary mb-1">Entrar</h1>
      <p class="text-on-surface-variant text-body-md mb-6">Acesse o ERP de gestão de leilões.</p>

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
          <input
            id="password"
            type="password"
            bind:value={form.password}
            class="ahq-input"
            autocomplete="current-password"
          />
        </div>
        <button type="submit" class="ahq-btn-primary w-full" disabled={form.processing}>Entrar</button>
      </form>

      <p class="mt-4 text-center text-sm">
        <a href="/forgot-password" use:inertia class="text-secondary font-medium">Esqueci a senha</a>
      </p>
    </div>
  </main>
</div>
