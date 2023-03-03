<script lang="ts">
    import { onMount } from "svelte";
    import { useNavigate } from "svelte-navigator";
    import { TodoClient } from "./TodoClient";
    import { register } from "@teamhanko/hanko-elements";

    const hankoAPI = import.meta.env.VITE_HANKO_API;
    const todoAPI = import.meta.env.VITE_TODO_API;
    const todoClient = new TodoClient(todoAPI);

    const navigate = useNavigate();

    let error: Error | null = null;

    onMount(async () => {
        register({ shadow: true }).catch((e) => error = e);
    });

    const logout = () => {
        todoClient
            .logout()
            .then(() => {
                navigate("/");
                return;
            })
            .catch((e) => error = e);
    }

    const todos = () => {
        navigate("/todo");
    }
</script>

<nav class="nav">
    <button class="button" on:click|preventDefault={logout}>Logout</button>
    <button class="button" disabled>Profile</button>
    <button class="button" on:click|preventDefault={todos}>Todos</button>
</nav>
<div class="content">
    <h1 class="headline">Profile</h1>
    {#if error}
        <div class="error">{ error?.message }</div>
    {/if}
    <hanko-profile api={hankoAPI}/>
</div>
