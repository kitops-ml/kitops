<script setup lang="ts">
import BunTop from '../assets/css/images/kernel-krunch/bun-top.svg?component';
import Sesame from '../assets/css/images/kernel-krunch/sesame.svg?component';
import Glasses from '../assets/css/images/kernel-krunch/glasses.svg?component';
import Mouth from '../assets/css/images/kernel-krunch/mouth.svg?component';
import Salad from '../assets/css/images/kernel-krunch/salad.svg?component';
import Tomatoes from '../assets/css/images/kernel-krunch/tomatos.svg?component';
import Cheese from '../assets/css/images/kernel-krunch/cheese.svg?component';
import Patty from '../assets/css/images/kernel-krunch/patty.svg?component';
import BunBottom from '../assets/css/images/kernel-krunch/bun-bottom.svg?component';
</script>

<template>
  <div class="relative kernel-krunch">
    <!-- Top Bun -->
    <div class="absolute top-0 z-5">
      <BunTop />
      <Sesame class="absolute top-2.5 right-2" />
      <Glasses class="absolute top-0 left-3" />
      <Mouth class="absolute bottom-5 left-1/2 -translate-x-[calc(50%-2px)]" />
    </div>

    <!-- Salad -->
    <Salad class="absolute top-[32%] -left-2 z-4" />

    <!-- Tomatoes -->
    <Tomatoes class="absolute top-[42%] z-3" />

    <!-- Cheese -->
    <Cheese class="absolute top-[52%] -left-2 z-2" />

    <!-- Patty -->
    <Patty class="absolute top-[60%] z-1" />

    <!-- Bottom Bun -->
    <BunBottom class="absolute bottom-0 -left-1" />
  </div>
</template>
<!--
<style>
.kernel-krunch {
  transform-origin: bottom left;
  animation: breathe 1.5s ease-in-out infinite;
}

@keyframes breathe {
  0%, 100% {
    transform: scaleY(1) skewX(0deg);
  }
  50% {
    transform: scaleY(1.05) skewX(-2deg);
  }
}
</style> -->
