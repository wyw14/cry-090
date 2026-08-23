export type DanceStyle = 'cuban' | 'la' | 'ny'
export type EventStatus = 'draft' | 'review' | 'published' | 'cancelled' | 'completed'
export interface Session { id: string; start: string; end: string; styles: DanceStyle[] }
export interface EventSummary { id: string; title: string; venue: string; status: EventStatus; priceMinor: number; currency: string; capacity: number; registered: number; sessions: Session[] }
export interface MatchReason { code: string; message: string; weight: number }
export interface MatchResult { needId: string; user: { userId: string; displayName?: string; danceYears?: number; roles?: string[]; tags?: string[]; distanceBucket?: number }; score: number; reasons: MatchReason[] }
