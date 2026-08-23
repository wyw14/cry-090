<script setup lang="ts">
import { onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useEventStore } from '../stores/eventStore'
import EventCard from '../components/EventCard.vue'
import EventFilters from '../features/events/EventFilters.vue'
const store = useEventStore()
onMounted(() => store.load())
async function register(id: string) { try { await store.load(); ElMessage.success(`已处理舞会 ${id}`) } catch { ElMessage.error('报名暂时失败') } }
</script>
<template><main class="shell"><header class="topbar"><div><span class="eyebrow">SALSA CIRCLE</span><h1>今晚去跳舞</h1></div><nav><router-link to="/events">舞会</router-link><router-link to="/profile">我的公开资料</router-link></nav></header><section class="toolbar"><EventFilters /><el-button :loading="store.loading" @click="store.load()">刷新</el-button></section><el-alert v-if="store.error" :title="store.error" type="error" show-icon /><section class="event-grid"><EventCard v-for="event in store.items" :key="event.id" :event="event" @register="register" /></section><el-empty v-if="!store.loading && !store.items.length" description="还没有符合条件的舞会" /></main></template>
