#!/bin/sh
set -eu
repo=$(git -C "$(dirname "$0")/../.." rev-parse --show-toplevel)
baseline=2871302c7d54ad7633a86c5925f00049e1ccc921
scratch=$(mktemp -d "${TMPDIR:-/tmp}/terma-regressions.XXXXXX")
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
git -C "$repo" archive "$baseline" | tar -x -C "$scratch"
for test in autocomplete_editing_regression_test.go list_filter_regression_test.go conditional_visibility_test.go; do
    cp "$repo/$test" "$scratch/$test"
done
pattern='TestAutocomplete_(UnicodeSuggestionRemainsEditable|TriggerAfterGrapheme|TextAreaEnterReplacesSelection)$|TestListFilterRefreshesWhenInputsChange$|TestDefaultListCursorProjectsOntoFilteredView$|TestVisibleWhenPreservesChildOutput$|TestVisibleWhenHiddenPreservesChildSize$'
if (cd "$scratch" && go test -count=1 -run "$pattern" .) > "$scratch/before.txt" 2>&1; then
    cat "$scratch/before.txt"
    printf 'Expected the original implementation to fail.\n' >&2
    exit 1
fi
for test in TestAutocomplete_UnicodeSuggestionRemainsEditable TestAutocomplete_TextAreaEnterReplacesSelection TestListFilterRefreshesWhenInputsChange TestDefaultListCursorProjectsOntoFilteredView TestVisibleWhenPreservesChildOutput; do
    grep -F -- "--- FAIL: $test" "$scratch/before.txt" > /dev/null
done
printf 'All five original defects reproduced at %s.\n' "$baseline"
(cd "$repo" && go test -count=1 -run "$pattern" .)
printf 'All five regressions pass on the current checkout.\n'
