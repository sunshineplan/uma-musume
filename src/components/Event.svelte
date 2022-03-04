<script lang="ts">
  import { events, filter, query, showFilter } from "../stores";
  import type { Event } from "../stores";

  let results: Event[] = [];

  $: $filter, search();

  const match = (value: string, event: Event) => {
    if (
      event.c.includes(value) ||
      event.e.includes(value) ||
      event.k.includes(value)
    )
      return true;
    let matched = false;
    event.o.forEach((option) => {
      if (option.b.includes(value)) matched = true;
    });
    return matched;
  };

  const search = () => {
    const div = document.getElementById("results");
    if (div) div.scrollTop = 0;
    if (!$query) {
      if (!$filter.name) results = [];
      else results = $events;
    } else if ($query == "*") results = $events;
    else if (!$filter.name && $query.length == 1) results = [];
    else {
      let r: Event[] = [];
      $events.forEach((event) => {
        if (match($query, event)) r.push(event);
      });
      results = r;
    }
  };

  const addlink = (option: { b: string; g: string; s: object }) => {
    let result = option.g;
    if (option.s)
      Object.entries(option.s).forEach(([key, value]) => {
        result = result.replaceAll(
          `『${key}』`,
          `<a href="https://gamewith.jp/uma-musume/article/show/${value}" target="_blank">『${key}』</a>`
        );
      });
    return result;
  };

  const reset = () => {
    $query = "";
    results = [];
  };
</script>

<div class="content" on:mousedown={showFilter.off}>
  <div class="input-group">
    <input
      class="form-control"
      type="search"
      placeholder="ウマ娘名、イベント、選択肢テキスト、*"
      bind:value={$query}
      on:keydown={(e) => {
        if (e.key === "Escape") reset();
      }}
      on:input={search}
    />
  </div>
  <div class="message">
    <span>{`※${$filter.name ? 1 : 2}文字以上入力すると検索を開始します`}</span>
    <span style="float:right">{`(${results.length}/${$events.length})`}</span>
  </div>
  {#if results.length}
    <div id="results">
      {#each results as result}
        <table class="table table-bordered">
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
                    <img
                      src={result.i
                        ? `https://cdn.jsdelivr.net/gh/sunshineplan/uma-musume@gh-pages/image/${result.i}`
                        : ""}
                      alt={result.c}
                      on:click={() => {
                        if (result.a)
                          window.open(
                            `https://gamewith.jp/uma-musume/article/show/${result.a}`
                          );
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
            {#each result.o as option (option.b)}
              <tr>
                <td style="vertical-align:middle">
                  {option.b}
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
  {/if}
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

  img {
    width: 72px;
    max-width: 72px;
    min-height: 72px;
    margin-right: 10px;
    cursor: pointer;
  }

  @media (max-width: 767px) {
    .content {
      left: 0;
      width: 100%;
    }
  }
</style>
