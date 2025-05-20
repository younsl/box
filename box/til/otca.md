# OpenTelemetry Certified Associate (OTCA) Mock Exam

## OTCA Exam Domains & Competencies (Example)

| Domain                        | Percentage |
|-------------------------------|------------|
| Observability Fundamentals    | 20%        |
| OpenTelemetry Concepts        | 30%        |
| Instrumentation & Collection  | 20%        |
| Data Export & Analysis        | 16%        |
| Tooling & Ecosystem           | 14%        |
| **Total**                     | **100%**   |

- [Certification Summary Page](https://training.linuxfoundation.org/certification/opentelemetry-certified-associate-otca/)

---

**1. In OpenTelemetry, what does 'Instrumentation' mean?**

- A) Manually monitoring application performance
- B) Adding code or libraries to collect traces, metrics, and logs from an application
- C) Exporting data to external systems
- D) Visualizing data on a dashboard

<details>
<summary>View Answer & Explanation</summary>

**Answer: B)**

**Domain: OpenTelemetry Concepts (30%)**

**Explanation:** Instrumentation refers to the process of adding code or using libraries to enable the collection of traces, metrics, and logs from an application. This allows for automatic or manual collection of observability data.

</details>

---

**2. What is the primary role of the OpenTelemetry Collector?**

- A) Visualizing observability data
- B) Processing, transforming, and exporting observability data from various sources
- C) Directly instrumenting application code
- D) Compressing log files

<details>
<summary>View Answer & Explanation</summary>

**Answer: B)**

**Domain: Instrumentation & Collection (20%)**

**Explanation:** The OpenTelemetry Collector receives data from various sources (such as applications or agents), processes or transforms the data as needed, and exports it to external backends (like Jaeger, Prometheus, etc.).

</details>

---

**3. Which of the following is NOT a primary signal supported by OpenTelemetry?**

- A) Traces
- B) Metrics
- C) Logs
- D) Alerts

<details>
<summary>View Answer & Explanation</summary>

**Answer: D)**

**Domain: OpenTelemetry Concepts (30%)**

**Explanation:** OpenTelemetry natively supports three primary signals: traces (distributed tracing), metrics, and logs. Alerts are not a core signal type in OpenTelemetry.

</details>

---

**4. What is the main benefit of using semantic conventions in OpenTelemetry?**

- A) They increase the amount of data collected
- B) They ensure consistent naming and structure for observability data across different services
- C) They reduce the need for instrumentation
- D) They automatically export data to all supported backends

<details>
<summary>View Answer & Explanation</summary>

**Answer: B)**

**Domain: OpenTelemetry Concepts (30%)**

**Explanation:** Semantic conventions provide standardized naming and structure for attributes and resources, making it easier to analyze and correlate observability data across different services and environments.

</details>

---

**5. Which of the following best describes the 'exporter' component in OpenTelemetry?**

- A) It collects data from application code
- B) It processes and filters observability data
- C) It sends collected data to external backends or storage systems
- D) It visualizes traces and metrics

<details>
<summary>View Answer & Explanation</summary>

**Answer: C)**

**Domain: Data Export & Analysis (16%)**

**Explanation:** The exporter is responsible for sending collected observability data (traces, metrics, logs) to external systems such as Jaeger, Prometheus, or other observability platforms.

</details>

---
