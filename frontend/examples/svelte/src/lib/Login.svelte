<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { useNavigate } from "svelte-navigator";
  import { createHankoClient, register } from '@teamhanko/hanko-elements';

  const api = import.meta.env.VITE_HANKO_API;
  const hankoClient = createHankoClient(api);

  const navigate = useNavigate();

  const redirectToTodos = () => {
    navigate('/todo');
  };

  let error: Error | null = null;

  onMount(() => {
    register(api).catch((e) => error = e);
    hankoClient.onAuthFlowCompleted(redirectToTodos);
  });

  onDestroy(() => {
    hankoClient.removeEventListeners();
  });
</script>

<div class="content">
  {#if error}
    <div class="error">{ error?.message }</div>
  {/if}
  <hanko-auth />
</div>

