# Signaloid Automated Execution Pipeline

This directory contains the GitHub Actions workflow (`signaloid_pipeline.yml`) that fully automates the building, execution, data fetching, plotting, and archiving of our Signaloid C programs.

## Overview

The pipeline is designed to run the Monte Carlo simulation alongside the Signaloid Uncertainty API (`UxHw`) implementation, compare their execution statistics, and automatically generate visualization plots.

### Workflow Steps
1. **Checkout & Setup:** Checks out the repository and sets up the Go environment (`v1.21`).
2. **Execute File 1 (Monte Carlo):** 
   - Builds `uniform-distribution-without-uxhw.c` on the Signaloid Cloud.
   - Submits exponentially increasing iteration tasks (from $10^0$ up to $10^7$).
   - Fetches and stores the outputs and execution statistics.
3. **Execute File 2 (Signaloid UxHw):**
   - Builds `uniform-distribution-with-uxhw.c`.
   - Submits parallel tasks (ignoring iteration arguments, as UxHw does this in $O(1)$ time).
   - Fetches and stores the outputs and execution statistics.
4. **Data Processing:**
   - Runs `plot_results.go` to generate interactive `<canvas>` HTML charts locally.
   - Runs `archive_json.go` to cleanly move raw JSON execution logs into a timestamped `history/` directory.
5. **Commit & Push:**
   - Automatically commits the generated `plots/`, `history/`, and tracking JSON files back to the repository using a GitHub Actions bot.

## How to Trigger the Pipeline

The workflow is configured with two triggers:
- **Push to `main`:** The pipeline automatically runs whenever new code is merged or pushed to the `main` branch.
- **Manual Dispatch:** You can manually trigger the pipeline from the **Actions** tab in your GitHub repository.

## Configuration & Variables

The workflow behavior is controlled by environment variables defined at the top of the `signaloid_pipeline.yml` file:

- `CORE_ID`: The Signaloid core architecture to target (e.g., `cor_b21e4de9927158c1a5b603c2affb8a09`).
- `CFILENAME_1` & `CFILENAME_2`: Paths to the C source files being evaluated.
- `UXHW_1` & `UXHW_2`: Boolean flags indicating whether the respective C file leverages the Uncertainty API.
- `MIN_VAL` & `MAX_VAL`: The range for the exponential iteration testing (e.g., `1` to `10000000`).

## Prerequisites (GitHub Secrets)

For the pipeline to authenticate with the Signaloid Cloud API, you **must** configure a Repository Secret:

1. Go to your repository on GitHub.
2. Navigate to **Settings** > **Secrets and variables** > **Actions**.
3. Click **New repository secret**.
4. Name the secret `API_KEY`.
5. Paste your Signaloid API Key as the value.

Once the secret is set, the pipeline will inject it securely into the Go scripts during runtime.