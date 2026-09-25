# Troubleshooting

If `csvwatcher` is not found, verify that the npm global binary directory is on `PATH`, or
run it with `npx --package=@wyverncode/csvwatcher csvwatcher`.

If no JSON appears, verify the `.csv` extension, header row, matching row lengths, input
permissions, and separate input/output folders. Errors are printed and retried later.

If the GUI cannot start, check whether port 8080 is already in use. The GUI is intentionally
bound to `127.0.0.1` and is not an authenticated public service.
