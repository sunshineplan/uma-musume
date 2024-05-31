<script lang="ts">
  import { fly } from "svelte/transition";
  import Image from "./Image.svelte";
  import { events, filter, query, showFilter } from "../stores";

  let results: Event[] = [];
  let count = 0;
  let current = 0;

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
    let res: Event[] = [];
    if (!$query) {
      if ($filter.name) res = $events;
    } else if ($filter.name ? $query.length > 0 : $query.length > 1) {
      $events.forEach((event) => {
        if (match($query, event)) res.push(event);
      });
    }
    const div = document.getElementById("results");
    if (div) div.scrollTop = 0;
    const interval = 70;
    current++;
    const pid = current;
    const length = results.length > 5 ? 5 : results.length;
    if (length) {
      results = results.slice(0, length);
      const id = setInterval(() => {
        if (results.length && pid == current) {
          results = results.slice(1, results.length);
          return;
        }
        clearInterval(id);
      }, interval);
    }
    if ((count = res.length))
      setTimeout(
        () => {
          if (pid == current) {
            let i = 1;
            const id = setInterval(() => {
              if (i <= 5 && i <= count && pid == current) {
                results = res.slice(0, i);
                i++;
                return;
              }
              clearInterval(id);
              if (count > 5 && pid == current) results = res;
            }, interval);
          }
        },
        length ? length * interval + 300 : 0
      );
  };

  const addlink = (option: { b: string; g: string; s: object }) => {
    let result = option.g;
    if (option.s)
      Object.entries(option.s).forEach(([key, value]) => {
        result = result.replaceAll(
          `『${key}』`,
          `<a href="${value}" target="_blank">『${key}』</a>`
        );
      });
    return result;
  };

  const reset = () => {
    $query = "";
    results = [];
  };
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="content" on:mousedown={showFilter.off}>
  <div class="input-group">
    <input
      class="form-control"
      type="search"
      placeholder="ウマ娘名、イベント、選択肢テキスト"
      bind:value={$query}
      on:keydown={(e) => {
        if (e.key === "Escape") reset();
      }}
      on:input={search}
    />
  </div>
  <div class="message">
    <span>{`※${$filter.name ? 1 : 2}文字以上入力すると検索を開始します`}</span>
    <span style="float:right">{`(${count}/${$events.length})`}</span>
  </div>
  <div id="results">
    {#each results as result (result.c + result.r + result.e + result.k + JSON.stringify(result.o))}
      <table
        class="table table-bordered"
        in:fly={{ x: "100%", duration: 400 }}
        out:fly={{ x: "100%", duration: 400 }}
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
                  <!-- svelte-ignore a11y-click-events-have-key-events -->
                  <Image
                    id={result.i}
                    alt={result.c}
                    type="link"
                    on:click={() => {
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
