<script lang="ts">
    import { onDestroy, onMount } from "svelte";
    import { useNavigate } from "svelte-navigator";
    import { createHankoClient, register } from "@teamhanko/hanko-elements";

    const hankoAPI = import.meta.env.VITE_HANKO_API;
    const hankoClient = createHankoClient(hankoAPI);

    const navigate = useNavigate();

    let error: Error | null = null;

    onMount(() => {
        register(hankoAPI).catch((e) => error = e);
        hankoClient.onSessionRemoved(redirectToLogin);
    });

    onDestroy(() => {
        hankoClient.removeEventListeners();
    });

    const logout = () => {
        hankoClient.user
            .logout()
            .catch((e) => error = e);
    }

    const redirectToLogin = () => {
        navigate("/");
    }

    const redirectToTodos = () => {
        navigate("/todo");
    }
</script>

<nav class="nav">
    <button class="button" on:click|preventDefault={logout}>Logout</button>
    <button class="button" disabled>Profile</button>
    <button class="button" on:click|preventDefault={redirectToTodos}>Todos</button>
</nav>
<div class="content">
    <h1 class="headline">Profile</h1>
    {#if error}
        <div class="error">{ error?.message }</div>
    {/if}
    <hanko-profile />
</div>
