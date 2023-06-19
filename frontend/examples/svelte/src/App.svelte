<script lang="ts">
  import { Route, Router } from "svelte-navigator";
  import Login from "./lib/Login.svelte";
  import Todo from "./lib/Todo.svelte";
  import Profile from "./lib/Profile.svelte";
  import { onMount } from "svelte";
  import { register } from "@teamhanko/hanko-elements";

  const hankoAPI = import.meta.env.VITE_HANKO_API;

  onMount(() => {
    register(hankoAPI);
  });
</script>

<Router>
<main>
  <Route path="/">
    <Login/>
  </Route>
  <Route path="todo">
    <Todo/>
  </Route>
  <Route path="profile">
    <Profile/>
  </Route>
</main>
</Router>

<style>
  :global(.nav) {
    width: 100%;
    padding: 10px;
    opacity: 0.9;
  }

  :global(.button) {
    font-size: 1rem;
    border: none;
    background: none;
    cursor: pointer;
  }

  :global(.button):disabled {
    color: grey!important;
    cursor: default;
    text-decoration: none!important;
  }

  :global(.nav .button):hover {
    text-decoration: underline;
  }

  :global(.nav .button) {
    color: white;
    float: right;
  }

  :global(.content) {
    padding: 24px;
    border-radius: 17px;
    color: black;
    background-color: white;
    width: 100%;
    max-width: 500px;
    min-width: 330px;
    margin: 10vh auto;
  }


  :global(.headline) {
    text-align: center;
    margin-top: 0;
  }

  :global(.error) {
    color: red;
    padding: 0 0 10px;
  }

</style>
