# Run the documentation locally

From the repository root, use [uv](https://docs.astral.sh/uv/) to install the dependencies and start MkDocs:

```sh
uv run --project docs mkdocs serve
```

Open http://127.0.0.1:8000. MkDocs reloads when you edit a page. Stop it with Ctrl+C.
The project requires Python 3.14 or later; uv can download a compatible Python version.
Run the command from the repository root so snippet paths resolve correctly.

To build the site and treat warnings as errors:

```sh
uv run --project docs mkdocs build --strict
```

The output is written to `site/`.

To run the widget examples or check their source and images:

```sh
go run ./docs/widget-examples -list
go run ./docs/widget-examples -widget button
python3 scripts/verify-widget-docs.py
```
