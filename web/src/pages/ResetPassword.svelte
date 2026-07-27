<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  export let errors = {}
  export let token = ''
  let form = useForm({ token, password: '', password_confirmation: '' })
  function submit() {
    $form.post('/reset-password')
  }
</script>

<div class="max-w-sm mx-auto mt-10 p-6 border rounded">
  <h1 class="text-xl font-semibold text-stone-800 mb-4">Redefinir senha</h1>
  {#if errors.token}<p class="text-red-600 text-sm mb-2">{errors.token}</p>{/if}
  <form on:submit|preventDefault={submit} class="space-y-3">
    <input type="hidden" bind:value={$form.token} />
    <div>
      <label class="block text-sm text-stone-600 mb-1" for="password">Nova senha</label>
      <input
        id="password"
        bind:value={$form.password}
        type="password"
        class="block w-full border p-2 rounded"
        autocomplete="new-password"
      />
      {#if errors.password}<p class="text-red-600 text-sm mt-1">{errors.password}</p>{/if}
    </div>
    <div>
      <label class="block text-sm text-stone-600 mb-1" for="password_confirmation">Confirmar senha</label>
      <input
        id="password_confirmation"
        bind:value={$form.password_confirmation}
        type="password"
        class="block w-full border p-2 rounded"
        autocomplete="new-password"
      />
      {#if errors.password_confirmation}
        <p class="text-red-600 text-sm mt-1">{errors.password_confirmation}</p>
      {/if}
    </div>
    <button
      type="submit"
      class="w-full bg-stone-800 text-white px-4 py-2 rounded"
      disabled={$form.processing}
    >
      Salvar senha
    </button>
  </form>
  <p class="mt-4 text-sm text-center">
    <a href="/login" use:inertia class="underline text-stone-600">Voltar ao login</a>
  </p>
</div>
