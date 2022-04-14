<script lang="ts">
  import Support from "./Support.svelte";
  import { characters, supports, filter, query, showFilter } from "../stores";

  let type = "character";
</script>

<span class="toggle" class:on={$showFilter} on:click={showFilter.switch}>
  <svg
    viewBox="0 0 16 16"
    width="32"
    height="32"
    fill={$filter.name ? "#007bff" : "#6c757d"}
  >
    <path
      d="M6 10.5a.5.5 0 0 1 .5-.5h3a.5.5 0 0 1 0 1h-3a.5.5 0 0 1-.5-.5zm-2-3a.5.5 0 0 1 .5-.5h7a.5.5 0 0 1 0 1h-7a.5.5 0 0 1-.5-.5zm-2-3a.5.5 0 0 1 .5-.5h11a.5.5 0 0 1 0 1h-11a.5.5 0 0 1-.5-.5z"
    />
  </svg>
</span>
<div class="filter" class:open={$showFilter}>
  <div class="info">
    <div>
      <h5>フィルタ:</h5>
      <div class="button" class:hidden={$filter.name == ""}>
        <button
          class="btn-close"
          on:click={() => {
            if (window.innerWidth <= 767) showFilter.off();
            $filter = { type: "character", name: "", image: "" };
          }}
        />
      </div>
    </div>
    <div class="display">
      {#if !$filter.name}
        <h5 style="color:gray">なし</h5>
      {:else if $filter.type == "character"}
        <img src="https://cdn.jsdelivr.net/gh/sunshineplan/uma-musume@gh-pages/image/{$filter.image}" alt={$filter.name} />
        <span>{$filter.name}</span>
      {:else if $filter.type == "support"}
        <img src="https://cdn.jsdelivr.net/gh/sunshineplan/uma-musume@gh-pages/image/{$filter.image}" alt={$filter.name} />
        <div style="display:grid">
          <span>{$filter.name}</span>
          <span>{$filter.rare}</span>
        </div>
      {/if}
    </div>
  </div>
  <ul class="nav nav-tabs">
    <li class="nav-item">
      <span
        class="nav-link"
        class:active={type == "character"}
        on:click={() => (type = "character")}
      >
        キャラ
      </span>
    </li>
    <li class="nav-item">
      <span
        class="nav-link"
        class:active={type == "support"}
        on:click={() => (type = "support")}
      >
        サポート
      </span>
    </li>
  </ul>
  <div class="items">
    {#if type == "character"}
      <div class="characters list">
        {#each characters as i (i.name)}
          <li>
            <img
              class:selected={$filter.type == "character" &&
                $filter.name == i.name}
              src="https://cdn.jsdelivr.net/gh/sunshineplan/uma-musume@gh-pages/image/{i.image}"
              alt={i.name}
              title={i.name}
              style="height:72px"
              on:click={() => {
                if (window.innerWidth <= 767) showFilter.off();
                if (
                  $filter.name &&
                  ($filter.type == "support" ||
                    ($filter.type == "character" && $filter.name != i.name))
                )
                  $query = "";
                if ($filter.type == "character" && $filter.name == i.name)
                  $filter = { type: "character", name: "", image: "" };
                else
                  $filter = { type: "character", name: i.name, image: i.image };
              }}
            />
          </li>
        {/each}
      </div>
    {:else}
      <Support />
      <div id="supports" class="list">
        {#each $supports as i (i.name + i.rare)}
          <li>
            <img
              class:selected={$filter.type == "support" &&
                $filter.name == i.name &&
                $filter.rare == i.rare}
              src="https://cdn.jsdelivr.net/gh/sunshineplan/uma-musume@gh-pages/image/{i.image}"
              alt={i.name}
              title={i.name}
              style="min-height:96px"
              on:click={() => {
                if (window.innerWidth <= 767) showFilter.off();
                if (
                  $filter.name &&
                  ($filter.type == "character" ||
                    ($filter.type == "support" &&
                      $filter.name != i.name &&
                      $filter.rare != i.rare))
                )
                  $query = "";
                if (
                  $filter.type == "support" &&
                  $filter.name == i.name &&
                  $filter.rare == i.rare
                )
                  $filter = { type: "support", name: "", image: "" };
                else
                  $filter = {
                    type: "support",
                    name: i.name,
                    rare: i.rare,
                    image: i.image,
                  };
              }}
            />
          </li>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .toggle {
    position: fixed;
    z-index: 100;
    top: 0;
    padding: 20px;
    display: none;
  }

  .on,
  .toggle:hover {
    background-color: rgb(232, 232, 232);
  }

  .filter {
    position: fixed;
    z-index: 1;
    top: var(--nav);
    height: calc(100% - var(--nav));
    width: var(--filter);
    border-right: 1px solid #e9ecef;
    background-color: white;
  }

  .info {
    padding: 5px;
  }

  h5 {
    display: inline-flex;
    cursor: default;
  }

  .button {
    background-color: red;
    display: inline-flex;
    border-radius: 5px;
  }

  .button.hidden {
    visibility: hidden;
  }

  .btn-close {
    box-sizing: content-box;
    width: 1em;
    height: 1em;
    padding: 0.25em 0.25em;
    background: transparent
      url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16' fill='%23000'%3e%3cpath d='M.293.293a1 1 0 011.414 0L8 6.586 14.293.293a1 1 0 111.414 1.414L9.414 8l6.293 6.293a1 1 0 01-1.414 1.414L8 9.414l-6.293 6.293a1 1 0 01-1.414-1.414L6.586 8 .293 1.707a1 1 0 010-1.414z'/%3e%3c/svg%3e")
      center/1em auto no-repeat;
    border: 0;
    border-radius: 0.25rem;
    opacity: 0.75;
    filter: invert(1) grayscale(100%) brightness(200%);
  }

  .display {
    height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .nav {
    justify-content: center;
    margin: 10px 0;
  }

  .nav-link {
    cursor: default;
    color: rgba(0, 0, 0, 0.55);
  }
  .nav-link:hover {
    color: rgba(0, 0, 0, 0.7);
  }
  .nav-link.active {
    color: rgba(0, 0, 0, 0.9);
  }

  .items {
    height: calc(100% - 206px);
    text-align: center;
  }

  .list {
    overflow: overlay;
  }

  .characters {
    height: 100%;
  }

  #supports {
    height: calc(100% - 90px);
  }

  .list:hover::-webkit-scrollbar-thumb {
    background: #6c757d;
  }

  ::-webkit-scrollbar {
    width: 3px;
  }

  li {
    display: inline-block;
  }

  img {
    width: 72px;
    max-width: 72px;
    border-width: 2px;
    border-style: solid;
    border-radius: 8px;
    border-color: transparent;
  }

  .selected {
    border-color: #007bff;
  }

  @media (max-width: 767px) {
    .filter {
      width: 100%;
      border-right: 0;
      left: -100%;
      transition: left 0.3s ease-in-out;
    }

    .toggle {
      display: block;
    }

    .open {
      left: 0;
    }
  }
</style>
