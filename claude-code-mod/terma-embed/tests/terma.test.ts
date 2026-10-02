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
  on('env.get', async () => ({ value: '/home/me' }))
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

test('/terma <dir> builds an app outside the repo against a terma with embedded mode', async ($, on) => {
  const APP = '/home/me/code/posting/cmd/posting'
  const runs: { argv: string[]; cwd?: string; env?: Record<string, string> }[] = []
  const written: { path: string; text: string }[] = []
  let spawned: { argv: string[]; cwd?: string } | undefined
  let started = () => {}
  const didStart = new Promise<void>(resolve => (started = resolve))

  on('ui.open', async () => ({ value: { isPlaced: true as const } }))
  on('env.get', async () => ({ value: '/home/me' }))
  on('fs.stat', async (_$, e) => {
    const known: Record<string, 'dir' | 'file'> = { [APP]: 'dir', '/home/me/code/terma/embed.go': 'file' }
    const kind = known[e.path]
    if (!kind) return { deny: 'ENOENT' }
    return { value: { kind, size: 0, mtimeMs: 0, isLink: false } }
  })
  on('fs.write', async (_$, e) => {
    written.push({ path: e.path, text: e.text })
    return { value: undefined }
  })
  on('process.run', async (_$, e) => {
    runs.push({ argv: [...e.argv], cwd: e.init?.cwd, env: e.init?.env ? { ...e.init.env } : undefined })
    const out = (stdout: string) => ({
      value: { exitCode: 0, stdout, stderr: '', isStdoutTruncated: false, isStderrTruncated: false },
    })
    const args = e.argv.join(' ')
    // The app resolves terma to a checkout without embed.go; the session to one with it.
    if (args.startsWith('go list')) return out(e.init?.cwd === APP ? '/home/me/code/jump-mode' : '/home/me/code/terma')
    if (args === 'go env GOMOD') return out('/home/me/code/posting/go.mod')
    if (args === 'go env GOVERSION') return out('go1.25.5')
    return out('')
  })
  on('process.spawn', async function* (_$, e) {
    spawned = { argv: [...e.argv], cwd: e.cwd }
    started()
    yield { stream: 'stdout' as const, text: '' }
    return { value: { code: 0, signal: null } }
  })

  await $.command.run({
    command: 'terma',
    args: '~/code/posting/cmd/posting --collection demo',
    origin: { kind: 'composer' },
    presentation: { isFullscreen: true, columns: 160 },
  })
  await didStart

  expect(written).toEqual([
    {
      path: '/tmp/terma-embed.go.work',
      text: 'go 1.25.5\n\nuse /home/me/code/posting\n\nreplace github.com/darrenburns/terma => /home/me/code/terma\n',
    },
  ])
  expect(runs.find(one => one.argv[1] === 'build')).toEqual({
    argv: ['go', 'build', '-o', '/tmp/terma-embed-app', '.'],
    cwd: APP,
    env: { GOWORK: '/tmp/terma-embed.go.work' },
  })
  expect(spawned).toEqual({ argv: ['/tmp/terma-embed-app', '--collection', 'demo'], cwd: APP })
})
