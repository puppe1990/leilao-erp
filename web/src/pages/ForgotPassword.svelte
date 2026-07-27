<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  export let errors = {}
  let form = useForm({ email: '' })
  function submit() {
    form.post('/forgot-password')
  }
</script>

<div class="max-w-sm mx-auto mt-10 p-6 border rounded">
  <h1 class="text-xl font-semibold text-stone-800 mb-4">Esqueci a senha</h1>
  <p class="text-sm text-stone-600 mb-4">
    Informe seu e-mail e enviaremos um link para redefinir a senha.
  </p>
  <form on:submit|preventDefault={submit} class="space-y-3">
    <div>
      <label class="block text-sm text-stone-600 mb-1" for="email">E-mail</label>
      <input
        id="email"
        type="email"
        bind:value={form.email}
        class="block w-full p-2 border rounded"
        autocomplete="username"
      />
      {#if errors.email}<p class="text-red-600 text-xs mt-1">{errors.email}</p>{/if}
    </div>
    <button
      type="submit"
      class="w-full bg-stone-800 text-white px-4 py-2 rounded"
      disabled={form.processing}
    >
      Enviar link
    </button>
  </form>
  <p class="mt-4 text-sm text-center">
    <a href="/login" use:inertia class="underline text-stone-600">Voltar ao login</a>
  </p>
</div>
