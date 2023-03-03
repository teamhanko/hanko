<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { useNavigate } from "svelte-navigator";
  import { register } from '@teamhanko/hanko-elements';

  const api = import.meta.env.VITE_HANKO_API;

  const navigate = useNavigate();
  let element;

  const redirectToTodos = () => {
    navigate('/todo');
  };

  let error: Error | null = null;

  onMount(async () => {
    register({ shadow: true }).catch((e) => error = e);
    element?.addEventListener('hankoAuthSuccess', redirectToTodos);
  });

  onDestroy(() => {
    element?.removeEventListener('hankoAuthSuccess', redirectToTodos);
  });
</script>

<div class="content">
  {#if error}
    <div class="error">{ error?.message }</div>
  {/if}
  <hanko-auth bind:this={element} {api}/>
</div>

