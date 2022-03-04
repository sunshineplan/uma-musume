<script lang="ts">
  import Nav from "./components/Nav.svelte";
  import Filter from "./components/Filter.svelte";
  import Event from "./components/Event.svelte";
</script>

<Nav />
<Filter />
<Event />

<style>
  :global(:root) {
    --nav: 80px;
    --filter: 289px;
  }

  :global(body) {
    font-family: var(--bs-font-monospace);
  }
</style>
