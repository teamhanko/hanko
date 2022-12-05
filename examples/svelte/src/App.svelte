<script lang="ts">
  import { Route, Router } from "svelte-navigator";
  import Login from "./lib/Login.svelte";
  import Todo from "./lib/Todo.svelte";
</script>

<Router>
<main>
  <Route path="/">
    <Login/>
  </Route>

  <Route path="todo">
    <Todo/>
  </Route>
</main>
</Router>
