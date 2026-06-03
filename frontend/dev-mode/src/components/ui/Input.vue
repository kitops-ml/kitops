<script setup lang="ts">
import { useVModel } from '@vueuse/core'
import { ref, useSlots, useAttrs, defineOptions, type HTMLAttributes } from 'vue'

defineOptions({
  inheritAttrs: false
})

const props = withDefaults(defineProps<{
  id?: string,
  wrapperClass?: HTMLAttributes['class'],
  class?: HTMLAttributes['class'],
  label?: string,
  description?: string,
  modelValue?: string | number | boolean
}>(), {
  id: () => `input-${Math.random().toString(24).slice(2)}`
})

const emit = defineEmits(['update:modelValue'])

const inputRef = ref(null)

const slots = useSlots()
const attrs = useAttrs()
const vModel = useVModel(props, 'modelValue', emit)

const labelClassName = 'block font-bold text-off-white text-left mb-1'

defineExpose({
  inputRef
})
</script>

<template>
<div :class="props.wrapperClass">
  <slot v-if="props.label || slots.label" name="label" :id="props.id" :className="labelClassName">
    <label :for="props.id" :class="labelClassName">
      {{ props.label }}
    </label>
  </slot>

  <div
    class="flex items-center input w-full"
    :class="props.class">

    <div v-show="slots.leading" class="pr-4 flex items-center">
      <slot name="leading" />
    </div>

    <input
      ref="inputRef"
      :id="props.id"
      v-bind="attrs"
      :class="{ '!pr-0': slots.trailing }"
      :type="(attrs.type as string) || 'text'"
      v-model="vModel">

    <div v-show="slots.trailing" class="pl-4 flex items-center">
      <slot name="trailing" />
    </div>
  </div>

  <p v-if="props.description || slots.description" class="text-xs text-gray-06 text-left mt-2">
    <slot name="description">{{ props.description }}</slot>
  </p>
</div>
</template>

<style scoped>
.input {
  @apply border border-gray-05 placeholder:text-gray-05 text-off-white bg-transparent;
  @apply focus-within:border-gold;
}

.input > input {
  @apply block px-4 py-2 focus:outline-none flex-1 bg-transparent w-full;
}
</style>
