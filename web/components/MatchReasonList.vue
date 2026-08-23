<script setup lang="ts">
import type { MatchReason } from '../types/domain'
const props = defineProps<{ reasons: Array<MatchReason> }>()
</script>
<template><ul class="reason-list"><li v-for="reason in props.reasons" :key="reason.code"><span>{{ reason.message }}</span><el-progress :percentage="Math.min(reason.weight * 2, 100)" :show-text="false" /></li></ul></template>
