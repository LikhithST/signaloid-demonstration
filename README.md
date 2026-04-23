# About Me
Software Engineer with 4 years of experience in the IT field at various companies, including Robert Bosch GmbH, specializing in cloud-native systems and system reliability. Experienced in developing microservices (Go, JavaScript, Python) and managing the full application lifecycle using Docker and Kubernetes across AWS and Azure, implementing CI/CD pipelines and enhancing system traceability through observability tools, such as OpenTelemetry and Prometheus.


# Signaloid API Demonstration Scripts

This repository contains shell scripts demonstrating how to interact with the [Signaloid Cloud API](https://signaloid.io/) to build and execute C programs. The scripts showcase two different approaches to calculating a portfolio's future value given an uncertain daily return: one using Signaloid's Uncertainty API (`uxhw.h`), and another using a traditional Monte Carlo simulation approach.

## Files Included

### 1. `run_signaloid_pipe_with_uxhw.sh`
This script submits a C program that leverages Signaloid's Uncertainty API (`uxhw.h`).
- Instead of looping through thousands of possibilities, it defines the daily return as an uncertain uniform distribution (`UxHwDoubleUniformDist(0.05, 0.07)`).
- Signaloid's hardware/microarchitecture propagates this uncertainty automatically through the calculation.
- The output is the entire probability distribution of the final portfolio value, calculated in a single pass without loops.

**C Code Snippet:**
```c
#include <stdio.h>
#include <uxhw.h>

int main() {
    double principal = 100000.0;

    // We define the market return as a known distribution of possibilities.
    // The hardware will propagate this uncertainty through the formula.
    double daily_return = UxHwDoubleUniformDist(0.05, 0.07);

    // One single calculation, zero loops.
    double final_value = principal * (1 + daily_return);

    // The output is the entire probability distribution of the result.
    printf("Portfolio outcome distribution: %lf\n", final_value);

    return 0;
}
```

### 2. `run_signaloid_pipe_without_uxhw.sh`
This script submits a traditional C program that relies on a standard Monte Carlo simulation.
- It uses the standard library (`rand()`) and loops 10,000 times to simulate different possible daily returns between 5% and 7%.
- The results are averaged to provide a projected average portfolio value.
- This serves as a baseline to compare against Signaloid's automated uncertainty-tracking capabilities.

**C Code Snippet:**
```c
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main() {
    double min = 0.05;
    double max = 0.07;
    int iterations = 10000;
    double principal = 100000.0;
    double sum_results = 0;

    srand(time(NULL));

    for (int i = 0; i < iterations; i++) {
        double daily_return = min + ((double)rand() / (double)RAND_MAX) * (max - min);
        double final_value = principal * (1 + daily_return);
        sum_results += final_value;
    }

    printf("Projected Average Portfolio Value: %.2f\n", sum_results / iterations);
    return 0;
}
```

## Prerequisites

To run these scripts, you need the following installed on your system:
- `curl` (for making HTTP requests to the API)
- `python3` (used as a lightweight JSON parser in the scripts)
- A valid Signaloid API Key.

## Usage

1. Make sure the scripts are executable:
   ```bash
   chmod +x run_signaloid_pipe_with_uxhw.sh
   chmod +x run_signaloid_pipe_without_uxhw.sh
   ```

2. Set your Signaloid API Key as an environment variable. The scripts read the `$API_KEY` variable from your environment:
   ```bash
   export API_KEY="your_actual_api_key_here"
   ```

3. Run the scripts:
   ```bash
   ./run_signaloid_pipe_with_uxhw.sh
   ./run_signaloid_pipe_without_uxhw.sh
   ```

## Script Workflow

Both scripts follow the same automated pipeline via the Signaloid API:
1. **Submit Build**: Uploads the embedded C code payload to be compiled.
2. **Poll Build Status**: Waits until the build is `Completed`.
3. **Submit Task**: Executes the compiled binary on a specified Signaloid Core (`CoreID`).
4. **Poll Task Status**: Waits until the execution is `Completed`.
5. **Retrieve Output**: Fetches and displays the standard output (stdout) from the executed task.