import { atom, read } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type { TermaEvent, TermaFrame, TermaPost, TermaStatus } from '../types'

const PANE = 'terma'
const DEFAULT_PACKAGE = './cmd/terma-demos'
const BINARY = '/tmp/terma-embed-app'
const WORKSPACE = '/tmp/terma-embed.go.work'
const TERMA_MODULE = 'github.com/darrenburns/terma'

const FRAME = { plugin: 'terma-embed', key: 'frame' } as const
const STATUS = { plugin: 'terma-embed', key: 'status' } as const
const frameAtom = atom(FRAME, null as TermaFrame | null)
const statusAtom = atom(STATUS, { phase: 'idle' } as TermaStatus)

// The child and its socket live with this module: a reload kills the child.
let child: { socket: string; stop: () => void } | undefined
let lastSeq = 0
let size = { cols: 80, rows: 24 }

async function send($: EngineInterface, events: TermaEvent[]) {
  if (!child || events.length === 0) return
  await $.http
    .fetch('http://terma/events', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(events),
      socketPath: child.socket,
    })
    .catch(() => undefined)
}

/** What /terma runs: a package to build first, or a binary as it is. */
type Launch = {
  label: string
  build?: { pkg: string; cwd?: string; env?: Record<string, string> }
  binary: string
  cwd?: string
  args: string[]
}

const lastLines = (text: string) => text.trim().split('\n').slice(-6).join('\n')

async function goOutput($: EngineInterface, argv: string[], cwd?: string) {
  const ran = await $.process.run(['go', ...argv], { cwd, timeoutMs: 60_000 })
  return ran.exitCode === 0 ? ran.stdout.trim() : undefined
}

async function hasEmbedMode($: EngineInterface, termaDir: string | undefined) {
  if (!termaDir) return false
  return $.fs.stat(`${termaDir}/embed.go`).then(() => true, () => false)
}

/**
 * Turns /terma's argument into a Launch. An absolute path names a binary, or
 * a directory holding a main package anywhere on disk; anything else is a
 * package in the session's directory. A directory whose terma dependency has
 * no embedded mode is built against the session's terma through a GOWORK file
 * of its own, so its go.mod and go.work stay untouched.
 */
async function resolve($: EngineInterface, arg: string): Promise<Launch | string> {
  const [first = DEFAULT_PACKAGE, ...args] = arg.split(/\s+/).filter(Boolean)
  const home = (await $.env.get('HOME')) ?? ''
  const target = first.replace(/^~(?=\/|$)/, home)
  if (!target.startsWith('/')) {
    return { label: target, build: { pkg: target }, binary: BINARY, args }
  }

  const found = await $.fs.stat(target).catch(() => undefined)
  if (!found) return `No such file or directory: ${target}`
  if (found.kind === 'file') return { label: target, binary: target, cwd: undefined, args }

  const own = await goOutput($, ['list', '-m', '-f', '{{.Dir}}', TERMA_MODULE], target)
  if (await hasEmbedMode($, own)) {
    return { label: target, build: { pkg: '.', cwd: target }, binary: BINARY, cwd: target, args }
  }
  const session = await goOutput($, ['list', '-m', '-f', '{{.Dir}}', TERMA_MODULE])
  if (!session || !(await hasEmbedMode($, session))) {
    return `${target} builds against a terma without embedded mode (${own ?? 'not found'}), and this session's directory has no terma with it to swap in.`
  }
  const goMod = await goOutput($, ['env', 'GOMOD'], target)
  const version = (await goOutput($, ['env', 'GOVERSION']))?.replace(/^go/, '')
  if (!goMod || goMod === '/dev/null' || !version) return `${target} is not inside a Go module.`
  await $.fs.write(
    WORKSPACE,
    `go ${version}\n\nuse ${goMod.replace(/\/go\.mod$/, '')}\n\nreplace ${TERMA_MODULE} => ${session}\n`,
  )
  return {
    label: `${target} (terma from ${session})`,
    build: { pkg: '.', cwd: target, env: { GOWORK: WORKSPACE } },
    binary: BINARY,
    cwd: target,
    args,
  }
}

async function start($: EngineInterface, arg: string) {
  child?.stop()
  child = undefined
  await $.state.set(FRAME, null)
  await $.state.set(STATUS, { phase: 'building', detail: arg || DEFAULT_PACKAGE })

  const launch = await resolve($, arg)
  if (typeof launch === 'string') {
    await $.state.set(STATUS, { phase: 'failed', detail: launch })
    return
  }
  if (launch.build) {
    await $.state.set(STATUS, { phase: 'building', detail: launch.label })
    const { pkg, cwd, env } = launch.build
    const built = await $.process.run(['go', 'build', '-o', BINARY, pkg], { cwd, env, timeoutMs: 300_000 })
    if (built.exitCode !== 0) {
      await $.state.set(STATUS, { phase: 'failed', detail: lastLines(built.stderr) })
      return
    }
  }

  const socket = `/tmp/terma-embed-${Math.random().toString(36).slice(2, 10)}.sock`
  const stream = $.process.spawn({
    argv: [launch.binary, ...launch.args],
    cwd: launch.cwd,
    env: { TERMA_EMBED_SOCKET: socket, TERMA_EMBED_SIZE: `${size.cols}x${size.rows}` },
  })
  const mine = { socket, stop: () => void stream.return?.(undefined as never) }
  child = mine
  lastSeq = 0
  await $.state.set(STATUS, { phase: 'running', detail: launch.label })

  void (async () => {
    let pending = ''
    let stderr = ''
    try {
      for await (const { stream: pipe, text } of stream) {
        if (pipe === 'stderr') {
          stderr = (stderr + text).slice(-2000)
          continue
        }
        pending += text
        const lines = pending.split('\n')
        pending = lines.pop() ?? ''
        const latest = lines.at(-1)
        if (latest) {
          await $.state.set(FRAME, JSON.parse(latest) as TermaFrame)
        }
      }
    } catch (err) {
      stderr += String(err)
    }
    if (child !== mine) return
    child = undefined
    await $.state.set(STATUS, { phase: 'exited', detail: lastLines(stderr) || undefined })
    // The module may have unloaded (a reload kills the child), leaving no state to write.
  })().catch(() => undefined)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'terma',
      description:
        'Run a terma app in a pane: /terma [package | /abs/dir | /abs/binary] [args...], default ./cmd/terma-demos; /terma stop',
    })
    return next(e)
  })

  on('command.run', { command: 'terma' }, async ($, e) => {
    const arg = e.args.trim()
    if (arg === 'stop') {
      child?.stop()
      await $.ui.close({ id: PANE })
      return { text: 'Stopped the terma app.' }
    }
    await $.ui.open({ id: PANE, title: 'terma', focus: true })
    void start($, arg || DEFAULT_PACKAGE)
    return { text: `Running ${arg || DEFAULT_PACKAGE} in the terma pane. Click it to type; ctrl+] sends Escape.` }
  })

  on('ui.message', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const post = e.data as TermaPost
    const newest = post.events.at(-1)?.seq ?? 0
    // A screen module that reloaded numbers from 1 again.
    if (newest < lastSeq) lastSeq = 0
    const fresh = post.events.filter(one => one.seq > lastSeq).map(one => one.event)
    lastSeq = Math.max(lastSeq, newest)
    for (const event of fresh) {
      if (event.t === 'resize') size = { cols: event.cols, rows: event.rows }
    }
    await send($, fresh)
    return {}
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const frame = await read($, frameAtom)
    const status = await read($, statusAtom)
    if (e.surface === 'terminal' || e.surface === 'desktop') {
      const { Client } = $.ui.resolve(e)
      const rows = Math.max(8, e.props.scroll.bodyRows || (e.viewport?.rows ?? 30) - 6)
      return <Client key="screen" module="./screen.tsx" props={{ frame, status }} width="100%" height={rows} />
    }
    const { Text } = $.ui.resolve(e)
    return <Text dimColor>terma draws in the terminal and the desktop app only.</Text>
  })
}
