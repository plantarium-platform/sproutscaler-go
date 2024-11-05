# SproutScaler: Intelligent Auto-Scaling Solution

SproutScaler is an intelligent auto-scaling component that dynamically manages the number of instances in an HAProxy-based infrastructure. It is designed to aggressively add and remove instances to optimize energy consumption while maintaining service performance.

## Current Structure and POC

The current structure of the SproutScaler project consists of the following key components:

### HAProxyClient

This component provides a high-level interface for managing the HAProxy configuration, including creating backends, adding and removing servers, and handling transactions.

### SproutScaler

This is the core component of the project, responsible for managing the instances in the HAProxy backend. It uses the `HAProxyClient` to perform the necessary operations.

### Main Application

The main application, located in the `graftnode-go/main.go` file, creates the `HAProxyClient` and `SproutScaler` instances, and utilizes the `SproutScaler` component to manage the instances.

The current POC (Proof of Concept) implementation of the SproutScaler focuses on the following functionality:

1. **Instance Management**
    - The SproutScaler can add and remove instances from the HAProxy backend, up to a maximum of 5 instances.
    - It uses a LIFO (Last-In-First-Out) approach to manage the instances, ensuring that the most recently added instances are the first to be removed.

2. **Transactional Operations**
    - All operations performed by the SproutScaler, such as adding and removing instances, are wrapped in transactions to ensure consistency and reliability.

3. **Hardcoded Instance Configuration**
    - The current POC assumes that the instances are Java applications running on ports `8080` to `8084`.
    - The instance names are also hardcoded as `java-service-1` to `java-service-5`.

## Future Enhancements

The current POC is a basic implementation of the SproutScaler component. To make it a more intelligent and dynamic auto-scaling solution, the following enhancements can be made:

1. **Dynamic Instance Configuration**
    - Instead of hardcoding the instance names and ports, the SproutScaler should be able to handle instances with dynamic configurations.
    - This will allow the SproutScaler to manage instances of different types, not just Java applications.

2. **Monitoring and Scaling Triggers**
    - The SproutScaler should monitor the performance and utilization of the HAProxy backend, such as throughput, response times, and resource usage.
    - Based on these metrics, the SproutScaler should be able to dynamically add or remove instances to ensure optimal performance and energy efficiency.

3. **Scaling Algorithms and Policies**
    - The SproutScaler should employ more advanced scaling algorithms and policies, such as threshold-based scaling, predictive scaling, or reinforcement learning-based approaches.
    - These algorithms should be able to make informed decisions about when and how many instances to add or remove, based on the observed metrics and historical data.

4. **Integration with Monitoring and Alerting Systems**
    - The SproutScaler should be able to integrate with external monitoring and alerting systems, such as Prometheus, Grafana, or PagerDuty.
    - This will allow the SproutScaler to receive real-time updates on the infrastructure's performance and trigger scaling actions based on predefined thresholds or alerts.

5. **Adaptive and Self-Learning Capabilities**
    - The SproutScaler should have the ability to learn from past scaling actions and adjust its algorithms and policies accordingly.
    - This will enable the SproutScaler to become more intelligent and efficient over time, optimizing the infrastructure's performance and energy usage.

By implementing these enhancements, the SproutScaler can evolve into a truly intelligent and dynamic auto-scaling solution, capable of maintaining optimal service performance while minimizing energy consumption and costs.