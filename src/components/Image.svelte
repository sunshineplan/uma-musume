<script lang="ts">
  import { db } from "../stores";

  export let id: string;
  export let alt: string;
  export let selected: boolean = false;
  export let title: string = "";
  export let type: string = "";
  export let style: string = "";

  let src: string =
    "data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEAAAAALAAAAAABAAEAgAAAAAAAAAICRAEAOw==";

  $: id, load();

  const load = async () => {
    let image: Blob;
    if (!id) return;
    const res = await db.transaction("r", db.table("images"), () => {
      return db.table("images").get({ id });
    });
    if (res) image = res.image;
    if (!image || !image.size) {
      let url = `/image/${id}`;
      if (!id.endsWith(".png")) url = `/support/${id}.png`;
      const resp = await fetch(url);
      image = await resp.blob();
      if (image.size) db.table("images").put({ id, image });
    }
    if (image.size) src = URL.createObjectURL(image);
    else src = "";
  };
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<img class={type} class:selected {style} {src} {title} {alt} on:click />

<style>
  img {
    width: 72px;
    max-width: 72px;
  }

  .type {
    height: 26px;
    width: 26px;
    border-radius: 5px;
  }

  .rare {
    height: 20px;
    width: 50px;
    border-radius: 15px;
  }

  .icon {
    border-width: 2px;
    border-style: solid;
    border-radius: 8px;
    border-color: transparent;
  }

  .link {
    min-height: 72px;
    margin-right: 10px;
    cursor: pointer;
  }

  .selected {
    border-color: #007bff !important;
  }
</style>
