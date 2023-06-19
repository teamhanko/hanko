<script lang="ts">
  import { useNavigate } from "svelte-navigator";

  const navigate = useNavigate();

  let dialog: HTMLDialogElement; // HTMLDialogElement

  const redirectToLogin = () => {
    navigate("/");
  }

  export const openSessionExpiredModal = () => {
    dialog.showModal();
  }
</script>

<hanko-events on:onSessionExpired={openSessionExpiredModal}></hanko-events>
<dialog bind:this={dialog}>
  Please login again.<br /><br />
  <button on:click={redirectToLogin}>Login</button>
</dialog>
