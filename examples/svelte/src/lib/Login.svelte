<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { useNavigate } from "svelte-navigator";
  import { register, testString } from '@teamhanko/hanko-elements';

  const api = import.meta.env.VITE_HANKO_API;

  const navigate = useNavigate();
  let element;

  const redirectToTodos = () => {
    navigate('/todo');
  };

  onMount(async () => {
    register({ shadow: true }).catch((e) => {
      console.error(e)
    });

    element?.addEventListener('hankoAuthSuccess', redirectToTodos);
  });

  onDestroy(() => {
    element?.removeEventListener('hankoAuthSuccess', redirectToTodos);
  });
</script>

<div class="content">
  <h1>Hello from {testString}</h1>
  <hanko-auth bind:this={element} {api}/>
</div>

<style>
  .content {
      padding: 24px;
      border-radius: 17px;
      color: black;
      background-color: white;
      width: 500px;
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
  }
</style>
