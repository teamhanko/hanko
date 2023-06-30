<script lang="ts">
    import { useNavigate } from "svelte-navigator";
    import { Hanko } from "@teamhanko/hanko-elements";
    import SessionExpiredModal from "./SessionExpiredModal.svelte";
    import { onMount } from "svelte";

    const hankoAPI = import.meta.env.VITE_HANKO_API;
    const hankoClient = new Hanko(hankoAPI);

    const navigate = useNavigate();

    let showModal = false;
    let error: Error | null = null;

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

    onMount(() => {
        if (!hankoClient.session.isValid()) {
            redirectToLogin();
        }
    });
</script>

<SessionExpiredModal bind:showModal></SessionExpiredModal>
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
    <hanko-profile on:onUserLoggedOut={redirectToLogin}></hanko-profile>
</div>
