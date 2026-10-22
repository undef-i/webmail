<script module>
  import { tv } from 'tailwind-variants';

  export const buttonVariants = tv({
    base: 'inline-flex items-center justify-center gap-2 whitespace-nowrap text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0',
    variants: {
      variant: {
        default: 'border border-input bg-secondary text-secondary-foreground hover:bg-accent active:bg-muted',
        destructive: 'border border-input bg-secondary text-destructive hover:bg-accent',
        outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
        secondary: 'border border-foreground bg-foreground text-background hover:opacity-85',
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        link: 'text-link underline hover:opacity-80'
      },
      size: {
        default: 'h-8 px-3 py-1',
        sm: 'h-7 px-2 text-xs',
        lg: 'h-9 px-6',
        icon: 'h-8 w-8'
      }
    },
    defaultVariants: {
      variant: 'default',
      size: 'default'
    }
  });
</script>

<script>
  import { cn } from '$lib/utils.js';

  let {
    class: className,
    variant = 'default',
    size = 'default',
    ref = $bindable(null),
    href = undefined,
    type = 'button',
    children,
    ...restProps
  } = $props();
</script>

{#if href}
  <a bind:this={ref} class={cn(buttonVariants({ variant, size }), className)} {href} {...restProps}>
    {@render children?.()}
  </a>
{:else}
  <button bind:this={ref} class={cn(buttonVariants({ variant, size }), className)} {type} {...restProps}>
    {@render children?.()}
  </button>
{/if}
