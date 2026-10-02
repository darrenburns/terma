#!/usr/bin/env python3
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
WIDGET_PAGES = {
    'Autocomplete': 'widgets/autocomplete.md', 'Breadcrumbs': 'widgets/breadcrumbs.md',
    'Button': 'widgets/button.md', 'Checkbox': 'widgets/checkbox.md',
    'CommandPalette': 'widgets/commandpalette.md', 'Dialog': 'widgets/dialog.md',
    'DirectoryTree': 'widgets/directorytree.md', 'EmptyWidget': 'widgets/empty.md',
    'Floating': 'floating.md', 'FocusTrap': 'widgets/focustrap.md',
    'Image': 'widgets/image.md', 'Jumper': 'widgets/jumper.md',
    'KeybindBar': 'widgets/keybindbar.md', 'List': 'widgets/list.md',
    'Menu': 'widgets/menu.md', 'PresentedText': 'widgets/presentedtext.md',
    'ProgressBar': 'widgets/progressbar.md', 'Sparkline': 'widgets/sparkline.md',
    'Spinner': 'widgets/spinner.md', 'Switcher': 'widgets/switcher.md',
    'Table': 'widgets/table.md', 'TabBar': 'widgets/tabs.md', 'TabView': 'widgets/tabs.md',
    'Text': 'widgets/text.md', 'TextArea': 'widgets/textarea.md',
    'TextInput': 'widgets/textinput.md', 'Tooltip': 'widgets/tooltip.md',
    'Tree': 'widgets/tree.md', 'Row': 'layout/row-column.md',
    'Column': 'layout/row-column.md', 'Dock': 'layout/dock.md',
    'Scrollable': 'layout/scrollable.md', 'Spacer': 'layout/spacer.md',
    'SplitPane': 'layout/splitpane.md', 'Stack': 'layout/stack.md',
    'Positioned': 'layout/stack.md',
}


ADDITIONAL_WIDGET_PAGES = {
    "SelectBox": "widgets/select.md", "Form": "widgets/form.md",
    "Field": "widgets/form.md", "Markdown": "widgets/markdown.md",
    "FilePicker": "widgets/filepicker.md",
}


def prose_lines(text):
    fenced = False
    for number, raw in enumerate(text.splitlines(), 1):
        line = raw.strip()
        if line.startswith('```'):
            fenced = not fenced
            continue
        if fenced or not line or line.startswith(('#', '<!--')):
            continue
        image = re.fullmatch(r'!\[([^]]*)\]\([^)]*\)', line)
        if image:
            yield number, image.group(1).strip()
        elif not re.fullmatch(r'\|?[\s:|\-]+\|?', line):
            yield number, re.sub(r'^[-*] |^\d+\. ', '', line)


def main():
    errors = []
    pages = {'docs/' + path for path in WIDGET_PAGES.values()}
    pages.update({'docs/widgets/index.md', 'docs/widgets/custom-widgets.md', 'docs/layout/index.md'})
    public = set()
    for path in ROOT.glob('*.go'):
        if not path.name.endswith('_test.go'):
            public.update(re.findall(r'^func \(\w+ \*?([A-Z]\w*)(?:\[[^]]+\])?\) Build\(', path.read_text(), re.M))
    documented = set(WIDGET_PAGES) | set(ADDITIONAL_WIDGET_PAGES)
    if public != documented:
        errors.append(f'widget coverage differs: missing {public - documented}, obsolete {documented - public}')
    for page in set(ADDITIONAL_WIDGET_PAGES.values()):
        if not (ROOT / "docs" / page).is_file():
            errors.append(f"missing additional widget guide: {page}")
    records = []
    for path in (ROOT / 'docs/widget-verification').glob('*.json'):
        records.extend(json.loads(path.read_text()))
    verified = {}
    for record in records:
        key = (record['page'], record['text'].strip())
        verified[key] = record
        source = ROOT / record['source']
        if not source.is_file():
            errors.append(f'missing evidence source: {source}')
        elif record.get('source_sha256') != hashlib.sha256(source.read_bytes()).hexdigest():
            errors.append(f'evidence source changed; review claim again: {key}')
        elif not record.get('evidence') or not record.get('symbol'):
            errors.append(f'incomplete evidence: {key}')
    claims = 0
    for page in sorted(pages):
        path = ROOT / page
        if not path.is_file():
            errors.append(f'missing page: {page}')
            continue
        content = path.read_text()
        for line, text in prose_lines(content):
            claims += 1
            if (page, text) not in verified:
                errors.append(f'{page}:{line}: missing evidence for {text}')
        if page not in {'docs/widgets/index.md', 'docs/layout/index.md'}:
            if not re.search(r'!\[[^]]+\]\(', content):
                errors.append(f'{page}: missing image')
            if '--8<--' not in content:
                errors.append(f'{page}: missing compiled snippet')
        for reference in re.findall(r'--8<-- "([^"]+)"', content):
            file, separator, region = reference.partition(':')
            snippet = ROOT / file
            if not snippet.is_file():
                errors.append(f'{page}: missing snippet {file}')
                continue
            if separator:
                code = snippet.read_text()
                if f'[start:{region}]' not in code or f'[end:{region}]' not in code:
                    errors.append(f'{page}: missing snippet region {region}')
        for target in re.findall(r'!?\[[^]]*\]\(([^)]+)\)', content):
            if '://' not in target and not target.startswith('#'):
                linked = path.parent / target.split('#')[0]
                if not linked.exists():
                    errors.append(f'{page}: missing link {target}')
    actual = {(page, text) for page in pages if (ROOT/page).exists()
              for _, text in prose_lines((ROOT/page).read_text())}
    for key in verified:
        if key[0] in pages and key not in actual:
            errors.append(f'stale evidence: {key}')
    if errors:
        print('\n'.join(errors), file=sys.stderr)
        return 1
    subprocess.run(['go', 'test', './docs/widget-examples/...', './docs/minimal-examples/widget-start', './docs/minimal-examples/widget-counter'], cwd=ROOT, check=True)
    subprocess.run(['go', 'run', './docs/widget-examples', '-check'], cwd=ROOT, check=True)
    print(f'Checked documentation links for {len(public)} public widgets, {len(pages)} evidence-backed pages, and {claims} prose entries.')
    print(f'{len(set(ADDITIONAL_WIDGET_PAGES.values()))} additional guides are inventoried without per-line evidence records.')
    print('Evidence coverage is mechanical; source review establishes whether each claim is true.')
    return 0


if __name__ == '__main__':
    sys.exit(main())
