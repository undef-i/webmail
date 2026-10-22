<script>
  import { Dialog as DialogPrimitive } from 'bits-ui';

  let { ref = $bindable(null), children, ...restProps } = $props();
</script>

<DialogPrimitive.Close bind:ref {...restProps}>
  {@render children?.()}
</DialogPrimitive.Close>
