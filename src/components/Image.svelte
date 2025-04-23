<script lang="ts">
  import { onMount } from "svelte";
  import { uma } from "../uma.svelte";

  let {
    id,
    alt,
    selected = false,
    title = "",
    type = "",
    style = "",
    onclick,
  }: {
    id: string;
    alt: string;
    selected?: boolean;
    title?: string;
    type?: string;
    style?: string;
    onclick?: () => void;
  } = $props();

  let src = $state("");
  let imageElement: HTMLImageElement;

  const blank =
    "data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEAAAAALAAAAAABAAEAgAAAAAAAAAICRAEAOw==";

  onMount(() => {
    const observer = new IntersectionObserver(async ([entry]) => {
      if (entry.isIntersecting) {
        observer.disconnect();
        if (!id) return;
        let image = await uma.loadImage(id);
        if (!image?.size) {
          let url = `/image/${id}`;
          if (!id.endsWith(".png")) url = `/support/${id}.png`;
          try {
            const resp = await fetch(url);
            if (resp.ok) {
              image = await resp.blob();
              if (id && image.size) await uma.saveImage({ id, image });
            }
          } catch {}
        }
        if (image?.size) src = URL.createObjectURL(image);
      }
    });
    observer.observe(imageElement);
    return () => observer.disconnect();
  });
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<img
  bind:this={imageElement}
  class={type}
  class:selected
  {style}
  src={src || blank}
  {title}
  {alt}
  {onclick}
  loading="lazy"
/>

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
