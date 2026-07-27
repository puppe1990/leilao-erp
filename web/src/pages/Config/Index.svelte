<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let email = ''
  export let companyName = 'AuctionHQ'
  export let companyForm = ''
  export let errors = {}
  export let flash = ''
  export let site = {}

  let company = useForm({
    company_name: companyForm || '',
  })

  let password = useForm({
    current_password: '',
    new_password: '',
    new_password_confirmation: '',
  })

  function saveCompany() {
    company.post('/config/company')
  }

  function savePassword() {
    password.post('/config/password', {
      onSuccess: () => {
        password.reset('current_password', 'new_password', 'new_password_confirmation')
      },
    })
  }
</script>

<AppShell active="config" title={companyName}>
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Configurações</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Empresa e segurança da conta.</p>
  </div>

  {#if flash}
    <div
      class="mb-4 ahq-card p-3 border-tertiary/30 bg-tertiary-fixed/15 text-on-tertiary-container text-sm font-medium"
    >
      {flash}
    </div>
  {/if}

  <!-- Company -->
  <section class="ahq-card p-5 mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-1">Empresa</h2>
    <p class="text-on-surface-variant text-sm mb-4">Nome exibido no topo do app.</p>

    <form on:submit|preventDefault={saveCompany} class="space-y-4">
      <div>
        <label class="ahq-label block mb-1.5" for="company_name">Nome da empresa</label>
        <input
          id="company_name"
          type="text"
          bind:value={company.company_name}
          class="ahq-input"
          maxlength="80"
          placeholder="Ex: Puppe Leilões"
        />
        {#if errors.company_name}
          <p class="text-error text-sm mt-1">{errors.company_name}</p>
        {/if}
      </div>
      <button type="submit" class="ahq-btn-primary" disabled={company.processing}>
        Salvar nome
      </button>
    </form>
  </section>

  <!-- Account / password -->
  <section class="ahq-card p-5">
    <h2 class="font-headline-md text-headline-md text-primary mb-1">Conta</h2>
    <p class="text-on-surface-variant text-sm mb-4">
      E-mail: <span class="font-mono text-primary">{email}</span>
    </p>

    <form on:submit|preventDefault={savePassword} class="space-y-4">
      <div>
        <label class="ahq-label block mb-1.5" for="current_password">Senha atual</label>
        <input
          id="current_password"
          type="password"
          bind:value={password.current_password}
          class="ahq-input"
          autocomplete="current-password"
        />
        {#if errors.current_password}
          <p class="text-error text-sm mt-1">{errors.current_password}</p>
        {/if}
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="new_password">Nova senha</label>
        <input
          id="new_password"
          type="password"
          bind:value={password.new_password}
          class="ahq-input"
          autocomplete="new-password"
        />
        {#if errors.new_password}
          <p class="text-error text-sm mt-1">{errors.new_password}</p>
        {/if}
        <p class="text-[10px] text-on-surface-variant mt-1">Mínimo 8 caracteres.</p>
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="new_password_confirmation">Confirmar nova senha</label>
        <input
          id="new_password_confirmation"
          type="password"
          bind:value={password.new_password_confirmation}
          class="ahq-input"
          autocomplete="new-password"
        />
        {#if errors.new_password_confirmation}
          <p class="text-error text-sm mt-1">{errors.new_password_confirmation}</p>
        {/if}
      </div>
      <button type="submit" class="ahq-btn-dark" disabled={password.processing}>
        Alterar senha
      </button>
    </form>
  </section>
</AppShell>
