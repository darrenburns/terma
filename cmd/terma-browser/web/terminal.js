'use strict';
const status = document.getElementById('status');
const container = document.getElementById('terminal');
const params = new URLSearchParams(location.search);
const fixedCols = Number(params.get('cols'));
const fixedRows = Number(params.get('rows'));
const fixedSize = Number.isInteger(fixedCols) && Number.isInteger(fixedRows) && fixedCols >= 2 && fixedCols <= 500 && fixedRows >= 2 && fixedRows <= 300;
const terminal = new Terminal({
  cols: fixedSize ? fixedCols : 100, rows: fixedSize ? fixedRows : 30,
  fontFamily: 'Menlo, Consolas, monospace', fontSize: 14,
  cursorBlink: false, scrollback: 1000, screenReaderMode: true,
  theme: { background: '#101216' }
});
terminal.open(container);
function fit() {
  if (fixedSize) return;
  const screen = container.querySelector('.xterm-screen');
  const cellWidth = screen.getBoundingClientRect().width / terminal.cols;
  const cellHeight = screen.getBoundingClientRect().height / terminal.rows;
  if (cellWidth > 0 && cellHeight > 0) {
    terminal.resize(Math.max(2, Math.min(500, Math.floor((container.clientWidth - 32) / cellWidth))), Math.max(2, Math.min(300, Math.floor((container.clientHeight - 16) / cellHeight))));
  }
}
fit();
const token = location.hash.slice(1);
const socket = new WebSocket(`ws://${location.host}/session?token=${encodeURIComponent(token)}`);
socket.binaryType = 'arraybuffer';
function send(message) {
  if (socket.readyState === WebSocket.OPEN) socket.send(JSON.stringify(message));
}
socket.onopen = () => {
  send({type: 'resize', cols: terminal.cols, rows: terminal.rows});
  status.textContent = 'Connected';
  terminal.focus();
};
socket.onmessage = event => {
  if (typeof event.data === 'string') {
    status.textContent = JSON.parse(event.data).message;
  } else {
    terminal.write(new Uint8Array(event.data));
  }
};
socket.onclose = () => {
  if (status.textContent === 'Connected' || status.textContent === 'Connecting…') status.textContent = 'Disconnected. Reload to start a new session.';
};
terminal.onData(data => send({type: 'input', data}));
terminal.onResize(({cols, rows}) => send({type: 'resize', cols, rows}));
new ResizeObserver(fit).observe(container);
document.getElementById('restart').onclick = () => location.reload();
window.addEventListener('pagehide', () => socket.close());
