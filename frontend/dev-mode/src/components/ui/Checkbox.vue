<script setup lang="ts">
import { useVModel } from '@vueuse/core'
import { useSlots, useAttrs, defineOptions, type HTMLAttributes } from 'vue'

defineOptions({
  inheritAttrs: false
})

const props = withDefaults(defineProps<{
  id?: string,
  wrapperClass?: HTMLAttributes['class'],
  class?: HTMLAttributes['class'],
  label?: string,
  modelValue?: boolean
}>(), {
  id: () => `radio-${Math.random().toString(24).slice(2)}`
})

const emit = defineEmits(['update:modelValue'])

const slots = useSlots()
const attrs = useAttrs()
const vModel = useVModel(props, 'modelValue', emit)

const labelClassName = 'block font-bold text-off-white text-left pl-2'
</script>

<template>
<div class="flex items-center" :class="props.wrapperClass">
  <input
    :id="props.id"
    type="checkbox"
    v-bind="attrs"
    :class="props.class"
    v-model="vModel">

  <slot v-if="props.label || slots.label" name="label" :id="props.id" :className="labelClassName">
    <label :for="props.id" :class="labelClassName">
      {{ props.label }}
    </label>
  </slot>
</div>
</template>

<style scoped>
input {
  @apply w-4 h-4 border border-off-white;
  appearance: none;
  transition: border 150ms;
}

input:checked {
  @apply border-[0.5rem] border-gold;
}
</style>
