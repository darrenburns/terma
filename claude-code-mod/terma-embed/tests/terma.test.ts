import { expect, test } from 'claude-code/testing'

const FRAME = {
  cols: 12,
  rows: 2,
  styles: [{}, { fg: '#ff8800', b: true }],
  lines: [[['Hello ', 0], ['terma', 1]], [['second row', 0]]],
}

const PANE_PROPS = {
  title: 'terma',
  isFocused: true,
  bodyColumns: 40,
  placement: 'dock' as const,
  scroll: { offset: 0, bodyRows: 10 },
  view: {},
}

test('/terma builds the app, draws its frames and forwards keys to its socket', async ($, on) => {
  const builds: string[][] = []
  const posted: unknown[] = []
  let release = () => {}
  const held = new Promise<void>(resolve => (release = resolve))
  let framed = () => {}
  const frameStored = new Promise<void>(resolve => (framed = resolve))

  on('process.run', async (_$, e) => {
    builds.push([...e.argv])
    return { value: { exitCode: 0, stdout: '', stderr: '', isStdoutTruncated: false, isStderrTruncated: false } }
  })
  on('process.spawn', async function* (_$, e) {
    expect(e.env?.TERMA_EMBED_SOCKET).toMatch(/\.sock$/)
    // A frame split across two reads, as a pipe may deliver it.
    yield { stream: 'stdout' as const, text: JSON.stringify(FRAME).slice(0, 20) }
    yield { stream: 'stdout' as const, text: JSON.stringify(FRAME).slice(20) + '\n' }
    // The mod pulls again only once it has stored the frame it just read.
    framed()
    await held
    return { value: { code: 0, signal: null } }
  })
  on('http.fetch', async (_$, e) => {
    posted.push(JSON.parse(String(e.init?.body)))
    return { value: { ok: true, status: 204, headers: {}, text: '' } }
  })

  on('ui.open', async () => ({ value: { isPlaced: true as const } }))
  await $.command.run({
    command: 'terma',
    args: '',
    origin: { kind: 'composer' },
    presentation: { isFullscreen: true, columns: 160 },
  })
  await frameStored
  expect(builds[0]).toEqual(['go', 'build', '-o', '/tmp/terma-embed-app', './cmd/terma-demos'])

  const ui = await $.ui.mount({
    plugin: 'terma-embed',
    surface: 'terminal',
    component: 'Pane',
    requestId: 'terma',
    props: PANE_PROPS,
  })
  await ui.resize({ columns: 40, rows: 10, in: 'screen' })

  const shown = await ui.find({ type: 'Text', text: /Hello terma/, in: 'screen' })
  expect(shown).toBeDefined()

  await ui.key({ key: 'down', in: 'screen' })
  await ui.key({ key: ']', ctrl: true, in: 'screen' })
  const events = posted.flat() as { t: string; key?: string; cols?: number }[]
  expect(events).toContainEqual({ t: 'resize', cols: 40, rows: 10 })
  expect(events).toContainEqual({ t: 'key', key: 'down' })
  expect(events).toContainEqual({ t: 'key', key: 'escape' })
  expect(events.filter(one => one.t === 'key' && one.key === 'down')).toHaveLength(1)

  release()
  await ui.unmount()
})
