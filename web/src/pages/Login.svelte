<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  export let errors = {}
  export let site = {}
  let form = useForm({ email: '', password: '' })
  function submit() {
    form.post('/login')
  }
</script>

<div class="max-w-sm mx-auto mt-10 p-6 border rounded">
  <h1 class="text-xl font-semibold text-stone-800 mb-4">Entrar</h1>
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
      {#if errors.email}<div class="text-red-600 text-xs mt-1">{errors.email}</div>{/if}
    </div>
    <div>
      <label class="block text-sm text-stone-600 mb-1" for="password">Senha</label>
      <input
        id="password"
        type="password"
        bind:value={form.password}
        class="block w-full p-2 border rounded"
        autocomplete="current-password"
      />
    </div>
    <button type="submit" class="mt-2 w-full bg-stone-800 text-white px-4 py-2 rounded" disabled={form.processing}>
      Entrar
    </button>
  </form>
  <p class="mt-4 text-sm text-center">
    <a href="/forgot-password" use:inertia class="underline text-stone-600">Esqueci a senha</a>
  </p>
</div>
