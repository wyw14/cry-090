<script setup lang="ts">
import type { EventSummary } from '../types/domain'
defineProps<{ event: EventSummary }>()
const emit = defineEmits<{ register: [id: string] }>()
</script>
<template><el-card class="event-card"><div class="event-head"><div><h2>{{ event.title }}</h2><p>{{ event.venue }}</p></div><el-tag :type="event.status === 'published' ? 'success' : 'info'">{{ event.status }}</el-tag></div><div class="event-meta"><span>{{ event.sessions.length }} 个时段</span><span>{{ (event.priceMinor / 100).toFixed(2) }} {{ event.currency }}</span><span>{{ event.registered }}/{{ event.capacity }} 人</span></div><el-button type="primary" @click="emit('register', event.id)">加入舞会</el-button></el-card></template>
