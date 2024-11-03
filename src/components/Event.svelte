<script lang="ts">
  import { fly } from "svelte/transition";
  import { uma } from "../uma.svelte";
  import Image from "./Image.svelte";

  const max = 100;
  let results = $state<Event[]>([]);
  let resultsDIV: HTMLElement;

  $effect(() => {
    resultsDIV.scrollTop = 0;
    results = [];
    const length = uma.events.length > max ? max : uma.events.length;
    results = uma.events.slice(0, length);
  });

  const addlink = (option: { b: string; g: string; s: object }) => {
    let result = option.g;
    if (option.s)
      Object.entries(option.s).forEach(([key, value]) => {
        result = result.replaceAll(
          `『${key}』`,
          `<a href="${value}" target="_blank">『${key}』</a>`,
        );
      });
    return result;
  };

  const reset = () => {
    uma.query = "";
    results = [];
  };
</script>

<div class="content">
  <div class="input-group">
    <input
      class="form-control"
      type="search"
      placeholder="ウマ娘名、イベント、選択肢テキスト"
      bind:value={uma.query}
      onkeydown={(e) => {
        if (e.key === "Escape") reset();
      }}
    />
  </div>
  <div class="message">
    <span>※文字入力すると検索を開始します</span>
    <span style="float:right">{`(${uma.events.length}/${uma.count})`}</span>
  </div>
  <div id="results" bind:this={resultsDIV}>
    {#each results as result, i (JSON.stringify(result))}
      <table
        class="table table-bordered"
        in:fly={{ x: "100%", delay: i * 70, duration: 400 }}
        out:fly={{ x: "100%", delay: i * 70, duration: 400 }}
      >
        <thead>
          <tr>
            <th colspan="2">
              {result.e}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td colspan="2">
              <div style="display:flex">
                {#if result.a || result.i}
                  <!-- svelte-ignore a11y_click_events_have_key_events -->
                  <Image
                    id={result.i}
                    alt={result.c}
                    type="link"
                    onclick={() => {
                      if (result.a) window.open(result.a);
                    }}
                  />
                {/if}
                <div style="display:grid">
                  {#if result.t == "m"}
                    <span>{result.c}</span>
                    <span>メインシナリオイベント</span>
                  {:else}
                    <span>{result.c}</span>
                    <span>{result.t == "c" ? "ウマ娘" : "サポート"}</span>
                    <span>{result.r}</span>
                  {/if}
                </div>
              </div>
            </td>
          </tr>
          {#each result.o as option (option.b + option.g)}
            <tr>
              <td style="vertical-align:middle">
                {@html option.b}
              </td>
              <td style="white-space:pre-line">
                {@html addlink(option)}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/each}
  </div>
</div>

<style>
  .content {
    position: fixed;
    top: var(--nav);
    left: var(--filter);
    width: calc(100% - var(--filter));
    height: calc(100% - var(--nav));
  }

  .input-group {
    padding-top: 1em;
  }

  .message {
    color: grey;
    font-size: 14px;
    padding-bottom: 1rem !important;
  }

  #results {
    height: calc(100% - 91px);
    overflow-y: overlay;
  }

  table {
    margin-bottom: 1rem !important;
    table-layout: fixed;
  }

  .input-group,
  .message,
  table {
    width: 90%;
    margin: auto;
  }

  tbody {
    border-width: 0 !important;
  }

  @media (max-width: 767px) {
    .content {
      left: 0;
      width: 100%;
    }
  }
</style>
