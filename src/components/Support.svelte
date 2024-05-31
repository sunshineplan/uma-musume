<script lang="ts">
  import Image from "./Image.svelte";
  import { support } from "../stores";

  //https://gametora.com/umamusume
  const type: { [key: string]: Support["type"] } = {
    speed: "スピ",
    stamina: "スタ",
    power: "パワ",
    guts: "根性",
    intelligence: "賢さ",
    friend: "友人",
    group: "グル",
  };

  const rare: Support["rare"][] = ["SSR", "SR", "R"];
</script>

<ul class="type">
  {#each Object.entries(type) as [key, value] (key)}
    <li>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <span
        class:checked={$support.type == value}
        on:click={() => {
          const div = document.getElementById("supports");
          if (div) div.scrollTop = 0;
          if ($support.type == value) $support.type = undefined;
          else $support.type = value;
        }}
      >
        <Image id={key} alt={key} type="type" />
      </span>
    </li>
  {/each}
</ul>
<ul class="rare">
  {#each rare as r (r)}
    <li>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <span
        class:checked={$support.rare == r}
        on:click={() => {
          const div = document.getElementById("supports");
          if (div) div.scrollTop = 0;
          if ($support.rare == r) $support.rare = undefined;
          else $support.rare = r;
        }}
      >
        {#if r}
          <Image id={r.toLowerCase()} alt={r} type="rare" />
        {/if}
      </span>
    </li>
  {/each}
</ul>

<style>
  ul {
    display: block;
    border: solid 1px #ced4da;
    margin: auto;
    width: max-content;
    margin-bottom: 10px;
    padding: 0;
    font-size: 0;
    text-align: center;
  }

  li {
    display: inline-block;
  }

  .checked {
    transform: none;
    filter: none;
    box-shadow: inset 3px 3px 4px rgba(0, 0, 0, 0.2);
  }

  span {
    display: block;
    margin: 0;
    padding: 5px 7px;
    cursor: pointer;
    transform: translate(-2px, -2px);
    filter: drop-shadow(2px 2px 3px rgba(0, 0, 0, 0.2));
    transition: all 0.1s;
  }

  .type {
    border-radius: 5px;
  }

  .rare {
    border-radius: 15px;
  }

  .type > li:first-child span {
    border-radius: 4px 0 0 4px;
  }

  .type > li:last-child span {
    border-radius: 0 4px 4px 0;
  }

  .rare > li:first-child span {
    border-radius: 14px 0 0 14px;
  }

  .rare > li:last-child span {
    border-radius: 0 14px 14px 0;
  }
</style>
