import { computed, type Ref } from 'vue'
export function useRelativeTime(input: Ref<string>) { return computed(() => { const delta = Date.now() - new Date(input.value).getTime(); const minutes = Math.floor(delta / 60000); if (minutes < 60) return `${Math.max(minutes, 0)} 分钟前`; const hours = Math.floor(minutes / 60); return `${hours} 小时前` }) }
