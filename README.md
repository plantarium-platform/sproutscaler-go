
# SproutScaler: Intelligent Auto-Scaling Solution for HAProxy-Based Infrastructure

**SproutScaler** is a core component of the [Plantarium Platform](https://github.com/plantarium-platform), a lightweight and resource-efficient solution inspired by cloud architecture principles, designed for running serverless functions and microservices.

SproutScaler is an intelligent auto-scaling tool that dynamically manages the number of instances in an HAProxy-based infrastructure. It is designed to dynamically adjust the number of instances of monitored services based on observed metrics, responding to both increases and decreases in demand.

## Project Structure

```
.
├── go.mod               # Go module definition for dependencies
├── go.sum               # Checksums for Go modules
├── LICENSE              # License for SproutScaler
├── main.go              # Main application to initialize components and start polling
├── README.md            # Project documentation
├── resources
│   └── java-service-example-0.1-all.jar # Example Java service used in the POC
├── sproutscaler         # Core package directory for SproutScaler
│   ├── scaler           # Core logic for polling, scaling, and algorithms
│   │   ├── poller.go          # Polling logic to gather stats from HAProxy
│   │   ├── scaler.go          # Core scaling operations (adding/removing instances)
│   │   └── scaling_algorithm.go # Logic for calculating adjustments based on EMA and other metrics
│   └── util             # Utility functions and configuration management
│       ├── config.go         # Loads environment-based configuration
│       └── stats_storage.go  # Stores and manages historical stats for EMA calculations
└── tester.go            # Script to simulate load on the Java service instances
```

## Proof of Concept (POC)

The SproutScaler POC runs a Java service with a simple `/hello` endpoint that responds with "hello" and has a response time of 1+ second. The objective of the POC is to validate the auto-scaling capabilities of SproutScaler, specifically:

1. **Running Java Instances in HAProxy Backend**:
   - Initializes and manages up to 5 Java instances.
   - Responds dynamically to increase or decrease in demand based on response times by changing the amount of servers in HAProxy Backend (so far between 1 and 5)

2. **Dynamic Scaling Based on Response Times**:
   - **Adding Instances**: Scales up instances when response times degrade by more than 5%.
   - **Removing Instances**: Scales down instances when response times improve by over 20% and remain stable, or when no requests have been detected over a defined period.

3. **Integration with Dataplane API**:
   - Uses HAProxy’s Dataplane API for seamless management of backend instances as servers in HAProxy Backend
   - Monitors backend metrics and adjusts service count accordingly.

## Running Instructions

1. **Prerequisites**:
   - Ensure Go is installed and configured.
   - Set up the required HAProxy instance by following instructions from the [Plantarium Platform Graftnode Go repository](https://github.com/plantarium-platform/graftnode-go).
   - Ensure HAProxy is configured with Dataplane API enabled.

2. **Environment Configuration**:
   - Define environment variables as per your requirements, or rely on defaults:
      - `MAX_ENTRIES` (default: 10) - Maximum historical entries for EMA calculations.
      - `POLLING_INTERVAL_SECONDS` (default: 1) - Interval for polling HAProxy in seconds.
      - `EMA_DEPTH` (default: 10) - Depth for EMA calculation.
      - `BASE_SENSITIVITY_UP` and `BASE_SENSITIVITY_DOWN` - Sensitivity thresholds for scaling decisions.

3. **Starting SproutScaler**:
   ```bash
   go run main.go
   ```

4. **Testing Load with Tester Script**:
   - Use `tester.go` to simulate load by sending requests to the `/hello` endpoint.
   
## Scaling Mechanism

The **SproutScaler** auto-scaling mechanism adjusts instance counts based on observed response times, using an Exponential Moving Average (EMA) for smoothness and stability. This approach enables SproutScaler to react to trends rather than short-term fluctuations, helping maintain optimal performance and efficiency.

### Monitoring and EMA Calculation

SproutScaler continuously monitors the response time (`Rtime`) of each instance and calculates an EMA over a configured historical window of `N` steps. EMA smooths out sudden fluctuations in response time, providing a more stable trend line to base scaling decisions on. Specifically:

- The **Previous EMA** is defined as the EMA recorded `N` steps ago.
- The **Current EMA** is the most recent EMA calculated for the current observation.

### Instance Adjustment Calculation

The scaling decision is driven by comparing the **Current EMA** to the **Previous EMA** over the historical window:

1. **Delta Percent Calculation**: This measures the percentage change in response time between the Current and Previous EMA. The formula for **Delta Percent** is:

   \[
   \text{Delta Percent} = \frac{\text{Current EMA} - \text{Previous EMA}}{\text{Previous EMA}}
   \]

   - A **positive Delta Percent** signals an increase in response time, potentially indicating a need to add instances.
   - A **negative Delta Percent** indicates a decrease in response time, which can trigger instance removal if conditions are met.

2. **Instance Adjustment Formula**: To determine how many instances to add or remove, SproutScaler uses an exponential sensitivity adjustment based on `BaseSensitivity` parameters and the current instance count:

   - **Adding Instances**: When Delta Percent is positive, the adjustment formula applies a **BaseSensitivityUp** parameter to calculate the number of instances to add:

     \[
     \text{Instance Adjustment (Up)} = \left( \text{Delta Percent} \times \text{BaseSensitivityUp} \times e^{\frac{6}{\text{Instance Count} + 1}} \right)
     \]

   - **Removing Instances**: When Delta Percent is negative, the formula uses a **BaseSensitivityDown** parameter to calculate the number of instances to remove:

     \[
     \text{Instance Adjustment (Down)} = \left( \text{Delta Percent} \times \text{BaseSensitivityDown} \times e^{\frac{4.83}{\text{Instance Count} + 1}} \right)
     \]

3. **Total Instance Adjustment**: The final instance adjustment value is calculated as:

   \[
   \text{instanceAdjustment} = \text{int}(\text{Delta Percent} \times \text{Adjusted Sensitivity})
   \]

### Parameter Explanation

- **Base Sensitivity (Up and Down)**: These parameters define the aggressiveness of scaling in either direction. `BaseSensitivityUp` is tuned to ensure new instances are added when the response time increases by a certain percentage (e.g., 5%). Similarly, `BaseSensitivityDown` is set to trigger instance removal only after sustained improvement or no demand.

- **Exponential Formula**: The exponential adjustment, \( e^{\frac{6}{\text{Instance Count} + 1}} \) or \( e^{\frac{4.83}{\text{Instance Count} + 1}} \), helps make scaling decisions progressively more conservative as the instance count rises. This approach prevents excessive scaling in larger deployments, encouraging stability by gradually reducing sensitivity as more instances are added.

### Cooldown Mechanism

The cooldown mechanism delays scaling adjustments after any recent changes in instance count, whether due to adding or removing instances. This approach prevents immediate oscillation in scaling and provides a stabilization period to observe the effect of recent adjustments on the system’s response time.

### Zero Response Condition

If response times remain zero for `N` steps, indicating a lack of demand, SproutScaler will reduce instances to a minimum of one, preserving resources without fully deactivating service availability.

## Future Enhancements

This POC is an initial phase in integrating SproutScaler into the broader Plantarium ecosystem. Future development will focus on:
- **Ecosystem Integration**: As components in Plantarium are finalized, SproutScaler will relay signals to add or remove service instances across the ecosystem.
- **Idle Instance Replacement**: If services remain idle for extended periods, SproutScaler will replace the active instance with a graftnode instance, ready to respond to renewed demand.
- **Multi-Backend Processing**: Each backend service (leaf) in HAProxy will be processed independently, with listeners specific to their metrics and response requirements.

## Contact

If you have questions or want to contribute, feel free to reach out on [GitHub](https://github.com/glorko) or [Telegram](https://t.me/glorfindeil).
