<script lang="ts">
  import { showFilter } from "../stores";
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<header
  class="navbar navbar-expand flex-column flex-md-row"
  on:mousedown={showFilter.off}
>
  <a class="navbar-brand text-primary m-0 mr-md-3" href="/">ウマ娘</a>
</header>

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
