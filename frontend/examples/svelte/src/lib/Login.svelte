<script lang="ts">
  import { useNavigate } from "svelte-navigator";
  import { Hanko } from '@teamhanko/hanko-elements';

  const hankoAPI = import.meta.env.VITE_HANKO_API;
  const hankoClient = new Hanko(hankoAPI);

  const navigate = useNavigate();

  const redirectToTodos = () => {
    navigate('/todo');
  };

  let error: Error | null = null;
</script>

<div class="content">
  {#if error}
    <div class="error">{ error?.message }</div>
  {/if}
  <hanko-auth on:onSessionCreated={redirectToTodos}></hanko-auth>
</div>

