<script>
  import Image from "./Image.svelte";
  import { support } from "../stores";

  //https://gametora.com/umamusume
  const type = {
    speed: "スピ",
    stamina: "スタ",
    power: "パワ",
    guts: "根性",
    intelligence: "賢さ",
    friend: "友人",
    group: "グル",
  };

  const rare = ["ssr", "sr", "r"];
</script>

<ul class="type">
  {#each Object.entries(type) as [key, value] (key)}
    <li>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
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
      <span
        class:checked={$support.rare == r.toUpperCase()}
        on:click={() => {
          const div = document.getElementById("supports");
          if (div) div.scrollTop = 0;
          if ($support.rare == r.toUpperCase()) $support.rare = undefined;
          else $support.rare = r.toUpperCase();
        }}
      >
        <Image id={r} alt={r} type="rare" />
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
    background: linear-gradient(to bottom, #6c757d, #5c636a);
  }

  span {
    display: block;
    margin: 0;
    padding: 5px 7px;
    cursor: pointer;
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
