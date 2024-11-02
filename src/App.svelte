<script lang="ts">
  import Event from "./components/Event.svelte";
  import Filter from "./components/Filter.svelte";
</script>

<header class="navbar navbar-expand flex-column flex-md-row">
  <a class="navbar-brand text-primary m-0 mr-md-3" href="/">ウマ娘</a>
</header>
<Filter />
<Event />

<style>
  .navbar {
    user-select: none;
    height: var(--nav);
    justify-content: space-between;
    letter-spacing: 0.3px;
    border-bottom: 5px solid #f2f2f2;
    padding: 0.5rem 1rem;
  }

  .navbar-brand {
    font-size: 24px;
    padding-left: 30px;
  }

  @media (max-width: 767px) {
    .navbar {
      border-color: transparent;
    }

    .navbar-brand {
      padding-left: 0;
    }
  }
</style>
