# FIT File CLI (processFitFile)

Small CLI to process a `.fit` activity file and print a compact JSON summary.

## Purpose

This binary was extracted from an AWS Lambda function. It decodes a FIT file, extracts activity records, runs the existing analysis logic, and prints a JSON object with only the fields you care about.

The CLI prints the following JSON fields:
- `HeartAnalysis` (map / zone buckets)
- `ElevationGain` (float)
- `StoppedTime` (seconds)
- `ElapsedTime` (seconds)
- `NormalizedPower` (float)
- `PowerAnalysis` (map / zone buckets)

## Prerequisites

- Go 1.20+ (module-aware). Ensure `go` is on your PATH.
- Run commands from the `src` directory (this repository's CLI package lives here).

## Build

To build a local binary:

```sh
cd /path/to/.../processFitFile/src
go build -o processFitFile
```

## Run (development)

You can run without building:

```sh
# run against the included fixture
go run . --fit fixtures/Morning_Ride-8.fit

# or with a built binary
./processFitFile --fit fixtures/Morning_Ride-8.fit
```

## Example output

The program prints a single JSON object to stdout. Example (trimmed):

```json
{
  "HeartAnalysis": { "1": 255, "10": 255, ... },
  "ElevationGain": 0,
  "StoppedTime": 15899,
  "ElapsedTime": 35276,
  "NormalizedPower": 183.89182,
  "PowerAnalysis": { "1": 738, "10": 607, ... }
}
```

## How it works (brief)

- `main.go` is the CLI entry. It reads the `.fit` file, calls `ProcessActivityRecords`, and prints the requested JSON fields.
- `process.go` contains the activity processing logic (power, cadence, simplification, etc.).
- `fitHelper` decodes the FIT file and returns a `fit.ActivityFile`.
- The program reuses existing `dynamo` types for the shape of analysis results.

## Notes and next steps

- Units: Elevation/Distance conversions follow the original code. If you want different units, we can change conversions in `process.go`.
- Output file: I can add a `--out <file>` flag to write the JSON to disk.
- Tests: Add small unit tests for `ProcessActivityRecords` to guard behavior on small fixture files.

If you want any changes to the JSON shape or extra CLI flags, tell me which fields or flags and I will add them.
