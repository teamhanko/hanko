<script lang="ts">
    import { useNavigate } from "svelte-navigator";
    import { Hanko } from "@teamhanko/hanko-elements";
    import SessionExpiredModal from "./SessionExpiredModal.svelte";
    import { onMount } from "svelte";

    const hankoAPI = import.meta.env.VITE_HANKO_API;
    const hankoClient = new Hanko(hankoAPI);

    const navigate = useNavigate();

    let showModal = $state(false);
    let error: Error | null = $state(null);

    const logout = (event) => {
        event.preventDefault()
        hankoClient.logout()
            .catch((e) => error = e);
    }

    const redirectToLogin = () => {
        navigate("/");
    }

    const redirectToTodos = (event) => {
        event.preventDefault()
        navigate("/todo");
    }

    onMount(async () => {
        const {is_valid} = await hankoClient.validateSession();
        if (!is_valid) {
            redirectToLogin();
        }
    });
</script>

<SessionExpiredModal bind:showModal></SessionExpiredModal>
<nav class="nav">
    <button class="button" onclick={logout}>Logout</button>
    <button class="button" disabled>Profile</button>
    <button class="button" onclick={redirectToTodos}>Todos</button>
</nav>
<div class="content">
    <h1 class="headline">Profile</h1>
    {#if error}
        <div class="error">{ error?.message }</div>
    {/if}
    <hanko-profile ononUserLoggedOut={redirectToLogin}></hanko-profile>
</div>
