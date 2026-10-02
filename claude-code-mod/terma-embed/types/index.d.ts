/** One style of a frame, as terma's embedded mode writes it (embed.go). */
export type TermaStyle = {
  fg?: string
  bg?: string
  b?: boolean
  d?: boolean
  i?: boolean
  u?: boolean
  s?: boolean
  r?: boolean
}

/** One screen: per line, runs of [text, index into styles]. */
export type TermaFrame = {
  cols: number
  rows: number
  styles: TermaStyle[]
  lines: [string, number][][]
  focus?: string
}

export type TermaStatus = {
  phase: 'idle' | 'building' | 'running' | 'exited' | 'failed'
  detail?: string
}

/** An input event in embed.go's shape. */
export type TermaEvent =
  | { t: 'key'; key: string; ctrl?: boolean; shift?: boolean; meta?: boolean }
  | { t: 'down' | 'up' | 'move'; x: number; y: number; button?: string }
  | { t: 'resize'; cols: number; rows: number }

/** What the screen module posts: every event not yet known delivered, numbered. */
export type TermaPost = { events: { seq: number; event: TermaEvent }[] }

export type TermaScreenProps = { frame: TermaFrame | null; status: TermaStatus }

declare module 'claude-code' {
  interface PluginState {
    'terma-embed': { frame: TermaFrame | null; status: TermaStatus }
  }
}
