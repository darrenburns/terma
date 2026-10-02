import type { ClientModule } from 'claude-code'

import type { TermaEvent, TermaScreenProps } from '../types'

type Local = { cols: number; rows: number }

// Events stay queued until a later post carries them: a post the next one
// replaces within a frame is never delivered, so each post repeats the tail.
let seq = 0
let outbox: { seq: number; event: TermaEvent }[] = []

function queue(event: TermaEvent) {
  const last = outbox.at(-1)
  if (event.t === 'move' && last?.event.t === 'move') outbox.pop()
  outbox.push({ seq: ++seq, event })
  outbox = outbox.slice(-64)
}

const Screen: ClientModule<TermaScreenProps, Local> = (props, surface) => {
  const { Box, Text } = surface.elements

  if (surface.state === undefined) {
    surface.onKey(k => {
      // Escape never reaches a Client, so ctrl+] stands in for it.
      const key = k.ctrl && k.key === ']' ? { key: 'escape' } : { key: k.key, ctrl: k.ctrl, shift: k.shift, meta: k.meta }
      queue({ t: 'key', ...key })
      surface.post({ events: outbox })
    })
    surface.onPointer(p => {
      if (p.type === 'enter' || p.type === 'leave') return
      queue({ t: p.type, x: p.x, y: p.y, button: p.button })
      surface.post({ events: outbox })
    })
  }
  if (surface.columns > 0 && (surface.state?.cols !== surface.columns || surface.state?.rows !== surface.rows)) {
    queue({ t: 'resize', cols: surface.columns, rows: surface.rows })
    surface.post({ events: outbox })
    surface.setState({ cols: surface.columns, rows: surface.rows })
  } else if (surface.state === undefined) {
    surface.setState({ cols: 0, rows: 0 })
  }

  const { frame, status } = props
  if (!frame) {
    const label =
      status.phase === 'building' ? `Building ${status.detail}…`
      : status.phase === 'failed' ? `Build failed:\n${status.detail ?? ''}`
      : status.phase === 'exited' ? `The app exited.${status.detail ? `\n${status.detail}` : ''}`
      : 'Starting…'
    return <Text dimColor>{label}</Text>
  }

  return (
    <Box flexDirection="column">
      {frame.lines.map(runs => (
        <Text wrap="truncate-end">
          {runs.map(([text, index]) => {
            const s = frame.styles[index] ?? {}
            return (
              <Text
                color={s.fg}
                backgroundColor={s.bg}
                bold={s.b}
                dimColor={s.d}
                italic={s.i}
                underline={s.u}
                strikethrough={s.s}
                inverse={s.r}
              >
                {text}
              </Text>
            )
          })}
        </Text>
      ))}
      {status.phase === 'exited' && <Text dimColor>The app exited. Run /terma to start it again.</Text>}
    </Box>
  )
}

export default Screen
