<script setup lang="ts">const emit = defineEmits<{ change: [style: string] }>(); const onChange = (value: string | number) => emit('change', String(value));</script>
<template><el-segmented :options="['全部','cuban','la','ny']" @change="onChange" /></template>
