<script>
  import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';

  let { children, ...restProps } = $props();
</script>

<AlertDialogPrimitive.Portal {...restProps}>
  {@render children?.()}
</AlertDialogPrimitive.Portal>
